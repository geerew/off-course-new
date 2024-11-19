package dao

import (
	"context"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WriteLog writes a new log
func (dao *DAO) WriteLog(ctx context.Context, log *models.Log) error {
	if log == nil {
		return utils.ErrNilPtr
	}

	return dao.Create(ctx, log)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List lists logs
// func (dao *LogDao) List(dbParams *database.DatabaseParams, tx *database.Tx) ([]*models.Log, error) {
// 	if dbParams == nil {
// 		dbParams = &database.DatabaseParams{}
// 	}

// 	// Always override the order by to created_at
// 	dbParams.OrderBy = []string{dao.Table() + ".created_at DESC"}

// 	// Default the columns if not specified
// 	if len(dbParams.Columns) == 0 {
// 		selectColumns, _ := tableColumnsOrPanic(models.Log{}, dao.Table())
// 		dbParams.Columns = selectColumns
// 	}

// 	return genericList(dao, dbParams, dao.scanRow, tx)
// }
