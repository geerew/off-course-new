package api

import (
	"database/sql"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type attachments struct {
	db database.Database
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type attachmentResponse struct {
	ID        string         `json:"id"`
	AssetId   string         `json:"assetId"`
	CourseID  string         `json:"courseId"`
	Title     string         `json:"title"`
	Path      string         `json:"path"`
	CreatedAt types.DateTime `json:"createdAt"`
	UpdatedAt types.DateTime `json:"updatedAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func bindAttachmentsApi(router fiber.Router, db database.Database) {
	api := attachments{db: db}

	subGroup := router.Group("/attachments")

	// Assets
	subGroup.Get("", api.getAttachments)
	subGroup.Get("/:id", api.getAttachment)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *attachments) getAttachments(c *fiber.Ctx) error {
	dbParams := &database.DatabaseParams{
		OrderBy:    []string{c.Query("orderBy", []string{"created_at desc"}...)},
		Pagination: pagination.New(c),
	}

	attachments, err := models.GetAttachments(c.UserContext(), api.db, dbParams)
	if err != nil {
		log.Err(err).Msg("error looking up attachments")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up attachments - " + err.Error(),
		})
	}

	pResult, err := dbParams.Pagination.BuildResult(toAttachmentResponse(attachments))
	if err != nil {
		log.Err(err).Msg("error building pagination result")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error building pagination result - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *attachments) getAttachment(c *fiber.Ctx) error {
	id := c.Params("id")

	attachment, err := models.GetAttachmentById(c.UserContext(), api.db, nil, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up attachment")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up attachment - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(toAttachmentResponse([]*models.Attachment{attachment})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPER
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func toAttachmentResponse(attachments []*models.Attachment) []*attachmentResponse {
	responses := []*attachmentResponse{}
	for _, attachment := range attachments {
		responses = append(responses, &attachmentResponse{
			ID:        attachment.ID,
			AssetId:   attachment.AssetID,
			CourseID:  attachment.CourseID,
			Title:     attachment.Title,
			Path:      attachment.Path,
			CreatedAt: attachment.CreatedAt,
			UpdatedAt: attachment.UpdatedAt,
		})
	}

	return responses
}
