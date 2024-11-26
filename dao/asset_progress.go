package dao

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateOrUpdateAssetProgress creates/updates an asset progress and refreshes course progress
func (dao *DAO) CreateOrUpdateAssetProgress(ctx context.Context, assetProgress *models.AssetProgress) error {
	if assetProgress == nil {
		return utils.ErrNilPtr
	}

	return dao.db.RunInTransaction(ctx, func(txCtx context.Context) error {
		if assetProgress.VideoPos < 0 {
			assetProgress.VideoPos = 0
		}

		// Check for existing asset progress
		existingAP := &models.AssetProgress{}
		err := dao.Get(
			txCtx,
			existingAP,
			&database.Options{Where: squirrel.Eq{models.ASSET_PROGRESS_TABLE + ".asset_id": assetProgress.AssetID}},
		)

		if err != nil && err != sql.ErrNoRows {
			return err
		}

		// Create
		if err == sql.ErrNoRows {
			if assetProgress.Completed {
				assetProgress.CompletedAt = types.NowDateTime()
			}

			err := dao.Create(txCtx, assetProgress)
			if err != nil {
				return err
			}
		} else {
			assetProgress.ID = existingAP.ID
			// Update
			if assetProgress.Completed {
				if existingAP.Completed {
					assetProgress.CompletedAt = existingAP.CompletedAt
				} else {
					assetProgress.CompletedAt = types.NowDateTime()
				}
			} else {
				assetProgress.CompletedAt = types.DateTime{}
			}

			_, err = dao.Update(txCtx, assetProgress)
			if err != nil {
				return err
			}
		}

		// Refresh course progress
		return dao.RefreshCourseProgress(txCtx, assetProgress.CourseID)
	})
}
