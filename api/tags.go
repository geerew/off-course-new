package api

import (
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
	tagDao *daos.TagDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type tagResponse struct {
	ID        string         `json:"id"`
	Tag       string         `json:"tag"`
	CreatedAt types.DateTime `json:"createdAt"`
	UpdatedAt types.DateTime `json:"updatedAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func bindTagsApi(router fiber.Router, db database.Database) {
	api := tags{
		tagDao: daos.NewTagDao(db),
	}

	subGroup := router.Group("/tags")

	subGroup.Get("/", api.getTags)
	subGroup.Delete("/:id", api.deleteTag)
	// subGroup.Get("/:courseId", api.getScan)
	// subGroup.Post("", api.createScan)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tags) getTags(c *fiber.Ctx) error {
	dbParams := &database.DatabaseParams{
		OrderBy:    []string{c.Query("orderBy", []string{"tag asc"}...)},
		Pagination: pagination.NewFromApi(c),
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
		responses = append(responses, &tagResponse{
			ID:        tag.ID,
			Tag:       tag.Tag,
			CreatedAt: tag.CreatedAt,
			UpdatedAt: tag.UpdatedAt,
		})
	}

	return responses
}
