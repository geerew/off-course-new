package dao

import (
	"context"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateTag creates a tag
func (dao *DAO) CreateTag(ctx context.Context, tag *models.Tag) error {
	if tag == nil {
		return utils.ErrNilPtr
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// // Get gets a tag with the given ID or name
// //
// // CourseTags can be included by setting the `IncludeRelations` field in the dbParams. The coutseTags
// // can then be ordered by setting the `OrderBy` field in the dbParams, specifically referencing
// // courses_tags.[column]
// func (dao *TagDao) Get(id string, byName bool, dbParams *database.DatabaseParams, tx *database.Tx) (*models.Tag, error) {
// 	selectColumns, _ := tableColumnsOrPanic(models.Tag{}, dao.Table())

// 	tagDbParams := &database.DatabaseParams{
// 		Columns: selectColumns,
// 	}

// 	if byName {
// 		if dbParams != nil && dbParams.CaseInsensitive {
// 			tagDbParams.Where = squirrel.Eq{dao.Table() + ".tag COLLATE NOCASE": id}
// 		} else {
// 			tagDbParams.Where = squirrel.Eq{dao.Table() + ".tag": id}
// 		}
// 	} else {
// 		tagDbParams.Where = squirrel.Eq{dao.Table() + ".id": id}
// 	}

// 	tag, err := genericGet(dao, tagDbParams, dao.scanRow, tx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Get the course tags
// 	courseTagDao := NewCourseTagDao(dao.db)
// 	if dbParams != nil && slices.Contains(dbParams.IncludeRelations, courseTagDao.Table()) {
// 		_, orderByColumns := tableColumnsOrPanic(models.CourseTag{}, courseTagDao.Table())

// 		courseTagDbParams := &database.DatabaseParams{
// 			OrderBy: genericProcessOrderBy(dbParams.OrderBy, orderByColumns, courseTagDao, true),
// 			Where:   squirrel.Eq{"tag_id": id},
// 		}

// 		// Get the course_tags
// 		courseTags, err := courseTagDao.List(courseTagDbParams, tx)
// 		if err != nil {
// 			return nil, err
// 		}

// 		tag.CourseTags = courseTags
// 	}

// 	return tag, nil
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// // List lists tags
// //
// // CourseTags can be included by setting the `IncludeRelations` field in the dbParams. The coutseTags
// // can then be ordered by setting the `OrderBy` field in the dbParams, specifically referencing
// // courses_tags.[column]
// func (dao *TagDao) List(dbParams *database.DatabaseParams, tx *database.Tx) ([]*models.Tag, error) {
// 	if dbParams == nil {
// 		dbParams = &database.DatabaseParams{}
// 	}

// 	selectColumns, orderByColumns := tableColumnsOrPanic(models.Tag{}, dao.Table())

// 	// Backup the original order by then remove invalid orderBy columns
// 	origOrderBy := dbParams.OrderBy
// 	dbParams.OrderBy = genericProcessOrderBy(dbParams.OrderBy, orderByColumns, dao, false)

// 	// Default the columns if not specified
// 	if len(dbParams.Columns) == 0 {
// 		dbParams.Columns = selectColumns
// 	}

// 	tags, err := genericList(dao, dbParams, dao.scanRow, tx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Get the course_tags
// 	courseTagDao := NewCourseTagDao(dao.db)
// 	if len(tags) > 0 && slices.Contains(dbParams.IncludeRelations, courseTagDao.Table()) {
// 		// Get the tag IDs
// 		tagIds := []string{}
// 		for _, t := range tags {
// 			tagIds = append(tagIds, t.ID)
// 		}

// 		_, orderByColumns := tableColumnsOrPanic(models.CourseTag{}, courseTagDao.Table())

// 		// Reduce the order by clause to only include columns specific to the course_tags table
// 		reducedOrderBy := genericProcessOrderBy(origOrderBy, orderByColumns, courseTagDao, true)

// 		dbParams = &database.DatabaseParams{
// 			OrderBy: reducedOrderBy,
// 			Where:   squirrel.Eq{"tag_id": tagIds},
// 		}

// 		// Get the course_tags
// 		courseTags, err := courseTagDao.List(dbParams, tx)
// 		if err != nil {
// 			return nil, err
// 		}

// 		// Map the course_tags to the tags
// 		tagMap := map[string][]*models.CourseTag{}
// 		for _, ct := range courseTags {
// 			tagMap[ct.TagId] = append(tagMap[ct.TagId], ct)
// 		}

// 		// Assign the course_tags to the tags
// 		for _, t := range tags {
// 			t.CourseTags = tagMap[t.ID]
// 		}
// 	}

// 	return tags, nil
// }
