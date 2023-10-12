package api

// // import (
// // 	"time"

// // 	"github.com/geerew/off-course/database"
// // 	"github.com/geerew/off-course/models"
// // 	"github.com/gofiber/fiber/v2"
// // 	"github.com/rs/zerolog/log"
// // 	"gorm.io/gorm"
// // )

// // // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// // type attachments struct {
// // 	db database.Database
// // }

// // // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// // type attachmentResponse struct {
// // 	ID        string    `json:"id"`
// // 	AssetId   string    `json:"assetId"`
// // 	CourseID  string    `json:"courseId"`
// // 	Title     string    `json:"title"`
// // 	Path      string    `json:"path"`
// // 	CreatedAt time.Time `json:"createdAt"`
// // 	UpdatedAt time.Time `json:"updatedAt"`
// // }

// // // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// // func bindAttachmentsApi(router fiber.Router, db database.Database) {
// // 	api := attachments{db: db}

// // 	subGroup := router.Group("/attachments")

// // 	// Assets
// // 	subGroup.Get("", api.getAttachments)
// // 	subGroup.Get("/:id", api.getAttachment)
// // }

// // // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// // func (api *attachments) getAttachments(c *fiber.Ctx) error {
// // 	// Get the order by param
// // 	orderBy := c.Query("orderBy", "created_at desc")

// // 	attachments, err := models.GetAttachments(api.db, &database.DatabaseParams{OrderBy: orderBy})
// // 	if err != nil {
// // 		log.Err(err).Msg("error looking up attachments")
// // 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// // 			"message": "error looking up attachments - " + err.Error(),
// // 		})
// // 	}

// // 	return c.Status(fiber.StatusOK).JSON(toAttachmentResponse(attachments))
// // }

// // // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// // func (api *attachments) getAttachment(c *fiber.Ctx) error {
// // 	id := c.Params("id")

// // 	attachment, err := models.GetAttachment(api.db, id)

// // 	if err != nil {
// // 		if err == gorm.ErrRecordNotFound {
// // 			return c.Status(fiber.StatusNotFound).SendString("Not found")
// // 		}

// // 		log.Err(err).Msg("error looking up attachment")
// // 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// // 			"message": "error looking up attachment - " + err.Error(),
// // 		})
// // 	}

// // 	return c.Status(fiber.StatusOK).JSON(toAttachmentResponse([]*models.Attachment{attachment})[0])
// // }

// // // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// // // HELPER
// // // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// // func toAttachmentResponse(attachments []*models.Attachment) []*attachmentResponse {
// // 	responses := []*attachmentResponse{}
// // 	for _, attachment := range attachments {
// // 		responses = append(responses, &attachmentResponse{
// // 			ID:        attachment.ID,
// // 			AssetId:   attachment.AssetID,
// // 			CourseID:  attachment.CourseID,
// // 			Title:     attachment.Title,
// // 			Path:      attachment.Path,
// // 			CreatedAt: attachment.CreatedAt,
// // 			UpdatedAt: attachment.UpdatedAt,
// // 		})
// // 	}

// // 	return responses
// // }
