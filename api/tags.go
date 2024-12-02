package api

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"sort"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type tagsAPI struct {
	logger *slog.Logger
	dao    *dao.DAO
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type courseTag struct {
	ID       string `json:"id"`
	CourseID string `json:"courseId"`
	Title    string `json:"title"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initTagRoutes initializes the tag routes
func (r *Router) initTagRoutes() {
	tagsAPI := tagsAPI{
		logger: r.config.Logger,
		dao:    r.dao,
	}

	tagGroup := r.api.Group("/tags")
	tagGroup.Get("", tagsAPI.getTags)
	tagGroup.Get("/:name", tagsAPI.getTag)
	tagGroup.Post("", tagsAPI.createTag)
	tagGroup.Put("/:id", tagsAPI.updateTag)
	tagGroup.Delete("/:id", tagsAPI.deleteTag)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tagsAPI) getTags(c *fiber.Ctx) error {
	filter := c.Query("filter", "")
	orderBy := c.Query("orderBy", models.TAG_TABLE+".tag asc")

	options := &database.Options{
		OrderBy:    strings.Split(orderBy, ","),
		Pagination: pagination.NewFromApi(c),
	}

	if filter != "" {
		options.Where = squirrel.Like{fmt.Sprintf("%s.%s", models.TAG_TABLE, models.TAG_TAG): "%" + filter + "%"}
	}

	tags := []*models.Tag{}
	err := api.dao.List(c.Context(), &tags, options)
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
				return iTag < jTag
			}
			return iContains && !jContains
		})
	}

	pResult, err := options.Pagination.BuildResult(tagResponseHelper(tags))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tagsAPI) getTag(c *fiber.Ctx) error {
	name := c.Params("name")

	var err error
	name, err = url.QueryUnescape(name)

	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error decoding name parameter", err)
	}

	options := &database.Options{
		Where: squirrel.Eq{fmt.Sprintf("%s.%s", models.TAG_TABLE, models.TAG_TAG): name},
	}

	tag := &models.Tag{}
	err = api.dao.Get(c.Context(), tag, options)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Tag not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up tag", err)
	}

	return c.Status(fiber.StatusOK).JSON(tagResponseHelper([]*models.Tag{tag})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tagsAPI) createTag(c *fiber.Ctx) error {
	tag := new(models.Tag)
	if err := c.BodyParser(tag); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	if tag.Tag == "" {
		return errorResponse(c, fiber.StatusBadRequest, "A tag is required", nil)
	}

	tag.ID = ""
	err := api.dao.CreateTag(c.Context(), tag)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return errorResponse(c, fiber.StatusBadRequest, "Tag already exists", err)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error creating tag", err)
	}

	return c.Status(fiber.StatusCreated).JSON(tagResponseHelper([]*models.Tag{tag})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tagsAPI) updateTag(c *fiber.Ctx) error {
	id := c.Params("id")

	reqTag := &tagRequest{}
	if err := c.BodyParser(reqTag); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	tag := &models.Tag{Base: models.Base{ID: id}}
	err := api.dao.GetById(c.Context(), tag)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Tag not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up tag", err)
	}

	tag.Tag = reqTag.Tag

	err = api.dao.UpdateTag(c.Context(), tag)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid tag", err)
		}

		if strings.HasPrefix(err.Error(), "constraint failed") {
			return errorResponse(c, fiber.StatusBadRequest, "Duplicate tag", err)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error updating tag", err)
	}

	return c.Status(fiber.StatusOK).JSON(tagResponseHelper([]*models.Tag{tag})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tagsAPI) deleteTag(c *fiber.Ctx) error {
	id := c.Params("id")

	tag := &models.Tag{Base: models.Base{ID: id}}
	err := api.dao.Delete(c.Context(), tag, nil)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting tag", err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
