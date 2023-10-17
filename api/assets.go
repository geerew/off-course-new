package api

import (
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type assets struct {
	db database.Database
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type assetResponse struct {
	ID        string         `json:"id"`
	CourseID  string         `json:"courseId"`
	Title     string         `json:"title"`
	Prefix    int            `json:"prefix"`
	Chapter   string         `json:"chapter"`
	Path      string         `json:"path"`
	Type      types.Asset    `json:"assetType"`
	Started   bool           `json:"started"`
	Finished  bool           `json:"finished"`
	CreatedAt types.DateTime `json:"createdAt"`
	UpdatedAt types.DateTime `json:"updatedAt"`

	// Association
	Attachments []*attachmentResponse `json:"attachments,omitempty"`
}

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func bindAssetsApi(router fiber.Router, db database.Database) {
// 	api := assets{db: db}

// 	subGroup := router.Group("/assets")

// 	// Assets
// 	subGroup.Get("", api.getAssets)
// 	subGroup.Get("/:id", api.getAsset)
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func (api *assets) getAssets(c *fiber.Ctx) error {
// 	// Get the order by param
// 	orderBy := c.Query("orderBy", "created_at desc")

// 	dbParams := &database.DatabaseParams{
// 		OrderBy:    orderBy,
// 		Preload:    true,
// 		Pagination: pagination.New(c),
// 	}

// 	assets, err := models.GetAssets(api.db, dbParams)
// 	if err != nil {
// 		log.Err(err).Msg("error looking up assets")
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "error looking up assets - " + err.Error(),
// 		})
// 	}

// 	pResult, err := dbParams.Pagination.BuildResult(toAssetResponse(assets))
// 	if err != nil {
// 		log.Err(err).Msg("error building pagination result")
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "error building pagination result - " + err.Error(),
// 		})
// 	}

// 	return c.Status(fiber.StatusOK).JSON(pResult)
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func (api *assets) getAsset(c *fiber.Ctx) error {
// 	id := c.Params("id")

// 	asset, err := models.GetAsset(api.db, id, &database.DatabaseParams{Preload: true})

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return c.Status(fiber.StatusNotFound).SendString("Not found")
// 		}

// 		log.Err(err).Msg("error looking up asset")
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "error looking up asset - " + err.Error(),
// 		})
// 	}

// 	return c.Status(fiber.StatusOK).JSON(toAssetResponse([]*models.Asset{asset})[0])
// }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPER
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func toAssetResponse(assets []*models.Asset) []*assetResponse {
	responses := []*assetResponse{}
	for _, asset := range assets {
		responses = append(responses, &assetResponse{
			ID:        asset.ID,
			CourseID:  asset.CourseID,
			Title:     asset.Title,
			Prefix:    asset.Prefix,
			Chapter:   asset.Chapter,
			Path:      asset.Path,
			Type:      asset.Type,
			Started:   asset.Started,
			Finished:  asset.Finished,
			CreatedAt: asset.CreatedAt,
			UpdatedAt: asset.UpdatedAt,

			// Association
			Attachments: toAttachmentResponse(asset.Attachments),
		})

	}

	return responses
}
