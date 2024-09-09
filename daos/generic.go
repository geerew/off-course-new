package daos

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GenericDao is the data access object for generic queries
type GenericDao struct {
	db     database.Database
	caller daoer
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewGenericDao returns a new GenericDao
func NewGenericDao(db database.Database, caller daoer) *GenericDao {
	return &GenericDao{
		db:     db,
		caller: caller,
	}
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

// Delete deletes a row from a table based upon a where clause
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ProcessOrderBy takes an array of strings representing orderBy clauses and returns a processed
// version of this array
//
// It will creates a new list of valid Table columns based upon columns() for the current
// DAO
func (dao *GenericDao) ProcessOrderBy(orderBy []string, validColumns []string, explicit bool) []string {
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
