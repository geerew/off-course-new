package daos

import (
	"database/sql"
	"math"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseProgressDao is the data access object for courses progress
type CourseProgressDao struct {
	BaseDao
	table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewCourseProgressDao returns a new CourseProgressDao
func NewCourseProgressDao(db database.Database) *CourseProgressDao {
	return &CourseProgressDao{
		BaseDao: BaseDao{db: db},
		table:   "courses_progress",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table returns the table name
func (dao *CourseProgressDao) Table() string {
	return dao.table
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create create a course progress
func (dao *CourseProgressDao) Create(cp *models.CourseProgress, tx *database.Tx) error {
	execFn := dao.db.Exec
	if tx != nil {
		execFn = tx.Exec
	}

	if cp.ID == "" {
		cp.RefreshId()
	}

	cp.RefreshCreatedAt()
	cp.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.Table()).
		SetMap(dao.data(cp)).
		ToSql()

	_, err := execFn(query, args...)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get gets a course progress with the given course ID
func (dao *CourseProgressDao) Get(courseId string, tx *database.Tx) (*models.CourseProgress, error) {
	dbParams := &database.DatabaseParams{
		Columns: dao.columns(),
		Where:   squirrel.Eq{dao.Table() + ".course_id": courseId},
	}

	return GenericGet(dao, dbParams, dao.scanRow, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Refresh refreshes the current course progress for the given ID
//
// It calculates the number of assets, number of completed assets and number of started video assets,
// then calculates the percent complete and whether the course has been started
//
// Based upon this calculation,
//   - If the course has been started and `started_at` is null, `started_at` will be set to NOW
//   - If the course is not started, `started_at` is set to null
//   - If the course is complete and `completed_at` is null, `completed_at` is set to NOW
//   - If the course is not complete, `completed_at` is set to null
func (dao *CourseProgressDao) Refresh(courseId string, tx *database.Tx) error {
	if courseId == "" {
		return ErrEmptyId
	}

	queryRowFn := dao.db.QueryRow
	execFn := dao.db.Exec

	if tx != nil {
		queryRowFn = tx.QueryRow
		execFn = tx.Exec
	}

	aDao := NewAssetDao(dao.db)
	apDao := NewAssetProgressDao(dao.db)

	// Count the number of assets, number of completed assets and number of video assets started for
	// this course
	query, args, _ := squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select(
			"COUNT(DISTINCT "+aDao.Table()+".id) AS total_count",
			"SUM(CASE WHEN "+apDao.Table()+".completed THEN 1 ELSE 0 END) AS completed_count",
			"SUM(CASE WHEN "+apDao.Table()+".video_pos > 0 THEN 1 ELSE 0 END) AS started_count").
		From(aDao.Table()).
		LeftJoin(apDao.Table() + " ON " + aDao.Table() + ".id = " + apDao.Table() + ".asset_id").
		Where(squirrel.And{squirrel.Eq{aDao.Table() + ".course_id": courseId}}).
		ToSql()

	var totalAssetCount sql.NullInt32
	var completedAssetCount sql.NullInt32
	var startedAssetCount sql.NullInt32
	err := queryRowFn(query, args...).Scan(&totalAssetCount, &completedAssetCount, &startedAssetCount)
	if err != nil {
		return err
	}

	// Default values
	isStarted := false
	startedAt := time.Time{}
	percent := int(math.Abs((float64(completedAssetCount.Int32) * float64(100)) / float64(totalAssetCount.Int32)))
	completedAt := time.Time{}
	updatedAt := time.Now()

	// When there are started assets or percent is between >0 and <=100, set started to true and set started_at
	if startedAssetCount.Int32 > 0 || percent > 0 && percent <= 100 {
		isStarted = true
		startedAt = updatedAt
	}

	// When percent is 100, set completed_at
	if percent == 100 {
		completedAt = startedAt
	}

	builder := squirrel.
		StatementBuilder.
		Update(dao.Table()).
		Set("started", isStarted).
		Set("percent", percent).
		Set("updated_at", FormatTime(updatedAt)).
		Where("course_id = ?", courseId)

	if isStarted {
		builder = builder.Set("started_at", squirrel.Expr("COALESCE(started_at, ?)", FormatTime(startedAt)))
	} else {
		builder = builder.Set("started_at", nil)
	}

	if percent == 100 {
		builder = builder.Set("completed_at", squirrel.Expr("COALESCE(completed_at, ?)", FormatTime(completedAt)))
	} else {
		builder = builder.Set("completed_at", nil)
	}

	query, args, _ = builder.ToSql()

	_, err = execFn(query, args...)
	return err

}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// countSelect returns the default count select builder
func (dao *CourseProgressDao) countSelect() squirrel.SelectBuilder {
	return squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.Table()).
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *CourseProgressDao) baseSelect() squirrel.SelectBuilder {
	return dao.countSelect()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// columns returns the columns to select
func (dao *CourseProgressDao) columns() []string {
	return []string{
		dao.Table() + ".*",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// data generates a map of key/values for a course progress
func (dao *CourseProgressDao) data(cp *models.CourseProgress) map[string]any {
	return map[string]any{
		"id":           cp.ID,
		"course_id":    NilStr(cp.CourseID),
		"started":      cp.Started,
		"started_at":   FormatTime(cp.StartedAt),
		"percent":      cp.Percent,
		"completed_at": FormatTime(cp.CompletedAt),
		"created_at":   FormatTime(cp.CreatedAt),
		"updated_at":   FormatTime(cp.UpdatedAt),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans a courses progress row
func (dao *CourseProgressDao) scanRow(scannable Scannable) (*models.CourseProgress, error) {
	var cp models.CourseProgress

	var createdAt string
	var updatedAt string
	var startedAt sql.NullString
	var completedAt sql.NullString

	err := scannable.Scan(
		&cp.ID,
		&cp.CourseID,
		&cp.Started,
		&startedAt,
		&cp.Percent,
		&completedAt,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	if cp.CreatedAt, err = ParseTime(createdAt); err != nil {
		return nil, err
	}

	if cp.UpdatedAt, err = ParseTime(updatedAt); err != nil {
		return nil, err
	}

	if cp.StartedAt, err = ParseTimeNull(startedAt); err != nil {
		return nil, err
	}

	if cp.CompletedAt, err = ParseTimeNull(completedAt); err != nil {
		return nil, err
	}

	return &cp, nil
}
