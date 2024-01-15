package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Defines a model for the table `courses_progress`
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"errors"
	"math"

	sq "github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/types"
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
	builder := coursesProgressBaseSelect().Columns("COUNT(DISTINCT " + TableCoursesProgress() + ".id)")

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
	builder := coursesProgressBaseSelect().Columns(TableCoursesProgress() + ".*")

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

	builder := coursesProgressBaseSelect().
		Columns(TableCoursesProgress() + ".*").
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

// UpdateCourseProgressStarted updates `started` and `started_at`. If `started` is true,
// `started_at` is set to the current time. If `started` is false `started_at` is set to null
func UpdateCourseProgressStarted(db database.Database, courseId string, started bool) (*CourseProgress, error) {
	if courseId == "" {
		return nil, errors.New("id cannot be empty")
	}

	cp, err := GetCourseProgress(db, courseId)
	if err != nil {
		return nil, err
	}

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

// UpdateCourseProgressPercent updates `percent` and `completed_at`. If `percent` is 100,
// `completed at` is set to the current time. If `percent` is less than 100, `completed_at` is
// set to null. This can be called at any time but mainly when an asset is marked as completed/not
// completed
func UpdateCourseProgressPercent(db database.Database, courseId string) (*CourseProgress, error) {
	if courseId == "" {
		return nil, errors.New("id cannot be empty")
	}

	// Get the current course progress
	cp, err := GetCourseProgress(db, courseId)
	if err != nil {
		return nil, err
	}

	// Calculate the percent
	percent, err := coursesProgressPercent(db, courseId)
	if err != nil {
		return nil, err
	}

	if cp.Percent == percent {
		// Nothing to do
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

	// Ensure started when percent is greater than 0
	if percent > 0 && !cp.Started {
		builder = builder.Set("started", true).Set("started_at", updatedAt)
	}

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

	if percent > 0 && !cp.Started {
		cp.Started = true
		cp.StartedAt = updatedAt
	}

	return cp, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// coursesProgressBaseSelect returns a select builder for the courses_progress table. It does not
// include any columns by default and as such, you must specify the columns with `.Columns(...)`
func coursesProgressBaseSelect() sq.SelectBuilder {
	return sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select("").
		From(TableCoursesProgress()).
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// coursesProgressPercent calculates the completed percent of a course based upon how many assets
// have been completed
func coursesProgressPercent(db database.Database, id string) (int, error) {
	// Count the number of assets for this course
	assetCount, err := CountAssets(db, &database.DatabaseParams{Where: sq.Eq{TableAssets() + ".course_id": id}})
	if err != nil {
		return -1, err
	}

	// When there are no assets, the percent is 0
	if assetCount == 0 {
		return 0, nil
	}

	// Count the number of assets that have been completed
	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select("COUNT(DISTINCT " + TableAssetsProgress() + ".id)").
		From(TableAssetsProgress()).
		Where(sq.And{
			sq.Eq{TableAssetsProgress() + ".course_id": id},
			sq.Eq{TableAssetsProgress() + ".completed": true},
		})

	query, args, err := builder.ToSql()
	if err != nil {
		return -1, err
	}

	var completedCount int
	err = db.QueryRow(query, args...).Scan(&completedCount)
	if err != nil {
		return -1, err
	}

	return int(math.Abs((float64(completedCount) * float64(100)) / float64(assetCount))), nil
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
