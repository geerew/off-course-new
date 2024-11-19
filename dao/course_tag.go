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

// // ListCourseIdsByTags lists course IDs containing all tags in the slice
// func (dao *DAO) ListCourseIdsByTags(tags []string, dbParams *database.DatabaseParams, tx *database.Tx) ([]string, error) {
// 	if len(tags) == 0 {
// 		return nil, nil
// 	}

// 	if dbParams == nil {
// 		dbParams = &database.DatabaseParams{}
// 	}

// 	selectColumns, _ := tableColumnsOrPanic(models.CourseTag{}, dao.Table())

// 	dbParams.OrderBy = genericProcessOrderBy(dbParams.OrderBy, selectColumns, dao, false)
// 	dbParams.Columns = []string{dao.Table() + ".course_id"}
// 	dbParams.Where = squirrel.Eq{NewTagDao(dao.db).Table() + ".tag": tags}
// 	dbParams.GroupBys = []string{dao.Table() + ".course_id"}
// 	dbParams.Having = squirrel.Expr("COUNT(DISTINCT "+NewTagDao(dao.db).Table()+".tag) = ?", len(tags))
// 	dbParams.Pagination = nil

// 	rows, err := genericListWithoutScan(dao, dbParams, tx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var courseIds []string
// 	for rows.Next() {
// 		var courseId string
// 		if err := rows.Scan(&courseId); err != nil {
// 			return nil, err
// 		}

// 		courseIds = append(courseIds, courseId)
// 	}

// 	if err := rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return courseIds, nil
// }
