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
	table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewLogDao returns a new LogDao
func NewLogDao(db database.Database) *LogDao {
	return &LogDao{
		BaseDao: BaseDao{db: db},
		table:   "logs",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table returns the table name
func (dao *LogDao) Table() string {
	return dao.table
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count counts the logs
func (dao *LogDao) Count(dbParams *database.DatabaseParams, tx *database.Tx) (int, error) {
	return GenericCount(dao, dbParams, tx)
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
		SetMap(dao.data(l)).
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
		dbParams.Columns = dao.columns()
	}

	return GenericList(dao, dbParams, dao.scanRow, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes logs based upon the where clause
func (dao *LogDao) Delete(dbParams *database.DatabaseParams, tx *database.Tx) error {
	if dbParams == nil || dbParams.Where == nil {
		return ErrMissingWhere
	}

	generic := NewGenericDao(dao.db, dao)
	return generic.Delete(dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// countSelect returns the default count select builder
func (dao *LogDao) countSelect() squirrel.SelectBuilder {
	return squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.Table()).
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
func (dao *LogDao) baseSelect() squirrel.SelectBuilder {
	return dao.countSelect()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// columns returns the columns to select
func (dao *LogDao) columns() []string {
	return []string{
		dao.Table() + ".*",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// data generates a map of key/values for a log
func (dao *LogDao) data(a *models.Log) map[string]any {
	return map[string]any{
		"id":         a.ID,
		"level":      a.Level,
		"message":    NilStr(a.Message),
		"data":       a.Data,
		"created_at": FormatTime(a.CreatedAt),
		"updated_at": FormatTime(a.UpdatedAt),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans a log row
func (dao *LogDao) scanRow(scannable Scannable) (*models.Log, error) {
	var l models.Log

	var createdAt string
	var updatedAt string

	err := scannable.Scan(
		&l.ID,
		&l.Level,
		&l.Message,
		&l.Data,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	if l.CreatedAt, err = ParseTime(createdAt); err != nil {
		return nil, err
	}

	if l.UpdatedAt, err = ParseTime(updatedAt); err != nil {
		return nil, err
	}

	return &l, nil
}
