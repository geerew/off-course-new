package dao

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateScan creates a scan
func (dao *DAO) CreateScan(ctx context.Context, scan *models.Scan) error {
	if scan == nil {
		return utils.ErrNilPtr
	}

	// A scan should always be in the waiting state when created
	if !scan.Status.IsWaiting() {
		scan.Status.SetWaiting()
	}

	return dao.Create(ctx, scan)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateScan updates a scan
func (dao *DAO) UpdateScan(ctx context.Context, scan *models.Scan) error {
	if scan == nil {
		return utils.ErrNilPtr
	}

	_, err := dao.Update(ctx, scan)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Next gets the next scan whose status is `waiting“ based upon the created_at column
func (dao *DAO) NextWaitingScan(ctx context.Context, model models.Modeler) error {
	if model == nil {
		return utils.ErrNilPtr
	}

	options := &database.Options{
		Where:   squirrel.Eq{model.Table() + ".status": types.ScanStatusWaiting},
		OrderBy: []string{model.Table() + ".created_at ASC"},
	}

	return dao.Get(ctx, model, options)
}
