package daos

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// LogDao is the data access object for logs
type LogDao struct {
	db    database.Database
	table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewLogDao returns a new LogDao
func NewLogDao(db database.Database) *LogDao {
	return &LogDao{
		db:    db,
		table: "logs",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table returns the table name
func (dao *LogDao) Table() string {
	return dao.table
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count returns the number of logs
func (dao *LogDao) Count(params *database.DatabaseParams, tx *sql.Tx) (int, error) {
	generic := NewGenericDao(dao.db, dao)
	return generic.Count(params, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Write inserts a new log
func (dao *LogDao) Write(l *models.Log, tx *sql.Tx) error {
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

	var err error
	if tx != nil {
		_, err = tx.Exec(query, args...)
	} else {
		_, err = dao.db.Exec(query, args...)
	}

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List selects logs
//
// `tx` allows for the function to be run within a transaction
func (dao *LogDao) List(dbParams *database.DatabaseParams, tx *sql.Tx) ([]*models.Log, error) {
	generic := NewGenericDao(dao.db, dao)

	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	// Always override the order by to created_at
	dbParams.OrderBy = []string{dao.Table() + ".created_at DESC"}

	// Default the columns if not specified
	if len(dbParams.Columns) == 0 {
		dbParams.Columns = dao.columns()
	}

	rows, err := generic.List(dbParams, tx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.Log

	for rows.Next() {
		log, err := dao.scanRow(rows)
		if err != nil {
			return nil, err
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes logs based upon the where clause
//
// `tx` allows for the function to be run within a transaction
func (dao *LogDao) Delete(dbParams *database.DatabaseParams, tx *sql.Tx) error {
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
		"created_at": a.CreatedAt,
		"updated_at": a.UpdatedAt,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans a log row
func (dao *LogDao) scanRow(scannable Scannable) (*models.Log, error) {
	var l models.Log

	err := scannable.Scan(
		&l.ID,
		&l.Level,
		&l.Message,
		&l.Data,
		&l.CreatedAt,
		&l.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &l, nil
}
