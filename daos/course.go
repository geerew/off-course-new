package daos

import (
	"database/sql"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseDao is the data access object for courses
type CourseDao struct {
	db    database.Database
	table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewCourseDao returns a new CourseDao
func NewCourseDao(db database.Database) *CourseDao {
	return &CourseDao{
		db:    db,
		table: TableCourses(),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TableCourses returns the name of the courses table
func TableCourses() string {
	return "courses"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count returns the number of courses
func (dao *CourseDao) Count(params *database.DatabaseParams) (int, error) {
	generic := NewGenericDao(dao.db, dao.table)
	return generic.Count(dao.baseSelect(), params, nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create inserts a new course and courses_progress row within a transaction
//
// NOTE: There is currently no support for users, but when there is, the default courses_progress
// should be inserted for the admin user
func (dao *CourseDao) Create(c *models.Course) error {
	if c.ID == "" {
		c.RefreshId()
	}

	c.RefreshCreatedAt()
	c.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.table).
		SetMap(dao.data(c)).
		ToSql()

	return dao.db.RunInTransaction(func(tx *sql.Tx) error {
		// Create the course
		if _, err := tx.Exec(query, args...); err != nil {
			return err
		}

		// Create the course progress
		cp := &models.CourseProgress{
			CourseID: c.ID,
		}

		cpDao := NewCourseProgressDao(dao.db)
		return cpDao.Create(cp, tx)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get selects a course with the given ID
//
// `tx` allows for the function to be run within a transaction
func (dao *CourseDao) Get(id string, tx *sql.Tx) (*models.Course, error) {
	generic := NewGenericDao(dao.db, dao.table)

	dbParams := &database.DatabaseParams{
		Columns: dao.selectColumns(),
		Where:   squirrel.Eq{generic.table + ".id": id},
	}

	row, err := generic.Get(dao.baseSelect(), dbParams, tx)
	if err != nil {
		return nil, err
	}

	course, err := dao.scanRow(row)
	if err != nil {
		return nil, err
	}

	return course, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List selects courses
//
// `tx` allows for the function to be run within a transaction
func (dao *CourseDao) List(dbParams *database.DatabaseParams, tx *sql.Tx) ([]*models.Course, error) {
	generic := NewGenericDao(dao.db, dao.table)

	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	// Process the order by clauses
	dbParams.OrderBy = dao.processOrderBy(dbParams.OrderBy)

	// Default the columns if not specified
	if len(dbParams.Columns) == 0 {
		dbParams.Columns = dao.selectColumns()
	}

	rows, err := generic.List(dao.baseSelect(), dbParams, tx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []*models.Course

	for rows.Next() {
		c, err := dao.scanRow(rows)
		if err != nil {
			return nil, err
		}

		courses = append(courses, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return courses, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update updates a course
//
// Note: Only `card_path` and `available` can be updated
func (dao *CourseDao) Update(course *models.Course) error {
	if course.ID == "" {
		return ErrEmptyId
	}

	course.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Update(dao.table).
		Set("card_path", NilStr(course.CardPath)).
		Set("available", course.Available).
		Set("updated_at", course.UpdatedAt).
		Where("id = ?", course.ID).
		ToSql()

	_, err := dao.db.Exec(query, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes a course with the given ID
//
// `tx` allows for the function to be run within a transaction
func (dao *CourseDao) Delete(dbParams *database.DatabaseParams, tx *sql.Tx) error {
	if dbParams == nil || dbParams.Where == nil {
		return ErrMissingWhere
	}

	generic := NewGenericDao(dao.db, dao.table)
	return generic.Delete(dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
//
// It performs 2 left joins
//   - scans table to get `scan_status`
//   - courses progress table to get `started`, `started_at`, `percent`, and `completed_at`
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *CourseDao) baseSelect() squirrel.SelectBuilder {
	return squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.table).
		LeftJoin(TableScans() + " ON " + dao.table + ".id = " + TableScans() + ".course_id").
		LeftJoin(TableCoursesProgress() + " ON " + dao.table + ".id = " + TableCoursesProgress() + ".course_id").
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// selectColumns returns the columns to select
func (dao *CourseDao) selectColumns() []string {
	return []string{
		dao.table + ".*",
		TableScans() + ".status as scan_status",
		TableCoursesProgress() + ".started",
		TableCoursesProgress() + ".started_at",
		TableCoursesProgress() + ".percent",
		TableCoursesProgress() + ".completed_at",
		TableCoursesProgress() + ".updated_at as progress_updated_at",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// data generates a map of key/values for a course
func (dao *CourseDao) data(c *models.Course) map[string]any {
	return map[string]any{
		"id":         c.ID,
		"title":      NilStr(c.Title),
		"path":       NilStr(c.Path),
		"card_path":  NilStr(c.CardPath),
		"available":  c.Available,
		"created_at": c.CreatedAt,
		"updated_at": c.UpdatedAt,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// processOrderBy takes an array of strings representing orderBy clauses and returns a processed
// version of this array
//
// It will creates a new list of valid table columns based upon selectColumns() for the current
// DAO. Additionally, it handles the special case of 'scan_status' column, which requires custom
// sorting logic, via a CASE statement.
//
// The custom sorting logic is defined as follows:
//   - NULL values are treated as the lowest value (sorted first in ASC, last in DESC)
//   - 'waiting' status is treated as the second value
//   - 'processing' status is treated as the third value
func (dao *CourseDao) processOrderBy(orderBy []string) []string {
	if len(orderBy) == 0 {
		return orderBy
	}

	validTableColumns := dao.selectColumns()
	var processedOrderBy []string

	for _, ob := range orderBy {
		table, column := extractTableColumn(ob)

		if isValidOrderBy(table, column, validTableColumns) {
			// When the column is 'scan_status', apply the custom sorting logic
			if column == "scan_status" || table+"."+column == TableScans()+".status" {
				// Determine the sort direction, defaulting to ASC if not specified
				parts := strings.Fields(ob)
				sortDirection := "ASC"
				if len(parts) > 1 {
					sortDirection = strings.ToUpper(parts[1])
				}

				caseStmt := "CASE " +
					"WHEN scan_status IS NULL THEN 1 " +
					"WHEN scan_status = 'waiting' THEN 2 " +
					"WHEN scan_status = 'processing' THEN 3 " +
					"END " + sortDirection

				processedOrderBy = append(processedOrderBy, caseStmt)
			} else {
				processedOrderBy = append(processedOrderBy, ob)
			}
		}
	}

	return processedOrderBy
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans a course row
func (dao *CourseDao) scanRow(scannable Scannable) (*models.Course, error) {
	var c models.Course

	// Nullable fields
	var cardPath sql.NullString
	var scanStatus sql.NullString

	err := scannable.Scan(
		// Course
		&c.ID,
		&c.Title,
		&c.Path,
		&cardPath,
		&c.Available,
		&c.CreatedAt,
		&c.UpdatedAt,
		// Scan
		&scanStatus,
		// Course progress
		&c.Started,
		&c.StartedAt,
		&c.Percent,
		&c.CompletedAt,
		&c.ProgressUpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if cardPath.Valid {
		c.CardPath = cardPath.String
	}

	if scanStatus.Valid {
		c.ScanStatus = scanStatus.String
	}

	return &c, nil
}
