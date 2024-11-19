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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update updates select columns of an asset progress, then refresh course progress
//
//   - `video_pos` (for video assets)
//   - `completed` (for all assets)
//
// When `completed` is true, `completed_at` is set to the current time, else it will be null
//
// A new transaction is created if `tx` is nil
// func (dao *AssetProgressDao) Update(ap *models.AssetProgress, tx *database.Tx) error {

//

// 		// Update (or create if it doesn't exist)
// 		query, args, _ := squirrel.
// 			StatementBuilder.
// 			Insert(dao.Table()).
// 			SetMap(modelToMapOrPanic(ap)).
// 			Suffix(
// 				"ON CONFLICT (asset_id) DO UPDATE SET video_pos = ?, completed = ?, completed_at = ?, updated_at = ?",
// 				ap.VideoPos, ap.Completed, ap.CompletedAt, ap.UpdatedAt,
// 			).
// 			ToSql()

// 		_, err = tx.Exec(query, args...)
// 		if err != nil {
// 			return err
// 		}

// 		// Refresh course progress
// 		cpDao := NewCourseProgressDao(dao.db)
// 		return cpDao.Refresh(ap.CourseID, tx)
// 	}
// }
