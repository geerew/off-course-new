package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Defines a model for the table `courses`
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Course defines the model for a course
//
// As this changes, update `scanCourseRow()`
type Course struct {
	BaseModel
	Title     string
	Path      string
	CardPath  string
	Available bool

	// --------------------------------
	// Not in this table, but added via a join
	// --------------------------------

	// Scan status
	ScanStatus string

	// Course Progress
	Started           bool
	StartedAt         types.DateTime
	Percent           int
	CompletedAt       types.DateTime
	ProgressUpdatedAt types.DateTime
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TableCourses returns the table name for the courses table
func TableCourses() string {
	return "courses"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CountCourses counts the number of courses
func CountCourses(db database.Database, params *database.DatabaseParams) (int, error) {
	builder := coursesBaseSelect().Columns("COUNT(DISTINCT " + TableCourses() + ".id)")

	// Add where clauses if necessary
	if params != nil && params.Where != "" {
		builder = builder.Where(params.Where)
	}

	// Build the query
	query, args, err := builder.ToSql()
	if err != nil {
		return -1, err
	}

	// Execute the query
	var count int
	err = db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetCourses selects courses
//
// It performs lefts joins
//   - scans table to set `scan_status`
//   - courses progress table to set `started`, `started_at`, `percent`, and `completed_at`
func GetCourses(db database.Database, params *database.DatabaseParams) ([]*Course, error) {
	var courses []*Course

	cols := []string{
		TableCourses() + ".*",
		TableScans() + ".status as scan_status",
		TableCoursesProgress() + ".started",
		TableCoursesProgress() + ".started_at",
		TableCoursesProgress() + ".percent",
		TableCoursesProgress() + ".completed_at",
		TableCoursesProgress() + ".updated_at as progress_updated_at",
	}

	builder := coursesBaseSelect().Columns(cols...)

	if params != nil {
		// ORDER BY
		if params != nil && len(params.OrderBy) > 0 {
			builder = coursesOrderBy(builder, params)
		}

		// WHERE
		if params.Where != "" {
			builder = builder.Where(params.Where)
		}

		// PAGINATION
		if params.Pagination != nil {
			var err error
			if builder, err = paginate(db, params, builder, CountCourses); err != nil {
				return nil, err
			}
		}
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		c, err := scanCourseRow(rows)
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

// GetCourse selects a course for the given ID
//
// It performs lefts joins
//   - scans table to set `scan_status`
//   - courses progress table to set `started`, `started_at`, `percent`, and `completed_at`
func GetCourse(db database.Database, id string) (*Course, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	cols := []string{
		TableCourses() + ".*",
		TableScans() + ".status as scan_status",
		TableCoursesProgress() + ".started",
		TableCoursesProgress() + ".started_at",
		TableCoursesProgress() + ".percent",
		TableCoursesProgress() + ".completed_at",
		TableCoursesProgress() + ".updated_at as progress_updated_at",
	}

	builder := coursesBaseSelect().
		Columns(cols...).
		Where(sq.Eq{TableCourses() + ".id": id})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	row := db.QueryRow(query, args...)
	if row.Err() != nil {
		return nil, row.Err()
	}

	course, err := scanCourseRow(row)
	if err != nil {
		return nil, err
	}

	return course, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateCourse inserts a new course. It also inserts a courses_progress row for the course
//
// NOTE: There is currently no support for users, but when there is, the default courses_progress
// should be inserted for the admin user
func CreateCourse(db database.Database, c *Course) error {
	c.RefreshId()
	c.RefreshCreatedAt()
	c.RefreshUpdatedAt()

	builder := sq.StatementBuilder.
		Insert(TableCourses()).
		Columns("id", "title", "path", "card_path", "available", "created_at", "updated_at").
		Values(c.ID, NilStr(c.Title), NilStr(c.Path), NilStr(c.CardPath), c.Available, c.CreatedAt, c.UpdatedAt)

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(query, args...)
	if err != nil {
		return err
	}

	// Insert a courses_progress row for this course
	cp := &CourseProgress{
		CourseID: c.ID,
	}
	err = CreateCourseProgress(db, cp)

	return err

}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateCourseCardPath updates `card_path`
func UpdateCourseCardPath(db database.Database, id string, newCardPath string) (*Course, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	c, err := GetCourse(db, id)
	if err != nil {
		return nil, err
	}

	// Nothing to do
	if c.CardPath == newCardPath {
		return c, nil
	}

	updatedAt := types.NowDateTime()

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Update(TableCourses()).
		Set("card_path", NilStr(newCardPath)).
		Set("updated_at", updatedAt).
		Where("id = ?", c.ID)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	c.CardPath = newCardPath
	c.UpdatedAt = updatedAt

	return c, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateCourseAvailability updates `available`
func UpdateCourseAvailability(db database.Database, id string, available bool) (*Course, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	c, err := GetCourse(db, id)
	if err != nil {
		return nil, err
	}

	// Nothing to do
	if c.Available == available {
		return c, nil
	}

	updatedAt := types.NowDateTime()

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Update(TableCourses()).
		Set("available", available).
		Set("updated_at", updatedAt).
		Where("id = ?", c.ID)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	c.Available = available
	c.UpdatedAt = updatedAt

	return c, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateCourseUpdatedAt updates `updated_at`
func UpdateCourseUpdatedAt(db database.Database, id string) (*Course, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	c, err := GetCourse(db, id)
	if err != nil {
		return nil, err
	}

	updatedAt := types.NowDateTime()

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Update(TableCourses()).
		Set("updated_at", updatedAt).
		Where("id = ?", c.ID)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	c.UpdatedAt = updatedAt

	return c, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteCourse deletes a course with the given ID
func DeleteCourse(db database.Database, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Delete(TableCourses()).
		Where(sq.Eq{"id": id})

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(query, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// coursesBaseSelect returns a select builder for the courses table. It does not include any columns
// by default and as such, you must specify the columns with `.Columns(...)`
func coursesBaseSelect() sq.SelectBuilder {
	return sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select("").
		From(TableCourses()).
		LeftJoin(TableScans() + " ON " + TableCourses() + ".id = " + TableScans() + ".course_id").
		LeftJoin(TableCoursesProgress() + " ON " + TableCourses() + ".id = " + TableCoursesProgress() + ".course_id").
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// coursesOrderBy adds the order by clauses to the select builder. It handles the special case of
// ordering by `scan_status` by adding a case statement to the order by clause
func coursesOrderBy(builder sq.SelectBuilder, dbParams *database.DatabaseParams) sq.SelectBuilder {
	var newOrderBy []string

	for _, orderBy := range dbParams.OrderBy {
		// Split the order by string into column name and sort direction
		parts := strings.Fields(orderBy)
		columnName := parts[0]

		if columnName == "scan_status" {
			// Determine the sort direction, defaulting to ASC if not specified
			sortDirection := "ASC"
			if len(parts) > 1 {
				sortDirection = strings.ToUpper(parts[1])
			}

			caseStmt := "CASE " +
				"WHEN scan_status IS NULL THEN 1 " +
				"WHEN scan_status = 'waiting' THEN 2 " +
				"WHEN scan_status = 'processing' THEN 3 " +
				"END " + sortDirection

			newOrderBy = append(newOrderBy, caseStmt)
		} else {
			newOrderBy = append(newOrderBy, orderBy)
		}
	}

	if len(newOrderBy) > 0 {
		builder = builder.OrderBy(newOrderBy...)
	}

	return builder
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanCourseRow scans a course row
func scanCourseRow(scannable Scannable) (*Course, error) {
	var c Course

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
