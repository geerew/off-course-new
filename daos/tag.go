package daos

import (
	"database/sql"
	"slices"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TagDao is the data access object for tags
type TagDao struct {
	db    database.Database
	Table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTagDao returns a new TagDao
func NewTagDao(db database.Database) *TagDao {
	return &TagDao{
		db:    db,
		Table: "tags",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count returns the number of tags
func (dao *TagDao) Count(params *database.DatabaseParams) (int, error) {
	generic := NewGenericDao(dao.db, dao.Table, dao)
	return generic.Count(params, nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create inserts a new tag
//
// `tx` allows for the function to be run within a transaction
func (dao *TagDao) Create(t *models.Tag, tx *sql.Tx) error {
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
		Insert(dao.Table).
		SetMap(dao.data(t)).
		ToSql()

	_, err := execFn(query, args...)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get selects a tag with the given ID or name
//
// `tx` allows for the function to be run within a transaction
func (dao *TagDao) Get(id string, byName bool, dbParams *database.DatabaseParams, tx *sql.Tx) (*models.Tag, error) {
	generic := NewGenericDao(dao.db, dao.Table, dao)

	tagDbParams := &database.DatabaseParams{
		Columns: dao.columns(),
	}

	if byName {
		tagDbParams.Where = squirrel.Eq{dao.Table + ".tag": id}
	} else {
		tagDbParams.Where = squirrel.Eq{dao.Table + ".id": id}
	}

	row, err := generic.Get(tagDbParams, tx)
	if err != nil {
		return nil, err
	}

	tag, err := dao.scanRow(row)
	if err != nil {
		return nil, err
	}

	// Get the course tags
	courseTagDao := NewCourseTagDao(dao.db)
	if dbParams != nil && slices.Contains(dbParams.IncludeRelations, courseTagDao.Table) {
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

// List selects tags
//
// `tx` allows for the function to be run within a transaction
func (dao *TagDao) List(dbParams *database.DatabaseParams, tx *sql.Tx) ([]*models.Tag, error) {
	generic := NewGenericDao(dao.db, dao.Table, dao)

	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	origOrderBy := dbParams.OrderBy

	dbParams.OrderBy = dao.ProcessOrderBy(dbParams.OrderBy, false)

	// Default the columns if not specified
	if len(dbParams.Columns) == 0 {
		dbParams.Columns = dao.columns()
	}

	rows, err := generic.List(dbParams, tx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*models.Tag
	tagIds := []string{}

	for rows.Next() {
		t, err := dao.scanRow(rows)
		if err != nil {
			return nil, err
		}

		tags = append(tags, t)
		tagIds = append(tagIds, t.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Get the course_tags
	courseTagDao := NewCourseTagDao(dao.db)
	if len(tags) > 0 && slices.Contains(dbParams.IncludeRelations, courseTagDao.Table) {
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

// Delete deletes a tag based upon the where clause
//
// `tx` allows for the function to be run within a transaction
func (dao *TagDao) Delete(dbParams *database.DatabaseParams, tx *sql.Tx) error {
	if dbParams == nil || dbParams.Where == nil {
		return ErrMissingWhere
	}

	generic := NewGenericDao(dao.db, dao.Table, dao)
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

	generic := NewGenericDao(dao.db, dao.Table, dao)
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
		From(dao.Table).
		LeftJoin(courseTagDao.Table + " ON " + dao.Table + ".id = " + courseTagDao.Table + ".tag_id").
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
		dao.Table + ".*",
		"COALESCE(COUNT(" + courseTagDao.Table + ".id), 0) AS course_count",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// data generates a map of key/values for a tag
func (dao *TagDao) data(t *models.Tag) map[string]any {
	return map[string]any{
		"id":         t.ID,
		"tag":        NilStr(t.Tag),
		"created_at": t.CreatedAt,
		"updated_at": t.UpdatedAt,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans an tag row
func (dao *TagDao) scanRow(scannable Scannable) (*models.Tag, error) {
	var t models.Tag

	err := scannable.Scan(
		&t.ID,
		&t.Tag,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.CourseCount,
	)

	if err != nil {
		return nil, err
	}

	return &t, nil
}
