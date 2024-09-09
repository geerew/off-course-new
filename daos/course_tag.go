package daos

import (
	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseTagDao is the data access object for courses tags
type CourseTagDao struct {
	db    database.Database
	table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewCourseTagDao returns a new CourseTagDao
func NewCourseTagDao(db database.Database) *CourseTagDao {
	return &CourseTagDao{
		db:    db,
		table: "courses_tags",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table returns the table name
func (dao *CourseTagDao) Table() string {
	return dao.table
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count the course tags
func (dao *CourseTagDao) Count(dbParams *database.DatabaseParams, tx *database.Tx) (int, error) {
	queryRowFn := dao.db.QueryRow
	if tx != nil {
		queryRowFn = tx.QueryRow
	}

	return GenericCount(dao.countSelect(), dao.Table(), dbParams, queryRowFn)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create a course tag and the tag itself if it does not exist
//
// A new transaction is created if `tx` is nil
func (dao *CourseTagDao) Create(ct *models.CourseTag, tx *database.Tx) error {
	createFn := func(tx *database.Tx) error {
		if ct.Tag == "" {
			return ErrMissingTag
		}

		if ct.CourseId == "" {
			return ErrMissingCourseId
		}

		if ct.ID == "" {
			ct.RefreshId()
		}

		ct.RefreshCreatedAt()
		ct.RefreshUpdatedAt()

		// Check if the tag exists. Tags are unique so this wil return 0 or 1
		tagDao := NewTagDao(dao.db)
		tags, err := tagDao.List(&database.DatabaseParams{Where: squirrel.Eq{"tag": ct.Tag}}, tx)
		if err != nil {
			return err
		}

		// Create the tag if it doesn't exist
		if len(tags) == 0 {
			tag := &models.Tag{
				Tag: ct.Tag,
			}

			if err := tagDao.Create(tag, tx); err != nil {
				return err
			}

			ct.TagId = tag.ID
		} else {
			ct.TagId = tags[0].ID
		}

		// Insert the course-tag
		query, args, _ := squirrel.
			StatementBuilder.
			Insert(dao.Table()).
			SetMap(dao.data(ct)).
			ToSql()

		_, err = tx.Exec(query, args...)

		return err
	}

	if tx == nil {
		return dao.db.RunInTransaction(func(tx *database.Tx) error {
			return createFn(tx)
		})
	} else {
		return createFn(tx)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List lists course tags
func (dao *CourseTagDao) List(dbParams *database.DatabaseParams, tx *database.Tx) ([]*models.CourseTag, error) {
	generic := NewGenericDao(dao.db, dao)

	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	// Process the order by clauses
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

	var cts []*models.CourseTag

	for rows.Next() {
		ct, err := dao.scanRow(rows)
		if err != nil {
			return nil, err
		}

		cts = append(cts, ct)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cts, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListCourseIdsByTags lists course IDs containing all tags in the slice
func (dao *CourseTagDao) ListCourseIdsByTags(tags []string, dbParams *database.DatabaseParams) ([]string, error) {
	if len(tags) == 0 {
		return nil, nil
	}

	generic := NewGenericDao(dao.db, dao)

	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	dbParams.OrderBy = dao.ProcessOrderBy(dbParams.OrderBy, false)
	dbParams.Columns = []string{dao.Table() + ".course_id"}
	dbParams.Where = squirrel.Eq{NewTagDao(dao.db).Table() + ".tag": tags}
	dbParams.GroupBys = []string{dao.Table() + ".course_id"}
	dbParams.Having = squirrel.Expr("COUNT(DISTINCT "+NewTagDao(dao.db).Table()+".tag) = ?", len(tags))
	dbParams.Pagination = nil

	rows, err := generic.List(dbParams, nil)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courseIds []string
	for rows.Next() {
		var courseId string
		if err := rows.Scan(&courseId); err != nil {
			return nil, err
		}

		courseIds = append(courseIds, courseId)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return courseIds, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes course tags based upon the where clause
func (dao *CourseTagDao) Delete(dbParams *database.DatabaseParams, tx *database.Tx) error {
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
func (dao *CourseTagDao) ProcessOrderBy(orderBy []string, explicit bool) []string {
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
func (dao *CourseTagDao) countSelect() squirrel.SelectBuilder {
	tagDao := NewTagDao(dao.db)
	courseDao := NewCourseDao(dao.db)

	return squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.Table()).
		LeftJoin(courseDao.Table() + " ON " + dao.Table() + ".course_id = " + courseDao.Table() + ".id").
		LeftJoin(tagDao.Table() + " ON " + dao.Table() + ".tag_id = " + tagDao.Table() + ".id").
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// baseSelect returns the default select builder
//
// It performs 2 left joins
//   - courses table to get `title`
//   - tags table to get `tag`
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *CourseTagDao) baseSelect() squirrel.SelectBuilder {
	return dao.countSelect()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// columns returns the columns to select
func (dao *CourseTagDao) columns() []string {
	tagDao := NewTagDao(dao.db)
	courseDao := NewCourseDao(dao.db)

	return []string{
		dao.Table() + ".*",
		courseDao.Table() + ".title as course",
		tagDao.Table() + ".tag",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// data generates a map of key/values for a course-tag
func (dao *CourseTagDao) data(ct *models.CourseTag) map[string]any {
	return map[string]any{
		"id":         ct.ID,
		"tag_id":     NilStr(ct.TagId),
		"course_id":  NilStr(ct.CourseId),
		"created_at": FormatTime(ct.CreatedAt),
		"updated_at": FormatTime(ct.UpdatedAt),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans a course-tag row
func (dao *CourseTagDao) scanRow(scannable Scannable) (*models.CourseTag, error) {
	var ct models.CourseTag

	var createdAt string
	var updatedAt string

	err := scannable.Scan(
		&ct.ID,
		&ct.TagId,
		&ct.CourseId,
		&createdAt,
		&updatedAt,
		&ct.Course,
		&ct.Tag,
	)

	if err != nil {
		return nil, err
	}

	if ct.CreatedAt, err = ParseTime(createdAt); err != nil {
		return nil, err
	}

	if ct.UpdatedAt, err = ParseTime(updatedAt); err != nil {
		return nil, err
	}

	return &ct, nil
}
