package daos

import (
	"database/sql"
	"errors"

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
func (dao *GenericDao) Count(baseSelect squirrel.SelectBuilder, dbParams *database.DatabaseParams) (int, error) {
	builder := baseSelect.
		Columns("COUNT(DISTINCT " + dao.table + ".id)")

	if dbParams != nil && dbParams.Where != nil {
		builder = builder.Where(dbParams.Where)
	}

	query, args, _ := builder.ToSql()

	var count int
	row := dao.db.QueryRow(query, args...).Scan(&count)

	return count, row
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get returns a row from a table
func (dao *GenericDao) Get(baseSelect squirrel.SelectBuilder, dbParams *database.DatabaseParams) (*sql.Row, error) {
	return dao.get(baseSelect, dbParams, dao.db.QueryRow)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get returns a row from a table in a transaction
func (dao *GenericDao) GetTx(baseSelect squirrel.SelectBuilder, dbParams *database.DatabaseParams, tx *sql.Tx) (*sql.Row, error) {
	return dao.get(baseSelect, dbParams, tx.QueryRow)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List returns rows from a table
func (dao *GenericDao) List(baseSelect squirrel.SelectBuilder, dbParams *database.DatabaseParams) (*sql.Rows, error) {
	builder := baseSelect

	if dbParams != nil {
		if dbParams.Columns != nil {
			builder = builder.Columns(dbParams.Columns...)
		}

		if dbParams.Where != "" {
			builder = builder.Where(dbParams.Where)
		}

		if dbParams.OrderBy != nil {
			builder = builder.OrderBy(dbParams.OrderBy...)
		}

		if dbParams.Pagination != nil {
			if count, err := dao.Count(baseSelect, dbParams); err != nil {
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

	return dao.db.Query(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes a row from a table
func (dao *GenericDao) Delete(id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	query, args, _ := squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Delete(dao.table).
		Where(squirrel.Eq{"id": id}).
		ToSql()

	_, err := dao.db.Exec(query, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (dao *GenericDao) get(baseSelect squirrel.SelectBuilder, dbParams *database.DatabaseParams, queryRowFn database.QueryRowFn) (*sql.Row, error) {
	if dbParams == nil || dbParams.Where == nil {
		return nil, ErrMissingWhere
	}

	builder := baseSelect

	if dbParams.Columns == nil {
		builder = builder.Columns(dao.table + "*")
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
