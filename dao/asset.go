package dao

import (
	"context"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAsset creates an asset and refreshes course progress
func (dao *DAO) CreateAsset(ctx context.Context, asset *models.Asset) error {
	if asset == nil {
		return utils.ErrNilPtr
	}

	return dao.db.RunInTransaction(ctx, func(txCtx context.Context) error {
		err := dao.Create(txCtx, asset)
		if err != nil {
			return err
		}

		return dao.RefreshCourseProgress(txCtx, asset.CourseID)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateAsset updates an asset
func (dao *DAO) UpdateAsset(ctx context.Context, asset *models.Asset) error {
	if asset == nil {
		return utils.ErrNilPtr
	}

	_, err := dao.Update(ctx, asset)
	return err
}
