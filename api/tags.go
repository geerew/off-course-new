package api

import (
	"database/sql"
	"log/slog"
	"net/url"
	"sort"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type tags struct {
	logger       *slog.Logger
	tagDao       *daos.TagDao
	courseTagDao *daos.CourseTagDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type tagResponse struct {
	ID          string         `json:"id"`
	Tag         string         `json:"tag"`
	CourseCount int            `json:"courseCount"`
	Courses     []*courseTag   `json:"courses,omitempty"`
	CreatedAt   types.DateTime `json:"createdAt"`
	UpdatedAt   types.DateTime `json:"updatedAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type courseTag struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tags) getTags(c *fiber.Ctx) error {
	expand := c.QueryBool("expand", false)
	filter := c.Query("filter", "")
	orderBy := c.Query("orderBy", api.tagDao.Table()+".tag asc")

	dbParams := &database.DatabaseParams{
		OrderBy:    strings.Split(orderBy, ","),
		Pagination: pagination.NewFromApi(c),
	}

	if expand {
		dbParams.IncludeRelations = []string{api.courseTagDao.Table()}
	}

	if filter != "" {
		dbParams.Where = squirrel.Like{api.tagDao.Table() + ".tag": "%" + filter + "%"}
	}

	tags, err := api.tagDao.List(dbParams, nil)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up tags", err)
	}

	if filter != "" {
		sort.SliceStable(tags, func(i, j int) bool {
			// Convert tags and filter to lower case for case insensitive comparison
			iTag, jTag := strings.ToLower(tags[i].Tag), strings.ToLower(tags[j].Tag)
			filterLower := strings.ToLower(filter)

			// Check for exact matches, starts with, and contains in a case insensitive manner
			iExact, jExact := iTag == filterLower, jTag == filterLower
			iStarts, jStarts := strings.HasPrefix(iTag, filterLower), strings.HasPrefix(jTag, filterLower)
			iContains, jContains := strings.Contains(iTag, filterLower), strings.Contains(jTag, filterLower)

			// Prioritize exact matches first
			if iExact && !jExact {
				return true
			} else if !iExact && jExact {
				return false
			}

			// Then prioritize tags starting with the filter
			if iStarts && !jStarts {
				return true
			} else if !iStarts && jStarts {
				return false
			}

			// Lastly, sort by those that contain the substring, alphabetically
			if iContains && jContains {
				return iTag < jTag // Use case insensitive comparison for alphabetical order
			}
			return iContains && !jContains
		})
	}

	pResult, err := dbParams.Pagination.BuildResult(tagResponseHelper(tags))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tags) getTag(c *fiber.Ctx) error {
	id := c.Params("id")
	expand := c.QueryBool("expand", false)
	byName := c.QueryBool("byName", false)
	insensitive := c.QueryBool("insensitive", false)

	dbParams := &database.DatabaseParams{
		CaseInsensitive: insensitive,
	}

	if expand {
		dbParams.IncludeRelations = []string{api.courseTagDao.Table()}
	}

	if byName {
		// Decode the URL-encoded parameter
		var err error
		id, err = url.QueryUnescape(id)

		if err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "Error decoding name parameter", err)
		}
	}

	tag, err := api.tagDao.Get(id, byName, dbParams, nil)

	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Tag not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up tag", err)
	}

	return c.Status(fiber.StatusOK).JSON(tagResponseHelper([]*models.Tag{tag})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tags) createTag(c *fiber.Ctx) error {
	tag := new(models.Tag)

	if err := c.BodyParser(tag); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	// Ensure there is a title and path
	if tag.Tag == "" {
		return errorResponse(c, fiber.StatusBadRequest, "A tag is required", nil)
	}

	// Empty stuff that should not be set
	tag.ID = ""

	if err := api.tagDao.Create(tag, nil); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return errorResponse(c, fiber.StatusBadRequest, "Tag already exists", err)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error creating tag", err)
	}

	return c.Status(fiber.StatusCreated).JSON(tagResponseHelper([]*models.Tag{tag})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tags) updateTag(c *fiber.Ctx) error {
	id := c.Params("id")

	// Parse the request body to get the updated fields
	reqTag := &tagResponse{}
	if err := c.BodyParser(reqTag); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	// Create an asset progress
	tag := &models.Tag{
		BaseModel: models.BaseModel{ID: id},
		Tag:       reqTag.Tag,
	}

	// Update the asset progress
	if err := api.tagDao.Update(tag, nil); err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid tag", err)
		}

		if strings.HasPrefix(err.Error(), "constraint failed") {
			return errorResponse(c, fiber.StatusBadRequest, "Duplicate tag", err)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error updating tag", err)
	}

	// Get the updated asset
	updatedTag, err := api.tagDao.Get(id, false, nil, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Tag not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up tag", err)
	}

	return c.Status(fiber.StatusOK).JSON(tagResponseHelper([]*models.Tag{updatedTag})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tags) deleteTag(c *fiber.Ctx) error {
	id := c.Params("id")

	err := api.tagDao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": id}}, nil)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting tag", err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func tagResponseHelper(tags []*models.Tag) []*tagResponse {
	responses := []*tagResponse{}

	for _, tag := range tags {
		t := &tagResponse{
			ID:          tag.ID,
			Tag:         tag.Tag,
			CreatedAt:   tag.CreatedAt,
			UpdatedAt:   tag.UpdatedAt,
			CourseCount: tag.CourseCount,
		}

		// Add the course tags
		if len(tag.CourseTags) > 0 {
			courses := []*courseTag{}

			for _, ct := range tag.CourseTags {
				courses = append(courses, &courseTag{
					ID:    ct.ID,
					Title: ct.Course,
				})
			}

			t.Courses = courses
		}

		responses = append(responses, t)
	}

	return responses
}
