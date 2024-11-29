package dao

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateTag creates a tag
func (dao *DAO) CreateTag(ctx context.Context, tag *models.Tag) error {
	if tag == nil {
		return utils.ErrNilPtr
	}

	// Check if the tag already exists
	options := &database.Options{
		Where: squirrel.Expr(
			fmt.Sprintf("LOWER(%s.%s) = LOWER(?)", models.TAG_TABLE, models.TAG_TAG),
			tag.Tag,
		),
	}
	existingTag := &models.Tag{}
	err := dao.Get(ctx, existingTag, options)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// The tag already exists, update the tag with the existing tag and attempt to create it. This
	// will result in an error but it gives a more specific error message
	if err == nil {
		tag.Tag = existingTag.Tag
	}

	return dao.Create(ctx, tag)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateTag updates a tag
func (dao *DAO) UpdateTag(ctx context.Context, tag *models.Tag) error {
	if tag == nil {
		return utils.ErrNilPtr
	}

	_, err := dao.Update(ctx, tag)
	return err
}
