package dao

import (
	"context"
	"database/sql"
	"math"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateCourseProgress creates a course progress
func (dao *DAO) CreateCourseProgress(ctx context.Context, courseProgress *models.CourseProgress) error {
	if courseProgress == nil {
		return utils.ErrNilPtr
	}

	return dao.Create(ctx, courseProgress)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateCourseProgress update a course progress
func (dao *DAO) UpdateCourseProgress(ctx context.Context, courseProgress *models.CourseProgress) error {
	if courseProgress == nil {
		return utils.ErrNilPtr
	}

	_, err := dao.Update(ctx, courseProgress)
	return err
}

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
func (dao *DAO) RefreshCourseProgress(ctx context.Context, courseID string) error {
	if courseID == "" {
		return utils.ErrInvalidId
	}

	// Count the number of assets, number of completed assets and number of video assets started for
	// this course
	query, args, _ := squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select(
			"COUNT(DISTINCT "+models.ASSET_TABLE+".id) AS total_count",
			"SUM(CASE WHEN "+models.ASSET_PROGRESS_TABLE+".completed THEN 1 ELSE 0 END) AS completed_count",
			"SUM(CASE WHEN "+models.ASSET_PROGRESS_TABLE+".video_pos > 0 THEN 1 ELSE 0 END) AS started_count").
		From(models.ASSET_TABLE).
		LeftJoin(models.ASSET_PROGRESS_TABLE + " ON " + models.ASSET_TABLE + ".id = " + models.ASSET_PROGRESS_TABLE + ".asset_id").
		Where(squirrel.And{squirrel.Eq{models.ASSET_TABLE + ".course_id": courseID}}).
		ToSql()

	var totalAssetCount sql.NullInt32
	var completedAssetCount sql.NullInt32
	var startedAssetCount sql.NullInt32

	q := database.QuerierFromContext(ctx, dao.db)
	err := q.QueryRow(query, args...).Scan(&totalAssetCount, &completedAssetCount, &startedAssetCount)
	if err != nil {
		return err
	}

	// Get the course progress
	courseProgress := &models.CourseProgress{}
	err = dao.Get(ctx, courseProgress, &database.Options{Where: squirrel.Eq{courseProgress.Table() + ".course_id": courseID}})
	if err != nil {
		return err
	}

	now := types.NowDateTime()

	courseProgress.Percent = int(math.Abs((float64(completedAssetCount.Int32) * float64(100)) / float64(totalAssetCount.Int32)))

	if startedAssetCount.Int32 > 0 || courseProgress.Percent > 0 && courseProgress.Percent <= 100 {
		courseProgress.Started = true
		courseProgress.StartedAt = now
	} else {
		courseProgress.Started = false
		courseProgress.StartedAt = types.DateTime{}
	}

	if courseProgress.Percent == 100 {
		courseProgress.CompletedAt = now
	} else {
		courseProgress.CompletedAt = types.DateTime{}
	}

	// Update the course progress
	_, err = dao.Update(ctx, courseProgress)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PluckIDsForStartedCourses returns a list of course IDs for courses that have been started but not
// completed
func (dao *DAO) PluckIDsForStartedCourses(ctx context.Context, options *database.Options) ([]string, error) {
	if options == nil {
		options = &database.Options{}
	}

	options.Where = squirrel.And{
		squirrel.Eq{models.COURSE_PROGRESS_TABLE + ".started": true},
		squirrel.NotEq{models.COURSE_PROGRESS_TABLE + ".percent": 100},
	}

	return dao.ListPluck(ctx, &models.CourseProgress{}, options, models.COURSE_PROGRESS_COURSE_ID)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PluckIDsForCompletedCourses returns a list of course IDs for courses that have been completed
func (dao *DAO) PluckIDsForCompletedCourses(ctx context.Context, options *database.Options) ([]string, error) {
	if options == nil {
		options = &database.Options{}
	}

	options.Where = squirrel.Eq{models.COURSE_PROGRESS_TABLE + ".percent": 100}
	return dao.ListPluck(ctx, &models.CourseProgress{}, options, models.COURSE_PROGRESS_COURSE_ID)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PluckNotStartedCourses returns a list of course IDs for courses that have not been started
func (dao *DAO) PluckIDsForNotStartedCourses(ctx context.Context, options *database.Options) ([]string, error) {
	if options == nil {
		options = &database.Options{}
	}

	options.Where = squirrel.Eq{models.COURSE_PROGRESS_TABLE + ".started": false}
	return dao.ListPluck(ctx, &models.CourseProgress{}, options, models.COURSE_PROGRESS_COURSE_ID)
}
