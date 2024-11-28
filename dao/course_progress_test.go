package dao

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

func Test_CreateCourseProgress(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)
		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.Create(ctx, course))

		courseProgress := &models.CourseProgress{CourseID: course.ID}
		require.NoError(t, dao.CreateCourseProgress(ctx, courseProgress))
	})

	t.Run("nil pointer", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.CreateCourseProgress(ctx, nil), utils.ErrNilPtr)
	})

	t.Run("invalid course id", func(t *testing.T) {
		dao, ctx := setup(t)
		courseProgress := &models.CourseProgress{CourseID: "invalid"}
		require.ErrorContains(t, dao.CreateCourseProgress(ctx, courseProgress), "FOREIGN KEY constraint failed")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_RefreshCourseProgress(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))
		require.False(t, course.Progress.Started)
		require.True(t, course.Progress.StartedAt.IsZero())
		require.Zero(t, course.Progress.Percent)
		require.True(t, course.Progress.CompletedAt.IsZero())

		// Create asset
		asset1 := &models.Asset{
			CourseID: course.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset1))

		// Set asset1 progress (video_pos > 0)
		assetProgress := &models.AssetProgress{AssetID: asset1.ID, VideoPos: 1}
		require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, assetProgress))

		require.NoError(t, dao.GetById(ctx, course))
		require.True(t, course.Progress.Started)
		require.False(t, course.Progress.StartedAt.IsZero())
		require.Zero(t, 0, course.Progress.Percent)
		require.True(t, course.Progress.CompletedAt.IsZero())

		// Set asset progress (video_pos = 0)
		assetProgress.VideoPos = 0
		require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, assetProgress))

		require.NoError(t, dao.GetById(ctx, course))
		require.False(t, course.Progress.Started)
		require.True(t, course.Progress.StartedAt.IsZero())
		require.Zero(t, 0, course.Progress.Percent)
		require.True(t, course.Progress.CompletedAt.IsZero())

		// Set asset progress (completed = true)
		assetProgress.Completed = true
		require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, assetProgress))

		require.NoError(t, dao.GetById(ctx, course))
		require.True(t, course.Progress.Started)
		require.False(t, course.Progress.StartedAt.IsZero())
		require.Equal(t, 100, course.Progress.Percent)
		require.False(t, course.Progress.CompletedAt.IsZero())

		// Add another asset
		asset2 := &models.Asset{
			CourseID: course.ID,
			Title:    "Asset 2",
			Prefix:   sql.NullInt16{Int16: 2, Valid: true},
			Chapter:  "Chapter 2",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/02 asset.mp4",
			Hash:     "5678",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset2))

		// Check course progress
		require.NoError(t, dao.GetById(ctx, course))
		require.True(t, course.Progress.Started)
		require.False(t, course.Progress.StartedAt.IsZero())
		require.Equal(t, 50, course.Progress.Percent)
		require.True(t, course.Progress.CompletedAt.IsZero())
	})

	t.Run("invalid course id", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.RefreshCourseProgress(ctx, ""), utils.ErrInvalidId)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Pluck(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course %d", i)}
			require.NoError(t, dao.CreateCourse(ctx, course))
			courses = append(courses, course)
		}

		assets := []*models.Asset{}
		for i, c := range courses {
			asset := &models.Asset{
				CourseID: c.ID,
				Title:    "asset 1",
				Prefix:   sql.NullInt16{Int16: 1, Valid: true},
				Chapter:  "Chapter 1",
				Type:     *types.NewAsset("mp4"),
				Path:     fmt.Sprintf("/course %d/chapter 1/01 asset 1.mp4", i+1),
				Hash:     security.RandomString(64),
			}

			require.NoError(t, dao.CreateAsset(ctx, asset))
			assets = append(assets, asset)
		}

		// Mark course 1 as started
		require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, &models.AssetProgress{AssetID: assets[0].ID, VideoPos: 10}))

		// Mark course 2 as completed
		require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, &models.AssetProgress{AssetID: assets[1].ID, VideoPos: 10, Completed: true}))

		// Find started courses
		ids, err := dao.PluckStartedCourses(ctx)
		require.NoError(t, err)
		require.Len(t, ids, 1)
		require.Equal(t, courses[0].ID, ids[0])

		// Find completed courses
		ids, err = dao.PluckCompletedCourses(ctx)
		require.NoError(t, err)
		require.Len(t, ids, 1)
		require.Equal(t, courses[1].ID, ids[0])

		// Find not started courses
		ids, err = dao.PluckNotStartedCourses(ctx)
		require.NoError(t, err)
		require.Len(t, ids, 1)
		require.Equal(t, courses[2].ID, ids[0])
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_PluckCompletedCourses(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 2 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course %d", i)}
			require.NoError(t, dao.CreateCourse(ctx, course))
			courses = append(courses, course)
		}

		assets := []*models.Asset{}
		for i, c := range courses {
			asset := &models.Asset{
				CourseID: c.ID,
				Title:    "asset 1",
				Prefix:   sql.NullInt16{Int16: 1, Valid: true},
				Chapter:  "Chapter 1",
				Type:     *types.NewAsset("mp4"),
				Path:     fmt.Sprintf("/course %d/chapter 1/01 asset 1.mp4", i+1),
				Hash:     security.RandomString(64),
			}

			require.NoError(t, dao.CreateAsset(ctx, asset))
			assets = append(assets, asset)
		}

		// Mark course 1 as started
		assetProgress := &models.AssetProgress{AssetID: assets[0].ID, VideoPos: 10, Completed: true}
		require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, assetProgress))

		ids, err := dao.PluckCompletedCourses(ctx)
		require.NoError(t, err)
		require.Len(t, ids, 1)
		require.Equal(t, courses[0].ID, ids[0])
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CourseProgressDeleteCascade(t *testing.T) {
	dao, ctx := setup(t)

	course := &models.Course{Title: "Course", Path: "/course"}
	require.NoError(t, dao.CreateCourse(ctx, course))

	courseProgress := &models.CourseProgress{CourseID: course.ID}
	require.NoError(t, dao.Get(ctx, courseProgress, &database.Options{Where: squirrel.Eq{courseProgress.Table() + ".course_id": course.ID}}))

	require.NoError(t, dao.Delete(ctx, courseProgress, nil))

	err := dao.Get(ctx, courseProgress, &database.Options{Where: squirrel.Eq{courseProgress.Table() + ".course_id": course.ID}})
	require.ErrorIs(t, err, sql.ErrNoRows)
}
