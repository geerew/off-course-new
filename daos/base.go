package daos

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Errors
var (
	ErrEmptyId         = errors.New("id cannot be empty")
	ErrMissingCourseId = errors.New("course id cannot be empty")
	ErrMissingWhere    = errors.New("where clause cannot be empty")
	ErrInvalidPrefix   = errors.New("prefix must be greater than 0")
	ErrNilTransaction  = errors.New("transaction cannot be nil")
	ErrMissingTag      = errors.New("tag cannot be empty")
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scannable is an interface for a database row that can be scanned into a struct
type Scannable interface {
	Scan(dest ...interface{}) error
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ScanFn is a type representing a generic function that scans a Scannable into
// a struct
type ScanFn[T any] func(Scannable) (*T, error)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// BaseDao
type BaseDao struct {
	db database.Database
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Db returns the database
func (dao *BaseDao) Db() database.Database {
	return dao.db
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count satisfies the daoer interface for Daos that do not require a count method
func (dao *BaseDao) Count(dbParams *database.DatabaseParams, tx *database.Tx) (int, error) {
	return 0, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type daoer interface {
	Db() database.Database
	Table() string
	Count(*database.DatabaseParams, *database.Tx) (int, error)
	countSelect() squirrel.SelectBuilder
	baseSelect() squirrel.SelectBuilder
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// FormatTime formats a time.Time to a string
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.000")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ParseTime parses the time string from SQLite to time.Time
func ParseTime(t string) (time.Time, error) {
	if t == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02 15:04:05.000", t)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ParseTimeNull parses a time string from a sql.NullString to time.Time
func ParseTimeNull(t sql.NullString) (time.Time, error) {
	if t.Valid {
		if value, err := ParseTime(t.String); err != nil {
			return time.Time{}, err
		} else {
			return value, nil
		}
	} else {
		return time.Time{}, nil
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NilStr returns nil when a string is empty
//
// Use this when inserting into the database to avoid inserting empty strings
func NilStr(s string) any {
	if s == "" {
		return nil
	}

	return s
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// extractTableColumn extracts the table and column name from an orderBy string. If no table prefix
// is found, the table part is returned as an empty string
func extractTableColumn(orderBy string) (string, string) {
	parts := strings.Fields(orderBy)
	tableColumn := strings.Split(parts[0], ".")

	if len(tableColumn) == 2 {
		return tableColumn[0], tableColumn[1]
	}

	return "", tableColumn[0]
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// isValidOrderBy returns true if the orderBy string is valid. The table and column are validated
// against the given list of valid table.columns (ex. courses.id, scans.status as scan_status).
func isValidOrderBy(table, column string, validateTableColumns []string) bool {
	// If the column is empty, always return false
	if column == "" {
		return false
	}

	for _, validTc := range validateTableColumns {
		// Wildcard match (ex. courses.* == id)
		if table == "" && strings.HasSuffix(validTc, ".*") {
			return true
		}

		// Exact match (ex. id == id || courses.id == courses.id || courses.id as courses_id == courses.id)
		if validTc == column || validTc == table+"."+column || strings.HasPrefix(validTc, table+"."+column+" as ") {
			return true
		}

		// Table + wildcard match (ex. courses.* == courses.id)
		if strings.HasSuffix(validTc, ".*") && strings.HasPrefix(validTc, table+".") {
			return true
		}

		// courses.id as course_id == course_id
		if strings.HasSuffix(validTc, " as "+column) {
			return true
		}
	}

	return false
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GenericProcessOrderBy processes the orderBy strings to ensure they are valid
func GenericProcessOrderBy(orderBy []string, validColumns []string, explicit bool) []string {
	if len(orderBy) == 0 {
		return orderBy
	}

	var processedOrderBy []string

	for _, ob := range orderBy {
		Table, column := extractTableColumn(ob)

		if explicit && Table == "" {
			continue
		}

		if isValidOrderBy(Table, column, validColumns) {
			processedOrderBy = append(processedOrderBy, ob)
		}
	}

	return processedOrderBy
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GenericCount counts the number of rows in a table
func GenericCount(dao daoer, dbParams *database.DatabaseParams, tx *database.Tx) (int, error) {
	queryRowFn := dao.Db().QueryRow
	if tx != nil {
		queryRowFn = tx.QueryRow
	}

	builder := dao.countSelect().
		Columns("COUNT(DISTINCT " + dao.Table() + ".id)")

	if dbParams != nil && dbParams.Where != nil {
		builder = builder.Where(dbParams.Where)
	}

	query, args, _ := builder.ToSql()

	var count int
	err := queryRowFn(query, args...).Scan(&count)

	return count, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GenericGet gets a row from the given table
func GenericGet[T any](
	dao daoer,
	dbParams *database.DatabaseParams,
	scanFn ScanFn[T],
	tx *database.Tx,
) (*T, error) {
	queryRowFn := dao.Db().QueryRow
	if tx != nil {
		queryRowFn = tx.QueryRow
	}

	if dbParams == nil || dbParams.Where == nil {
		return nil, ErrMissingWhere
	}

	builder := dao.baseSelect()

	if dbParams.Columns == nil {
		builder = builder.Columns(dao.Table() + ".*")
	} else {
		builder = builder.Columns(dbParams.Columns...)
	}

	if dbParams.OrderBy != nil {
		builder = builder.OrderBy(dbParams.OrderBy...)
	}

	query, args, _ := builder.
		Where(dbParams.Where).
		ToSql()

	row := queryRowFn(query, args...)
	if row.Err() != nil {
		return nil, row.Err()
	}

	res, err := scanFn(row)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GenericList lists rows from the given table
func GenericList[T any](
	dao daoer,
	dbParams *database.DatabaseParams,
	scanFn ScanFn[T],
	tx *database.Tx,
) ([]*T, error) {
	rows, err := GenericListWithoutScan(dao, dbParams, tx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*T
	for rows.Next() {
		r, err := scanFn(rows)
		if err != nil {
			return nil, err
		}

		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GenericListWithoutScan lists rows from the given table but leaves the scanning to the
// caller
func GenericListWithoutScan(
	dao daoer,
	dbParams *database.DatabaseParams,
	tx *database.Tx,
) (*sql.Rows, error) {
	queryFn := dao.Db().Query
	if tx != nil {
		queryFn = tx.Query
	}

	builder := dao.baseSelect()

	if dbParams != nil {
		if dbParams.Columns == nil {
			builder = builder.Columns(dao.Table() + ".*")
		} else {
			builder = builder.Columns(dbParams.Columns...)
		}

		if dbParams.Where != "" {
			builder = builder.Where(dbParams.Where)
		}

		if dbParams.OrderBy != nil {
			builder = builder.OrderBy(dbParams.OrderBy...)
		}

		if dbParams.Pagination != nil {
			if count, err := dao.Count(dbParams, tx); err != nil {
				return nil, err
			} else {
				dbParams.Pagination.SetCount(count)
				builder = builder.
					Offset(uint64(dbParams.Pagination.Offset())).
					Limit(uint64(dbParams.Pagination.Limit()))
			}
		}

		if dbParams.GroupBys != nil {
			builder = builder.GroupBy(dbParams.GroupBys...)
		}

		if dbParams.Having != nil {
			builder = builder.Having(dbParams.Having)
		}
	}

	query, args, _ := builder.ToSql()

	return queryFn(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GenericDelete deletes a row from a table based upon a where clause
func GenericDelete(dao daoer, dbParams *database.DatabaseParams, tx *database.Tx) error {
	execFn := dao.Db().Exec
	if tx != nil {
		execFn = tx.Exec
	}

	if dbParams == nil || dbParams.Where == nil {
		return ErrMissingWhere
	}

	query, args, _ := squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Delete(dao.Table()).
		Where(dbParams.Where).
		ToSql()

	_, err := execFn(query, args...)
	return err
}
