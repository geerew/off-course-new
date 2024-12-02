package dao

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateParam creates a parameter
func (dao *DAO) CreateParam(ctx context.Context, param *models.Param) error {
	if param == nil {
		return utils.ErrNilPtr
	}

	return dao.Create(ctx, param)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetParam gets a parameter by its key
func (dao *DAO) GetParamByKey(ctx context.Context, param *models.Param) error {
	if param == nil {
		return utils.ErrNilPtr
	}

	if param.Key == "" {
		return utils.ErrInvalidKey
	}

	options := &database.Options{
		Where: squirrel.Eq{param.Table() + ".key": param.Key},
	}

	return dao.Get(ctx, param, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateParam updates a parameter
func (dao *DAO) UpdateParam(ctx context.Context, param *models.Param) error {
	if param == nil {
		return utils.ErrNilPtr
	}

	_, err := dao.Update(ctx, param)
	return err
}
