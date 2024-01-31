package daos

import (
	"database/sql"
	"math"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseProgressDao is the data access object for courses progress
type CourseProgressDao struct {
	db    database.Database
	table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewCourseProgressDao returns a new CourseProgressDao
func NewCourseProgressDao(db database.Database) *CourseProgressDao {
	return &CourseProgressDao{
		db:    db,
		table: TableCoursesProgress(),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TableCourseProgress returns the name of the courses progress table
func TableCoursesProgress() string {
	return "courses_progress"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create inserts a new course progress
//
// If this is part of a transaction, use `CreateTx`
func (dao *CourseProgressDao) Create(cp *models.CourseProgress) error {
	return dao.create(cp, dao.db.Exec)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateTx inserts a new course progress in a transaction
func (dao *CourseProgressDao) CreateTx(cp *models.CourseProgress, tx *sql.Tx) error {
	return dao.create(cp, tx.Exec)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get selects a course progress with the given course ID
func (dao *CourseProgressDao) Get(courseId string) (*models.CourseProgress, error) {
	return dao.get(courseId, dao.db.QueryRow)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateTx selects a course progress with the given course ID in a transaction
func (dao *CourseProgressDao) GetTx(courseId string, tx *sql.Tx) (*models.CourseProgress, error) {
	return dao.get(courseId, tx.QueryRow)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Refresh refreshes the current course progress for the given course ID
func (dao *CourseProgressDao) Refresh(courseId string) error {
	return dao.db.RunInTransaction(func(tx *sql.Tx) error {
		return dao.refresh(courseId, tx.QueryRow, tx.Exec)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshTx refreshes the current course progress for the given course ID in a transaction
func (dao *CourseProgressDao) RefreshTx(courseId string, tx *sql.Tx) error {
	return dao.refresh(courseId, tx.QueryRow, tx.Exec)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// create inserts a new course progress
//
// NOTE: There is currently no support for users, but when there is, the default courses_progress
// should be inserted for the admin user
func (dao *CourseProgressDao) create(cp *models.CourseProgress, execFunc database.ExecFn) error {
	if cp.ID == "" {
		cp.RefreshId()
	}

	cp.RefreshCreatedAt()
	cp.RefreshUpdatedAt()

	data := map[string]interface{}{
		"id":           cp.ID,
		"course_id":    NilStr(cp.CourseID),
		"started":      cp.Started,
		"started_at":   NilStr(cp.StartedAt.String()),
		"percent":      cp.Percent,
		"completed_at": NilStr(cp.CompletedAt.String()),
		"created_at":   cp.CreatedAt,
		"updated_at":   cp.UpdatedAt,
	}

	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.table).
		SetMap(data).
		ToSql()

	_, err := execFunc(query, args...)

	return err
}

// Get selects a course progress with the given course ID
func (dao *CourseProgressDao) get(courseId string, queryRowFn database.QueryRowFn) (*models.CourseProgress, error) {
	generic := NewGenericDao(dao.db, dao.table)

	dbParams := &database.DatabaseParams{
		Columns: dao.selectColumns(),
		Where:   squirrel.Eq{generic.table + ".course_id": courseId},
	}

	row, err := generic.get(dao.baseSelect(), dbParams, queryRowFn)
	if err != nil {
		return nil, err
	}

	cp, err := dao.scanRow(row)
	if err != nil {
		return nil, err
	}

	return cp, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// refresh does a refresh of the current course progress for the given course ID
//
// It calculates the number of assets, number of completed assets and number of videos started. It
// then calculates the percent complete and whether the course has been started or not.
//
// When the course is started and `started_at` is null, `started_at` is set to the current time.
// When not started, `started_at` is set to null
//
// When percent complete is 100 and `completed_at` is null, `completed_at` is set to the current
// time. When not complete, `completed_at` is set to null
func (dao *CourseProgressDao) refresh(courseId string, queryRowFn database.QueryRowFn, execFn database.ExecFn) error {
	if courseId == "" {
		return ErrEmptyId
	}

	// Count the number of assets, number of completed assets and number of videos started
	query, args, _ := squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select(
			"COUNT(DISTINCT "+TableAssets()+".id) AS total_count",
			"SUM(CASE WHEN "+TableAssetsProgress()+".completed THEN 1 ELSE 0 END) AS completed_count",
			"SUM(CASE WHEN "+TableAssetsProgress()+".video_pos > 0 THEN 1 ELSE 0 END) AS started_count").
		From(TableAssets()).
		LeftJoin(TableAssetsProgress() + " ON " + TableAssets() + ".id = " + TableAssetsProgress() + ".asset_id").
		Where(squirrel.And{squirrel.Eq{TableAssets() + ".course_id": courseId}}).
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
	startedAt := types.DateTime{}
	percent := int(math.Abs((float64(completedAssetCount.Int32) * float64(100)) / float64(totalAssetCount.Int32)))
	completedAt := types.DateTime{}
	updatedAt := types.NowDateTime()

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
		Update(dao.table).
		Set("started", isStarted).
		Set("percent", percent).
		Set("updated_at", updatedAt).
		Where("course_id = ?", courseId)

	if isStarted {
		builder = builder.Set("started_at", squirrel.Expr("COALESCE(started_at, ?)", startedAt))
	} else {
		builder = builder.Set("started_at", nil)
	}

	if percent == 100 {
		builder = builder.Set("completed_at", squirrel.Expr("COALESCE(completed_at, ?)", completedAt))
	} else {
		builder = builder.Set("completed_at", nil)
	}

	query, args, _ = builder.ToSql()

	_, err = execFn(query, args...)
	return err
}

// baseSelect returns the default select builder
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *CourseProgressDao) baseSelect() squirrel.SelectBuilder {
	return squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.table).
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// selectColumns returns the columns to select
func (dao *CourseProgressDao) selectColumns() []string {
	return []string{
		dao.table + ".*",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans a courses progress row
func (dao *CourseProgressDao) scanRow(scannable Scannable) (*models.CourseProgress, error) {
	var cp models.CourseProgress

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
