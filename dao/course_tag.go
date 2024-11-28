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

// CreateCourseTag creates a course tag
func (dao *DAO) CreateCourseTag(ctx context.Context, courseTag *models.CourseTag) error {
	if courseTag == nil {
		return utils.ErrNilPtr
	}

	if courseTag.TagID == "" && courseTag.Tag == "" {
		return fmt.Errorf("tag ID and tag cannot be empty")
	}

	return dao.db.RunInTransaction(ctx, func(txCtx context.Context) error {
		if courseTag.TagID != "" {
			return dao.Create(txCtx, courseTag)
		}

		// Get the tag by tag name
		tag := models.Tag{}
		err := dao.Get(txCtx, &tag, &database.Options{Where: squirrel.Eq{models.TAG_TABLE + ".tag": courseTag.Tag}})
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		// If the tag does not exist, create it
		if err == sql.ErrNoRows {
			tag.Tag = courseTag.Tag
			err = dao.Create(txCtx, &tag)
			if err != nil {
				return err
			}
		}

		courseTag.TagID = tag.ID

		return dao.Create(txCtx, courseTag)

	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PluckCourseIDsWithTags returns a list of course IDs where the course has all the tags in the
// slice
func (dao *DAO) PluckCourseIDsWithTags(ctx context.Context, tags []string, options *database.Options) ([]string, error) {
	if len(tags) == 0 {
		return nil, nil
	}

	if options == nil {
		options = &database.Options{}
	}

	options.Where = squirrel.Eq{models.TAG_TABLE + ".tag": tags}
	options.GroupBy = []string{models.COURSE_TAG_TABLE + ".course_id"}
	options.Having = squirrel.Expr("COUNT(DISTINCT "+models.TAG_TABLE+".tag) = ?", len(tags))

	return dao.ListPluck(ctx, &models.CourseTag{}, options, models.COURSE_TAG_COURSE_ID)
}
