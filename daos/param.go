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
	dbParams := &database.DatabaseParams{
		Columns: dao.columns(),
		Where:   squirrel.Eq{dao.Table() + ".key": key},
	}

	return GenericGet(dao, dbParams, dao.scanRow, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update updates a parameter value by its key
func (dao *ParamDao) Update(p *models.Param, tx *database.Tx) error {
	if p.ID == "" {
		return ErrEmptyId
	}

	p.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Update(dao.Table()).
		Set("value", p.Value).
		Set("updated_at", FormatTime(p.UpdatedAt)).
		Where("id = ?", p.ID).
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

// countSelect returns the default select builder for counting
func (dao *ParamDao) countSelect() squirrel.SelectBuilder {
	return squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.Table()).
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
func (dao *ParamDao) baseSelect() squirrel.SelectBuilder {
	return dao.countSelect()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// columns returns the columns to select
func (dao *ParamDao) columns() []string {
	return []string{
		"id",
		"key",
		"value",
		"created_at",
		"updated_at",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans a parameter row
func (dao *ParamDao) scanRow(scannable Scannable) (*models.Param, error) {
	var p models.Param

	var createdAt string
	var updatedAt string

	err := scannable.Scan(
		&p.ID,
		&p.Key,
		&p.Value,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	if p.CreatedAt, err = ParseTime(createdAt); err != nil {
		return nil, err
	}

	if p.UpdatedAt, err = ParseTime(updatedAt); err != nil {
		return nil, err
	}

	return &p, nil
}
