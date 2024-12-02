package dao

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateScan(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1", Available: true, CardPath: "/course-1/card-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID}
		require.Nil(t, dao.CreateScan(ctx, scan))
	})

	t.Run("nil pointer", func(t *testing.T) {
		dao, ctx := setup(t)

		require.ErrorIs(t, dao.CreateScan(ctx, nil), utils.ErrNilPtr)
	})

	t.Run("invalid course ID", func(t *testing.T) {
		dao, ctx := setup(t)

		scan := &models.Scan{CourseID: "invalid"}
		require.ErrorContains(t, dao.CreateScan(ctx, scan), "FOREIGN KEY constraint failed")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		originalScan := &models.Scan{CourseID: course.ID}
		require.NoError(t, dao.CreateScan(ctx, originalScan))

		newS := &models.Scan{
			Base:     originalScan.Base,
			CourseID: "1234",                          // Immutable
			Status:   types.NewScanStatusProcessing(), // Mutable
		}

		time.Sleep(1 * time.Millisecond)
		require.NoError(t, dao.UpdateScan(ctx, newS))

		scanResult := &models.Scan{Base: models.Base{ID: originalScan.ID}}
		require.Nil(t, dao.GetById(ctx, scanResult))
		require.Equal(t, originalScan.ID, scanResult.ID)                     // No change
		require.Equal(t, originalScan.CourseID, scanResult.CourseID)         // No change
		require.True(t, scanResult.CreatedAt.Equal(originalScan.CreatedAt))  // No change
		require.False(t, scanResult.Status.IsWaiting())                      // Changed
		require.False(t, scanResult.UpdatedAt.Equal(originalScan.UpdatedAt)) // Changed
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID}
		require.NoError(t, dao.CreateScan(ctx, scan))

		// Empty ID
		scan.ID = ""
		require.ErrorIs(t, dao.UpdateScan(ctx, scan), utils.ErrInvalidId)

		// Nil Model
		require.ErrorIs(t, dao.UpdateScan(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_NextWaitingScan(t *testing.T) {
	t.Run("first", func(t *testing.T) {
		dao, ctx := setup(t)

		scans := []*models.Scan{}
		for i := range 3 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, dao.CreateCourse(ctx, course))

			scan := &models.Scan{CourseID: course.ID}
			require.NoError(t, dao.CreateScan(ctx, scan))
			scans = append(scans, scan)

			time.Sleep(1 * time.Millisecond)
		}

		scanResult := &models.Scan{}
		require.NoError(t, dao.NextWaitingScan(ctx, scanResult))
		require.Equal(t, scans[0].ID, scanResult.ID)
	})

	t.Run("next", func(t *testing.T) {
		dao, ctx := setup(t)

		scans := []*models.Scan{}
		for i := range 3 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, dao.CreateCourse(ctx, course))

			scan := &models.Scan{CourseID: course.ID}
			require.NoError(t, dao.CreateScan(ctx, scan))
			scans = append(scans, scan)

			time.Sleep(1 * time.Millisecond)
		}

		scans[0].Status = types.NewScanStatusProcessing()
		require.NoError(t, dao.UpdateScan(ctx, scans[0]))

		scanResult := &models.Scan{}
		require.NoError(t, dao.NextWaitingScan(ctx, scanResult))
		require.Equal(t, scans[1].ID, scanResult.ID)
	})

	t.Run("empty", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.NextWaitingScan(ctx, &models.Scan{}), sql.ErrNoRows)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ScanDeleteCascade(t *testing.T) {
	dao, ctx := setup(t)

	course := &models.Course{Title: "Course", Path: "/course"}
	require.NoError(t, dao.CreateCourse(ctx, course))

	scan := &models.Scan{CourseID: course.ID}
	require.NoError(t, dao.Create(ctx, scan))

	require.Nil(t, dao.Delete(ctx, course, nil))

	err := dao.GetById(ctx, scan)
	require.ErrorIs(t, err, sql.ErrNoRows)
}
