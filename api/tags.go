package api

import (
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type tags struct {
	tagDao       *daos.TagDao
	courseTagDao *daos.CourseTagDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type tagResponse struct {
	ID        string         `json:"id"`
	Tag       string         `json:"tag"`
	Courses   []*courseTag   `json:"courses,omitempty"`
	CreatedAt types.DateTime `json:"createdAt"`
	UpdatedAt types.DateTime `json:"updatedAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type courseTag struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func bindTagsApi(router fiber.Router, db database.Database) {
	api := tags{
		tagDao:       daos.NewTagDao(db),
		courseTagDao: daos.NewCourseTagDao(db),
	}

	subGroup := router.Group("/tags")

	subGroup.Get("/", api.getTags)
	subGroup.Delete("/:id", api.deleteTag)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tags) getTags(c *fiber.Ctx) error {
	expand := c.QueryBool("expand", false)
	orderBy := c.Query("orderBy", "tag asc")

	dbParams := &database.DatabaseParams{
		OrderBy:    strings.Split(orderBy, ","),
		Pagination: pagination.NewFromApi(c),
	}

	if expand {
		dbParams.IncludeRelations = []string{api.courseTagDao.Table}
	}

	tags, err := api.tagDao.List(dbParams, nil)
	if err != nil {
		log.Err(err).Msg("error looking up tags")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up tags - " + err.Error(),
		})
	}

	pResult, err := dbParams.Pagination.BuildResult(tagResponseHelper(tags))
	if err != nil {
		log.Err(err).Msg("error building pagination result")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error building pagination result - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tags) deleteTag(c *fiber.Ctx) error {
	id := c.Params("id")

	err := api.tagDao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": id}}, nil)
	if err != nil {
		log.Err(err).Msg("error deleting tag")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error deleting tag - " + err.Error(),
		})

	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPER
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func tagResponseHelper(tags []*models.Tag) []*tagResponse {
	responses := []*tagResponse{}

	for _, tag := range tags {
		t := &tagResponse{
			ID:        tag.ID,
			Tag:       tag.Tag,
			CreatedAt: tag.CreatedAt,
			UpdatedAt: tag.UpdatedAt,
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
