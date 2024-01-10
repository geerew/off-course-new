package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Defines a model for the table `courses_progress`
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseProgress defines the model for a course progress
//
// As this changes, update `scanCourseProgressRow()`
type CourseProgress struct {
	BaseModel

	CourseID    string
	Started     bool
	StartedAt   types.DateTime
	Percent     int
	CompletedAt types.DateTime
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TableCoursesProgress returns the table name for the courses progress table
func TableCoursesProgress() string {
	return "courses_progress"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CountCoursesProgress counts the number of courses progress
func CountCoursesProgress(db database.Database, params *database.DatabaseParams) (int, error) {
	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select("COUNT(*)").
		From(TableCoursesProgress())

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

// GetCoursesProgress selects courses progress
func GetCoursesProgress(db database.Database, params *database.DatabaseParams) ([]*CourseProgress, error) {
	var cps []*CourseProgress

	// Start building the query
	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select(TableCoursesProgress() + ".*").
		From(TableCoursesProgress())

	// Add additional clauses
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
			if builder, err = paginate(db, params, builder, CountCoursesProgress); err != nil {
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
		cp, err := scanCourseProgressRow(rows)
		if err != nil {
			return nil, err
		}

		cps = append(cps, cp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cps, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetCourseProgress selects a course progress for the given course ID
func GetCourseProgress(db database.Database, courseId string) (*CourseProgress, error) {
	if courseId == "" {
		return nil, errors.New("id cannot be empty")
	}

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select(TableCoursesProgress() + ".*").
		From(TableCoursesProgress()).
		Where(sq.Eq{TableCoursesProgress() + ".course_id": courseId})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	row := db.QueryRow(query, args...)
	if row.Err() != nil {
		return nil, row.Err()
	}

	cp, err := scanCourseProgressRow(row)
	if err != nil {
		return nil, err
	}

	return cp, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateCourseProgress inserts a new course progress
func CreateCourseProgress(db database.Database, cp *CourseProgress) error {
	cp.RefreshId()
	cp.RefreshCreatedAt()
	cp.RefreshUpdatedAt()

	builder := sq.StatementBuilder.
		Insert(TableCoursesProgress()).
		Columns("id", "course_id", "started", "started_at", "percent", "completed_at", "created_at", "updated_at").
		Values(cp.ID, NilStr(cp.CourseID), cp.Started, cp.StartedAt, cp.Percent, cp.CompletedAt, cp.CreatedAt, cp.UpdatedAt)

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(query, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateCourseProgressStarted updates `started` and `started_at`. When a course progress entry
// does not exist, it will be created. When an entry does exist, it will be updated.
//
// If `started` is true, `started_at` is set to the current time. If `started` is false,
// `started_at` is set to null
func UpdateCourseProgressStarted(db database.Database, courseId string, started bool) (*CourseProgress, error) {
	if courseId == "" {
		return nil, errors.New("id cannot be empty")
	}

	cp, err := GetCourseProgress(db, courseId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// --------------------------------
	// Create if it does not exist
	// --------------------------------
	if err == sql.ErrNoRows {
		cp = &CourseProgress{
			CourseID: courseId,
			Started:  started,
		}

		if started {
			cp.StartedAt = types.NowDateTime()
		}

		err = CreateCourseProgress(db, cp)
		if err != nil {
			return nil, err
		}

		return cp, nil
	}

	// --------------------------------
	// Update existing
	// --------------------------------

	// Nothing to do
	if cp.Started == started {
		return cp, nil
	}

	updatedAt := types.NowDateTime()

	var startedAt types.DateTime
	if started {
		startedAt = updatedAt
	}

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Update(TableCoursesProgress()).
		Set("started", started).
		Set("started_at", startedAt).
		Set("updated_at", updatedAt).
		Where("id = ?", cp.ID)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	cp.Started = started
	cp.StartedAt = startedAt
	cp.UpdatedAt = updatedAt

	return cp, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateCourseProgressPercent updates `percent` and `completed_at`. When a course progress entry
// does not exist, it will be created. When an entry does exist, it will be updated.
//
// The percent should be based off of the number of completed assets for this course. When an asset
// is completed, the percent should be updated. If `percent` is 100, `completed at` is set to
// the current time. If `percent` is less than 100, `completed_at` is set to null
func UpdateCourseProgressPercent(db database.Database, courseId string, percent int) (*CourseProgress, error) {
	if courseId == "" {
		return nil, errors.New("id cannot be empty")
	}

	cp, err := GetCourseProgress(db, courseId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Keep the percent between 0 and 100
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}

	// --------------------------------
	// Create if it does not exist
	// --------------------------------
	if err == sql.ErrNoRows {
		cp = &CourseProgress{
			CourseID: courseId,
			Percent:  percent,
		}

		if percent == 100 {
			cp.CompletedAt = types.NowDateTime()
		}

		err = CreateCourseProgress(db, cp)
		if err != nil {
			return nil, err
		}

		return cp, nil
	}

	// --------------------------------
	// Update existing
	// --------------------------------

	// Nothing to do
	if cp.Percent == percent {
		return cp, nil
	}

	updatedAt := types.NowDateTime()

	var completedAt types.DateTime
	if percent == 100 {
		completedAt = updatedAt
	}

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Update(TableCoursesProgress()).
		Set("percent", percent).
		Set("completed_at", completedAt).
		Set("updated_at", updatedAt).
		Where("id = ?", cp.ID)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	cp.Percent = percent
	cp.CompletedAt = completedAt
	cp.UpdatedAt = updatedAt

	return cp, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTestCourseProgress creates a course progress for each course in the slice. If a db is
// provided, a DB insert will be performed
//
// THIS IS FOR TESTING PURPOSES
func NewTestCoursesProgress(t *testing.T, db database.Database, courses []*Course) []*CourseProgress {
	cps := []*CourseProgress{}

	for i := 0; i < len(courses); i++ {
		cp := &CourseProgress{}

		cp.RefreshId()
		cp.RefreshCreatedAt()
		cp.RefreshUpdatedAt()

		cp.CourseID = courses[i].ID

		if db != nil {
			err := CreateCourseProgress(db, cp)
			require.Nil(t, err)

			// This allows the created/updated times to be different when inserting multiple rows
			time.Sleep(time.Millisecond * 1)
		}

		cps = append(cps, cp)
	}

	return cps
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanCourseProgressRow scans a course progress row
func scanCourseProgressRow(scannable Scannable) (*CourseProgress, error) {
	var cp CourseProgress

	err := scannable.Scan(
		&cp.ID,
		&cp.CourseID,
		&cp.Started,
		&cp.StartedAt,
		&cp.Percent,
		&cp.CompletedAt,
		&cp.CreatedAt,
		&cp.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &cp, nil
}
