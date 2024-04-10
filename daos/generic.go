package daos

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GenericDao is the data access object for generic queries
type GenericDao struct {
	db    database.Database
	table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewGenericDao returns a new GenericDao
func NewGenericDao(db database.Database, table string) *GenericDao {
	return &GenericDao{db: db, table: table}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count returns the number of rows in a table
//
// `tx` allows for the function to be run within a transaction
func (dao *GenericDao) Count(baseSelect squirrel.SelectBuilder, dbParams *database.DatabaseParams, tx *sql.Tx) (int, error) {
	queryRowFn := dao.db.QueryRow
	if tx != nil {
		queryRowFn = tx.QueryRow
	}

	builder := baseSelect.
		Columns("COUNT(DISTINCT " + dao.table + ".id)")

	// if dbParams == nil || dbParams.Columns == nil {
	// 	builder = builder.Columns(dao.table + ".*")
	// } else {
	// 	builder = builder.Columns(dbParams.Columns...)
	// }

	if dbParams != nil && dbParams.Where != nil {
		builder = builder.Where(dbParams.Where)
	}

	query, args, _ := builder.ToSql()

	var count int
	err := queryRowFn(query, args...).Scan(&count)

	return count, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get returns a row from a table
//
// `tx` allows for the function to be run within a transaction
func (dao *GenericDao) Get(baseSelect squirrel.SelectBuilder, dbParams *database.DatabaseParams, tx *sql.Tx) (*sql.Row, error) {
	queryRowFn := dao.db.QueryRow
	if tx != nil {
		queryRowFn = tx.QueryRow
	}

	if dbParams == nil || dbParams.Where == nil {
		return nil, ErrMissingWhere
	}

	builder := baseSelect

	if dbParams.Columns == nil {
		builder = builder.Columns(dao.table + ".*")
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

	return row, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List returns rows from a table
//
// `tx` allows for the function to be run within a transaction
func (dao *GenericDao) List(baseSelect squirrel.SelectBuilder, dbParams *database.DatabaseParams, tx *sql.Tx) (*sql.Rows, error) {
	queryFn := dao.db.Query
	if tx != nil {
		queryFn = tx.Query
	}

	builder := baseSelect

	if dbParams != nil {
		if dbParams.Columns == nil {
			builder = builder.Columns(dao.table + ".*")
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
			if count, err := dao.Count(baseSelect, dbParams, tx); err != nil {
				return nil, err
			} else {
				dbParams.Pagination.SetCount(count)
				builder = builder.
					Offset(uint64(dbParams.Pagination.Offset())).
					Limit(uint64(dbParams.Pagination.Limit()))
			}
		}
	}

	query, args, _ := builder.
		Where(dbParams.Where).
		ToSql()

	return queryFn(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes a row from a table based upon a where clause
//
// `tx` allows for the function to be run within a transaction
func (dao *GenericDao) Delete(dbParams *database.DatabaseParams, tx *sql.Tx) error {
	if dbParams == nil || dbParams.Where == nil {
		return ErrMissingWhere
	}

	execFn := dao.db.Exec
	if tx != nil {
		execFn = tx.Exec
	}

	query, args, _ := squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Delete(dao.table).
		Where(dbParams.Where).
		ToSql()

	_, err := execFn(query, args...)
	return err
}
