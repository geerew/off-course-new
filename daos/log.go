package daos

import (
	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// LogDao is the data access object for logs
type LogDao struct {
	BaseDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewLogDao returns a new LogDao
func NewLogDao(db database.Database) *LogDao {
	return &LogDao{
		BaseDao: BaseDao{
			db:    db,
			table: "logs",
		},
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count counts the logs
func (dao *LogDao) Count(dbParams *database.DatabaseParams, tx *database.Tx) (int, error) {
	return genericCount(dao, dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Write writes a new log
func (dao *LogDao) Write(l *models.Log, tx *database.Tx) error {
	if l.ID == "" {
		l.RefreshId()
	}

	l.RefreshCreatedAt()
	l.UpdatedAt = l.CreatedAt

	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.Table()).
		SetMap(toDBMapOrPanic(l)).
		ToSql()

	execFn := dao.db.Exec
	if tx != nil {
		execFn = tx.Exec
	}

	_, err := execFn(query, args...)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List lists logs
func (dao *LogDao) List(dbParams *database.DatabaseParams, tx *database.Tx) ([]*models.Log, error) {
	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	// Always override the order by to created_at
	dbParams.OrderBy = []string{dao.Table() + ".created_at DESC"}

	// Default the columns if not specified
	if len(dbParams.Columns) == 0 {
		selectColumns, _ := tableColumnsOrPanic(models.Log{}, dao.Table())
		dbParams.Columns = selectColumns
	}

	return genericList(dao, dbParams, dao.scanRow, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes logs based upon the where clause
func (dao *LogDao) Delete(dbParams *database.DatabaseParams, tx *database.Tx) error {
	return genericDelete(dao, dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans a log row
func (dao *LogDao) scanRow(scannable Scannable) (*models.Log, error) {
	var l models.Log

	err := scannable.Scan(
		&l.ID,
		&l.CreatedAt,
		&l.UpdatedAt,
		&l.Level,
		&l.Message,
		&l.Data,
	)

	if err != nil {
		return nil, err
	}

	return &l, nil
}
