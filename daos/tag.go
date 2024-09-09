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
	table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTagDao returns a new TagDao
func NewTagDao(db database.Database) *TagDao {
	return &TagDao{
		BaseDao: BaseDao{db: db},
		table:   "tags",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table returns the table name
func (dao *TagDao) Table() string {
	return dao.table
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count counts the tags
func (dao *TagDao) Count(dbParams *database.DatabaseParams, tx *database.Tx) (int, error) {
	return GenericCount(dao, dbParams, tx)
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
		SetMap(dao.data(t)).
		ToSql()

	_, err := execFn(query, args...)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get gets a tag with the given ID or name
func (dao *TagDao) Get(id string, byName bool, dbParams *database.DatabaseParams, tx *database.Tx) (*models.Tag, error) {
	tagDbParams := &database.DatabaseParams{
		Columns: dao.columns(),
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

	tag, err := GenericGet(dao, tagDbParams, dao.scanRow, tx)
	if err != nil {
		return nil, err
	}

	// Get the course tags
	courseTagDao := NewCourseTagDao(dao.db)
	if dbParams != nil && slices.Contains(dbParams.IncludeRelations, courseTagDao.Table()) {
		courseTagDbParams := &database.DatabaseParams{
			OrderBy: courseTagDao.ProcessOrderBy(dbParams.OrderBy, true),
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
func (dao *TagDao) List(dbParams *database.DatabaseParams, tx *database.Tx) ([]*models.Tag, error) {
	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	origOrderBy := dbParams.OrderBy

	dbParams.OrderBy = dao.ProcessOrderBy(dbParams.OrderBy, false)

	// Default the columns if not specified
	if len(dbParams.Columns) == 0 {
		dbParams.Columns = dao.columns()
	}

	tags, err := GenericList(dao, dbParams, dao.scanRow, tx)
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

		// Reduce the order by clause to only include columns specific to the course_tags table
		reducedOrderBy := courseTagDao.ProcessOrderBy(origOrderBy, true)

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

	query, args, _ := squirrel.
		StatementBuilder.
		Update(dao.Table()).
		Set("tag", NilStr(tag.Tag)).
		Set("updated_at", FormatTime(tag.UpdatedAt)).
		Where("id = ?", tag.ID).
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
	if dbParams == nil || dbParams.Where == nil {
		return ErrMissingWhere
	}

	generic := NewGenericDao(dao.db, dao)
	return generic.Delete(dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ProcessOrderBy takes an array of strings representing orderBy clauses and returns a processed
// version of this array
//
// It will creates a new list of valid table columns based upon columns() for the current
// DAO
func (dao *TagDao) ProcessOrderBy(orderBy []string, explicit bool) []string {
	if len(orderBy) == 0 {
		return orderBy
	}

	generic := NewGenericDao(dao.db, dao)
	return generic.ProcessOrderBy(orderBy, dao.columns(), explicit)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// countSelect returns the default count select builder
func (dao *TagDao) countSelect() squirrel.SelectBuilder {
	courseTagDao := NewCourseTagDao(dao.db)

	return squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.Table()).
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
	// GroupBy("tags.id", "tags.tag", "tags.created_at", "tags.updated_at").
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// columns returns the columns to select
func (dao *TagDao) columns() []string {
	courseTagDao := NewCourseTagDao(dao.db)

	return []string{
		dao.Table() + ".*",
		"COALESCE(COUNT(" + courseTagDao.Table() + ".id), 0) AS course_count",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// data generates a map of key/values for a tag
func (dao *TagDao) data(t *models.Tag) map[string]any {
	return map[string]any{
		"id":         t.ID,
		"tag":        NilStr(t.Tag),
		"created_at": FormatTime(t.CreatedAt),
		"updated_at": FormatTime(t.UpdatedAt),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans an tag row
func (dao *TagDao) scanRow(scannable Scannable) (*models.Tag, error) {
	var t models.Tag

	var createdAt string
	var updatedAt string

	err := scannable.Scan(
		&t.ID,
		&t.Tag,
		&createdAt,
		&updatedAt,
		&t.CourseCount,
	)

	if err != nil {
		return nil, err
	}

	if t.CreatedAt, err = ParseTime(createdAt); err != nil {
		return nil, err
	}

	if t.UpdatedAt, err = ParseTime(updatedAt); err != nil {
		return nil, err
	}

	return &t, nil
}
