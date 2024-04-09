package api

import (
	"database/sql"

	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type attachments struct {
	appFs         *appFs.AppFs
	attachmentDao *daos.AttachmentDao
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

func bindAttachmentsApi(router fiber.Router, appFs *appFs.AppFs, db database.Database) {
	api := attachments{
		appFs:         appFs,
		attachmentDao: daos.NewAttachmentDao(db),
	}

	subGroup := router.Group("/attachments")

	// Assets
	subGroup.Get("", api.getAttachments)
	subGroup.Get("/:id", api.getAttachment)
	subGroup.Get("/:id/serve", api.serveAttachment)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *attachments) getAttachments(c *fiber.Ctx) error {
	dbParams := &database.DatabaseParams{
		OrderBy:    []string{c.Query("orderBy", []string{"created_at desc"}...)},
		Pagination: pagination.NewFromApi(c),
	}

	attachments, err := api.attachmentDao.List(dbParams, nil)
	if err != nil {
		log.Err(err).Msg("error looking up attachments")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up attachments - " + err.Error(),
		})
	}

	pResult, err := dbParams.Pagination.BuildResult(attachmentResponseHelper(attachments))
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

	attachment, err := api.attachmentDao.Get(id, nil)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up attachment")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up attachment - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(attachmentResponseHelper([]*models.Attachment{attachment})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *attachments) serveAttachment(c *fiber.Ctx) error {
	id := c.Params("id")

	attachment, err := api.attachmentDao.Get(id, nil)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up attachment")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up attachment - " + err.Error(),
		})
	}

	// Check for invalid path
	if exists, err := afero.Exists(api.appFs.Fs, attachment.Path); err != nil || !exists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "attachment does not exist",
		})
	}

	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+attachment.Title+`"`)
	return filesystem.SendFile(c, afero.NewHttpFs(api.appFs.Fs), attachment.Path)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPER
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func attachmentResponseHelper(attachments []*models.Attachment) []*attachmentResponse {
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
