package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Defines a model for the table `courses`
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"
	"errors"

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
	Title    string
	Path     string
	CardPath string

	// --------------------------------
	// Not in this table, but added via a join
	// --------------------------------

	// Scan status
	ScanStatus string

	// Course Progress
	Started     bool
	StartedAt   types.DateTime
	Percent     int
	CompletedAt types.DateTime
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TableCourses returns the table name for the courses table
func TableCourses() string {
	return "courses"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CountCourses counts the number of courses
func CountCourses(db database.Database, params *database.DatabaseParams) (int, error) {
	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select("COUNT(*)").
		From(TableCourses())

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
		TableScans() + ".status",
		TableCoursesProgress() + ".started",
		TableCoursesProgress() + ".started_at",
		TableCoursesProgress() + ".percent",
		TableCoursesProgress() + ".completed_at",
	}

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select(cols...).
		From(TableCourses()).
		LeftJoin(TableScans() + " ON " + TableCourses() + ".id = " + TableScans() + ".course_id").
		LeftJoin(TableCoursesProgress() + " ON " + TableCourses() + ".id = " + TableCoursesProgress() + ".course_id")

	if params != nil {
		// ORDER BY
		if params != nil && len(params.OrderBy) > 0 {
			builder = builder.OrderBy(params.OrderBy...)
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
		TableScans() + ".status",
		TableCoursesProgress() + ".started",
		TableCoursesProgress() + ".started_at",
		TableCoursesProgress() + ".percent",
		TableCoursesProgress() + ".completed_at",
	}

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select(cols...).
		From(TableCourses()).
		LeftJoin(TableScans() + " ON " + TableCourses() + ".id = " + TableScans() + ".course_id").
		LeftJoin(TableCoursesProgress() + " ON " + TableCourses() + ".id = " + TableCoursesProgress() + ".course_id").
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

// CreateCourse inserts a new course
func CreateCourse(db database.Database, c *Course) error {
	c.RefreshId()
	c.RefreshCreatedAt()
	c.RefreshUpdatedAt()

	builder := sq.StatementBuilder.
		Insert(TableCourses()).
		Columns("id", "title", "path", "card_path", "created_at", "updated_at").
		Values(c.ID, NilStr(c.Title), NilStr(c.Path), NilStr(c.CardPath), c.CreatedAt, c.UpdatedAt)

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(query, args...)
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

// scanCourseRow scans a course row
func scanCourseRow(scannable Scannable) (*Course, error) {
	var c Course

	// Nullable fields
	var cardPath sql.NullString
	var scanStatus sql.NullString
	var started sql.NullBool
	var percent sql.NullInt16

	err := scannable.Scan(
		// Course
		&c.ID,
		&c.Title,
		&c.Path,
		&cardPath,
		&c.CreatedAt,
		&c.UpdatedAt,
		// Scan
		&scanStatus,
		// Course progress
		&started,
		&c.StartedAt,
		&percent,
		&c.CompletedAt,
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

	c.Started = started.Bool

	if percent.Valid {
		c.Percent = int(percent.Int16)
	}

	return &c, nil
}
