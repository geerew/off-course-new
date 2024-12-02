package dao

import (
	"context"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAttachment creates an attachment
func (dao *DAO) CreateAttachment(ctx context.Context, attachment *models.Attachment) error {
	if attachment == nil {
		return utils.ErrNilPtr
	}

	return dao.Create(ctx, attachment)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateAttachment updates an attachment
func (dao *DAO) UpdateAttachment(ctx context.Context, attachment *models.Attachment) error {
	if attachment == nil {
		return utils.ErrNilPtr
	}

	_, err := dao.Update(ctx, attachment)
	return err
}
