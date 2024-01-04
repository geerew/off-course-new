package api

import (
	"database/sql"
	"fmt"
	"mime"
	"path/filepath"
	"strconv"
	"strings"

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

type assets struct {
	appFs *appFs.AppFs
	db    database.Database
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type assetResponse struct {
	ID          string         `json:"id"`
	CourseID    string         `json:"courseId"`
	Title       string         `json:"title"`
	Prefix      int            `json:"prefix"`
	Chapter     string         `json:"chapter"`
	Path        string         `json:"path"`
	Type        types.Asset    `json:"assetType"`
	Progress    int            `json:"progress"`
	Completed   bool           `json:"completed"`
	CompletedAt types.DateTime `json:"completedAt"`
	CreatedAt   types.DateTime `json:"createdAt"`
	UpdatedAt   types.DateTime `json:"updatedAt"`

	// Association
	Attachments []*attachmentResponse `json:"attachments,omitempty"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const bufferSize = 1024 * 8                 // 8KB per chunk, adjust as needed
const maxInitialChunkSize = 1024 * 1024 * 5 // 5MB, adjust as needed

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func bindAssetsApi(router fiber.Router, appFs *appFs.AppFs, db database.Database) {
	api := assets{appFs: appFs, db: db}

	subGroup := router.Group("/assets")

	// Assets
	subGroup.Get("", api.getAssets)
	subGroup.Get("/:id", api.getAsset)
	subGroup.Put("/:id", api.updateAsset)
	subGroup.Get("/:id/serve", api.serveAsset)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *assets) getAssets(c *fiber.Ctx) error {
	dbParams := &database.DatabaseParams{
		OrderBy:    []string{c.Query("orderBy", []string{"created_at desc"}...)},
		Pagination: pagination.New(c),
	}

	if c.QueryBool("expand", false) {
		dbParams.Relation = []database.Relation{
			{Struct: "Attachments", OrderBy: []string{"title asc"}},
		}
	}

	assets, err := models.GetAssets(c.UserContext(), api.db, dbParams)
	if err != nil {
		log.Err(err).Msg("error looking up assets")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up assets - " + err.Error(),
		})
	}

	pResult, err := dbParams.Pagination.BuildResult(toAssetResponse(assets))
	if err != nil {
		log.Err(err).Msg("error building pagination result")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error building pagination result - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *assets) getAsset(c *fiber.Ctx) error {
	id := c.Params("id")

	// Include relations
	dbParams := &database.DatabaseParams{}
	if c.QueryBool("expand", false) {
		dbParams.Relation = []database.Relation{
			{Struct: "Attachments", OrderBy: []string{"title asc"}},
		}
	}

	asset, err := models.GetAssetById(c.UserContext(), api.db, dbParams, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up asset")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up asset - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(toAssetResponse([]*models.Asset{asset})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *assets) updateAsset(c *fiber.Ctx) error {
	id := c.Params("id")

	// Parse the request body to get the updated fields
	reqAsset := &assetResponse{}
	if err := c.BodyParser(reqAsset); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Failed to parse request body",
		})
	}

	var updatedAsset *models.Asset
	var err error

	// Update progress
	if _, err = models.UpdateAssetProgress(c.UserContext(), api.db, id, reqAsset.Progress); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error updating asset")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error updating asset - " + err.Error(),
		})
	}

	// Update completed
	if updatedAsset, err = models.UpdateAssetCompleted(c.UserContext(), api.db, id, reqAsset.Completed); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error updating asset")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error updating asset - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(toAssetResponse([]*models.Asset{updatedAsset})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *assets) serveAsset(c *fiber.Ctx) error {
	id := c.Params("id")

	asset, err := models.GetAssetById(c.UserContext(), api.db, nil, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up asset")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up asset - " + err.Error(),
		})
	}

	// Check for invalid path
	if exists, err := afero.Exists(api.appFs.Fs, asset.Path); err != nil || !exists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "asset does not exist",
		})
	}

	if asset.Type.IsVideo() {
		return handleVideo(c, api.appFs, asset)
	}

	// Handle pdf
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "done",
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPER
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func toAssetResponse(assets []*models.Asset) []*assetResponse {
	responses := []*assetResponse{}
	for _, asset := range assets {
		responses = append(responses, &assetResponse{
			ID:          asset.ID,
			CourseID:    asset.CourseID,
			Title:       asset.Title,
			Prefix:      asset.Prefix,
			Chapter:     asset.Chapter,
			Path:        asset.Path,
			Type:        asset.Type,
			Progress:    asset.Progress,
			Completed:   asset.Completed,
			CompletedAt: asset.CompletedAt,
			CreatedAt:   asset.CreatedAt,
			UpdatedAt:   asset.UpdatedAt,

			// Association
			Attachments: toAttachmentResponse(asset.Attachments),
		})

	}

	return responses
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// handleVideo handles the video streaming logic
func handleVideo(c *fiber.Ctx, appFs *appFs.AppFs, asset *models.Asset) error {
	// Open the video
	file, err := appFs.Fs.Open(asset.Path)
	if err != nil {
		log.Err(err).Str("path", asset.Path).Msg("error opening file")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "internal error - " + err.Error(),
		})
	}
	defer file.Close()

	// Get the file info
	fileInfo, err := file.Stat()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "internal error - " + err.Error(),
		})
	}

	// Get the range header and return the entire video if there is no range header
	rangeHeader := c.Get("Range", "")
	if rangeHeader == "" {
		return filesystem.SendFile(c, afero.NewHttpFs(appFs.Fs), asset.Path)
	}

	// Parse the "bytes=START-END" format
	bytesPos := strings.Split(rangeHeader, "=")[1]
	rangeStartEnd := strings.Split(bytesPos, "-")
	start, _ := strconv.Atoi(rangeStartEnd[0])
	var end int
	if rangeStartEnd[1] == "" {
		// Calculate the initial chunk end based on maxInitialChunkSize
		end = start + maxInitialChunkSize - 1
		if end >= int(fileInfo.Size()) {
			end = int(fileInfo.Size()) - 1
		}
	} else {
		end, _ = strconv.Atoi(rangeStartEnd[1])
	}

	if start > end {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "range start cannot be greater than end",
		})
	}

	// Setting required response headers
	c.Set(fiber.HeaderContentRange, fmt.Sprintf("bytes %d-%d/%d", start, end, fileInfo.Size()))
	c.Set(fiber.HeaderContentLength, strconv.Itoa(end-start+1))
	c.Set(fiber.HeaderContentType, mime.TypeByExtension(filepath.Ext(asset.Path)))
	c.Set(fiber.HeaderAcceptRanges, "bytes")

	// Set the status code to 206 Partial Content
	c.Status(fiber.StatusPartialContent)

	file.Seek(int64(start), 0)
	buffer := make([]byte, bufferSize)
	bytesToSend := end - start + 1

	// Respond in chunks
	for bytesToSend > 0 {
		bytesRead, err := file.Read(buffer)
		if err != nil {
			break
		}

		if bytesRead > bytesToSend {
			bytesRead = bytesToSend
		}

		c.Write(buffer[:bytesRead])
		bytesToSend -= bytesRead
	}

	return nil
}
