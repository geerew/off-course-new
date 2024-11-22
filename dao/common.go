package dao

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/schema"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DAO is a data access object
type DAO struct {
	db database.Database
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewDAO creates a new DAO
func NewDAO(db database.Database) *DAO {
	return &DAO{db: db}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create is a generic function to create a model in the database
func (dao *DAO) Create(ctx context.Context, model models.Modeler) error {
	sch, err := schema.Parse(model)
	if err != nil {
		return err
	}

	if model.Id() == "" {
		model.RefreshId()
	}

	model.RefreshCreatedAt()
	model.RefreshUpdatedAt()

	q := database.QuerierFromContext(ctx, dao.db)
	_, err = sch.Insert(model, q)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count is a generic function to count the number of rows in a table as determined by the model
func (dao *DAO) Count(ctx context.Context, model any, options *database.Options) (int, error) {
	sch, err := schema.Parse(model)
	if err != nil {
		return 0, err
	}

	q := database.QuerierFromContext(ctx, dao.db)
	return sch.Count(options, q)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get is a generic function to get a model (row)
func (dao *DAO) Get(ctx context.Context, model any, options *database.Options) error {
	sch, err := schema.Parse(model)
	if err != nil {
		return err
	}

	q := database.QuerierFromContext(ctx, dao.db)
	return sch.Select(model, options, q)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetById is a generic function to get a model (row) based on the ID of the model
func (dao *DAO) GetById(ctx context.Context, model models.Modeler) error {
	if model == nil {
		return utils.ErrNilPtr
	}

	if model.Id() == "" {
		return utils.ErrInvalidId
	}

	options := &database.Options{
		Where: squirrel.Eq{model.Table() + ".id": model.Id()},
	}

	return dao.Get(ctx, model, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List is a generic function to list models (rows)
func (dao *DAO) List(ctx context.Context, model any, options *database.Options) error {
	sch, err := schema.Parse(model)
	if err != nil {
		return err
	}

	if options != nil && options.Pagination != nil {
		count, err := dao.Count(ctx, model, options)
		if err != nil {
			return err
		}

		options.Pagination.SetCount(count)
	}

	q := database.QuerierFromContext(ctx, dao.db)
	err = sch.Select(model, options, q)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update is a generic function to update a model in the database
func (dao *DAO) Update(ctx context.Context, model models.Modeler) (bool, error) {
	sch, err := schema.Parse(model)
	if err != nil {
		return false, err
	}

	if model.Id() == "" {
		return false, utils.ErrInvalidId
	}

	model.RefreshUpdatedAt()

	q := database.QuerierFromContext(ctx, dao.db)
	res, err := sch.Update(model, &database.Options{Where: squirrel.Eq{model.Table() + ".id": model.Id()}}, q)
	if err != nil {
		return false, err
	}

	rowCount, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowCount > 0, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete is a generic function to delete a model (row)
//
// If options is nil or options.Where is nil, the function will delete the model based on the ID
// of the model
func (dao *DAO) Delete(ctx context.Context, model models.Modeler, options *database.Options) error {
	sch, err := schema.Parse(model)
	if err != nil {
		return err
	}

	if options == nil || options.Where == nil {
		options = &database.Options{Where: squirrel.Eq{model.Table() + ".id": model.Id()}}
	}

	q := database.QuerierFromContext(ctx, dao.db)
	_, err = sch.Delete(options, q)
	return err
}
