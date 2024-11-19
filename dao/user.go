package dao

import (
	"context"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateUser creates a user
func (dao *DAO) CreateUser(ctx context.Context, user *models.User) error {
	if user == nil {
		return utils.ErrNilPtr
	}

	return dao.Create(ctx, user)
}
