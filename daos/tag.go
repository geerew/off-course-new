package daos

import (
	"slices"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TagDao is the data access object for tags
type TagDao struct {
	BaseDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTagDao returns a new TagDao
func NewTagDao(db database.Database) *TagDao {
	return &TagDao{
		BaseDao: BaseDao{
			db:    db,
			table: "tags",
		},
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count counts the tags
func (dao *TagDao) Count(dbParams *database.DatabaseParams, tx *database.Tx) (int, error) {
	return genericCount(dao, dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create creates a tag
func (dao *TagDao) Create(t *models.Tag, tx *database.Tx) error {
	execFn := dao.db.Exec
	if tx != nil {
		execFn = tx.Exec
	}

	if t.ID == "" {
		t.RefreshId()
	}

	t.RefreshCreatedAt()
	t.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.Table()).
		SetMap(modelToMapOrPanic(t)).
		ToSql()

	_, err := execFn(query, args...)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get gets a tag with the given ID or name
//
// CourseTags can be included by setting the `IncludeRelations` field in the dbParams. The coutseTags
// can then be ordered by setting the `OrderBy` field in the dbParams, specifically referencing
// courses_tags.[column]
func (dao *TagDao) Get(id string, byName bool, dbParams *database.DatabaseParams, tx *database.Tx) (*models.Tag, error) {
	selectColumns, _ := tableColumnsOrPanic(models.Tag{}, dao.Table())

	tagDbParams := &database.DatabaseParams{
		Columns: selectColumns,
	}

	if byName {
		if dbParams != nil && dbParams.CaseInsensitive {
			tagDbParams.Where = squirrel.Eq{dao.Table() + ".tag COLLATE NOCASE": id}
		} else {
			tagDbParams.Where = squirrel.Eq{dao.Table() + ".tag": id}
		}
	} else {
		tagDbParams.Where = squirrel.Eq{dao.Table() + ".id": id}
	}

	tag, err := genericGet(dao, tagDbParams, dao.scanRow, tx)
	if err != nil {
		return nil, err
	}

	// Get the course tags
	courseTagDao := NewCourseTagDao(dao.db)
	if dbParams != nil && slices.Contains(dbParams.IncludeRelations, courseTagDao.Table()) {
		_, orderByColumns := tableColumnsOrPanic(models.CourseTag{}, courseTagDao.Table())

		courseTagDbParams := &database.DatabaseParams{
			OrderBy: genericProcessOrderBy(dbParams.OrderBy, orderByColumns, courseTagDao, true),
			Where:   squirrel.Eq{"tag_id": id},
		}

		// Get the course_tags
		courseTags, err := courseTagDao.List(courseTagDbParams, tx)
		if err != nil {
			return nil, err
		}

		tag.CourseTags = courseTags
	}

	return tag, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List lists tags
//
// CourseTags can be included by setting the `IncludeRelations` field in the dbParams. The coutseTags
// can then be ordered by setting the `OrderBy` field in the dbParams, specifically referencing
// courses_tags.[column]
func (dao *TagDao) List(dbParams *database.DatabaseParams, tx *database.Tx) ([]*models.Tag, error) {
	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	selectColumns, orderByColumns := tableColumnsOrPanic(models.Tag{}, dao.Table())

	// Backup the original order by then remove invalid orderBy columns
	origOrderBy := dbParams.OrderBy
	dbParams.OrderBy = genericProcessOrderBy(dbParams.OrderBy, orderByColumns, dao, false)

	// Default the columns if not specified
	if len(dbParams.Columns) == 0 {
		dbParams.Columns = selectColumns
	}

	tags, err := genericList(dao, dbParams, dao.scanRow, tx)
	if err != nil {
		return nil, err
	}

	// Get the course_tags
	courseTagDao := NewCourseTagDao(dao.db)
	if len(tags) > 0 && slices.Contains(dbParams.IncludeRelations, courseTagDao.Table()) {
		// Get the tag IDs
		tagIds := []string{}
		for _, t := range tags {
			tagIds = append(tagIds, t.ID)
		}

		_, orderByColumns := tableColumnsOrPanic(models.CourseTag{}, courseTagDao.Table())

		// Reduce the order by clause to only include columns specific to the course_tags table
		reducedOrderBy := genericProcessOrderBy(origOrderBy, orderByColumns, courseTagDao, true)

		dbParams = &database.DatabaseParams{
			OrderBy: reducedOrderBy,
			Where:   squirrel.Eq{"tag_id": tagIds},
		}

		// Get the course_tags
		courseTags, err := courseTagDao.List(dbParams, tx)
		if err != nil {
			return nil, err
		}

		// Map the course_tags to the tags
		tagMap := map[string][]*models.CourseTag{}
		for _, ct := range courseTags {
			tagMap[ct.TagId] = append(tagMap[ct.TagId], ct)
		}

		// Assign the course_tags to the tags
		for _, t := range tags {
			t.CourseTags = tagMap[t.ID]
		}
	}

	return tags, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update updates the tag column in a tag
func (dao *TagDao) Update(tag *models.Tag, tx *database.Tx) error {
	if tag.ID == "" {
		return ErrEmptyId
	}

	tag.RefreshUpdatedAt()

	// Convert to a map so we have the rendered values
	data := modelToMapOrPanic(tag)

	query, args, _ := squirrel.
		StatementBuilder.
		Update(dao.Table()).
		Set("tag", data["tag"]).
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

// Delete deletes tags based upon the where clause
func (dao *TagDao) Delete(dbParams *database.DatabaseParams, tx *database.Tx) error {
	return genericDelete(dao, dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// countSelect returns the default count select builder
func (dao *TagDao) countSelect() squirrel.SelectBuilder {
	courseTagDao := NewCourseTagDao(dao.db)

	return dao.BaseDao.countSelect().
		LeftJoin(courseTagDao.Table() + " ON " + dao.Table() + ".id = " + courseTagDao.Table() + ".tag_id").
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *TagDao) baseSelect() squirrel.SelectBuilder {
	return dao.countSelect().GroupBy("tags.id")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans an tag row
func (dao *TagDao) scanRow(scannable Scannable) (*models.Tag, error) {
	var t models.Tag

	err := scannable.Scan(
		&t.ID,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.Tag,
		&t.CourseCount,
	)

	if err != nil {
		return nil, err
	}

	return &t, nil
}
