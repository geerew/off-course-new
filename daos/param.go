package daos

import (
	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ParamDao is the data access object for params
type ParamDao struct {
	BaseDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewParamDao returns a new ParamDao
func NewParamDao(db database.Database) *ParamDao {
	return &ParamDao{
		BaseDao: BaseDao{
			db:    db,
			table: "params",
		},
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get gets a parameter by its key
func (dao *ParamDao) Get(key string, tx *database.Tx) (*models.Param, error) {
	selectColumns, _ := tableColumnsOrPanic(models.Param{}, dao.Table())

	dbParams := &database.DatabaseParams{
		Columns: selectColumns,
		Where:   squirrel.Eq{dao.Table() + ".key": key},
	}

	return genericGet(dao, dbParams, dao.scanRow, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update updates a parameter value by its key
func (dao *ParamDao) Update(p *models.Param, tx *database.Tx) error {
	if p.ID == "" {
		return ErrEmptyId
	}

	p.RefreshUpdatedAt()

	// Convert to a map so we have the rendered values
	data := toDBMapOrPanic(p)

	query, args, _ := squirrel.
		StatementBuilder.
		Update(dao.Table()).
		Set("value", data["value"]).
		Set("updated_at", data["updated_at"]).
		Where("id = ?", data["id"]).
		ToSql()

	execFn := dao.db.Exec
	if tx != nil {
		execFn = tx.Exec
	}

	_, err := execFn(query, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans a parameter row
func (dao *ParamDao) scanRow(scannable Scannable) (*models.Param, error) {
	var p models.Param

	err := scannable.Scan(
		&p.ID,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.Key,
		&p.Value,
	)

	if err != nil {
		return nil, err
	}

	return &p, nil
}
