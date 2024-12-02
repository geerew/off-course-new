package dao

import (
	"database/sql"
	"testing"
	"time"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateAsset(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.CreateAsset(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateAsset(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		originalAsset := &models.Asset{
			CourseID: course.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, originalAsset))

		time.Sleep(1 * time.Millisecond)

		newAsset := &models.Asset{
			Base:    originalAsset.Base,
			Title:   "Asset 2",                            // Mutable
			Prefix:  sql.NullInt16{Int16: 2, Valid: true}, // Mutable
			Chapter: "Chapter 2",                          // Mutable
			Type:    *types.NewAsset("html"),              // Mutable
			Path:    "/course-1/02 asset.html",            // Mutable
			Hash:    "5678",                               // Mutable
		}
		require.NoError(t, dao.UpdateAsset(ctx, newAsset))

		assertResult := &models.Asset{Base: models.Base{ID: originalAsset.ID}}
		require.NoError(t, dao.GetById(ctx, assertResult))
		require.Equal(t, newAsset.ID, assertResult.ID)                          // No change
		require.True(t, newAsset.CreatedAt.Equal(originalAsset.CreatedAt))      // No change
		require.Equal(t, newAsset.Title, assertResult.Title)                    // Changed
		require.Equal(t, newAsset.Prefix, assertResult.Prefix)                  // Changed
		require.Equal(t, newAsset.Chapter, assertResult.Chapter)                // Changed
		require.Equal(t, newAsset.Type, assertResult.Type)                      // Changed
		require.Equal(t, newAsset.Path, assertResult.Path)                      // Changed
		require.Equal(t, newAsset.Hash, assertResult.Hash)                      // Changed
		require.False(t, assertResult.UpdatedAt.Equal(originalAsset.UpdatedAt)) // Changed
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		// Empty ID
		asset.ID = ""
		require.ErrorIs(t, dao.UpdateAsset(ctx, asset), utils.ErrInvalidId)

		// Nil Model
		require.ErrorIs(t, dao.UpdateAsset(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_AssetDeleteCascade(t *testing.T) {
	dao, ctx := setup(t)

	course := &models.Course{Title: "Course", Path: "/course"}
	require.NoError(t, dao.CreateCourse(ctx, course))

	asset := &models.Asset{
		CourseID: course.ID,
		Title:    "Asset 1",
		Prefix:   sql.NullInt16{Int16: 1, Valid: true},
		Chapter:  "Chapter 1",
		Type:     *types.NewAsset("mp4"),
		Path:     "/course-1/01 asset.mp4",
		Hash:     "1234",
	}
	require.NoError(t, dao.CreateAsset(ctx, asset))

	require.Nil(t, dao.Delete(ctx, course, nil))

	err := dao.GetById(ctx, asset)
	require.ErrorIs(t, err, sql.ErrNoRows)
}
