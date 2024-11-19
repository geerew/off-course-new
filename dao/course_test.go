package dao

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateCourse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)
		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.CreateCourse(ctx, nil), utils.ErrNilPtr)
	})

	t.Run("duplicate", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Base: models.Base{ID: "1"}, Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		// Duplicate ID
		course = &models.Course{Base: models.Base{ID: "1"}, Title: "Course 2", Path: "/course-2"}
		require.ErrorContains(t, dao.CreateCourse(ctx, course), "UNIQUE constraint failed: "+models.COURSE_TABLE+".id")

		// Duplicate Path
		course = &models.Course{Base: models.Base{ID: "2"}, Title: "Course 2", Path: "/course-1"}
		require.ErrorContains(t, dao.CreateCourse(ctx, course), "UNIQUE constraint failed: "+models.COURSE_TABLE+".path")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateCourse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		originalCourse := &models.Course{Title: "Course 1", Path: "/course-1", Available: true, CardPath: "/course-1/card-1"}
		require.NoError(t, dao.CreateCourse(ctx, originalCourse))

		time.Sleep(1 * time.Millisecond)

		newCourse := &models.Course{
			Base:      originalCourse.Base,
			Title:     "Course 2",         // Immutable
			Path:      "/course-2",        // Immutable
			Available: false,              // Mutable
			CardPath:  "/course-2/card-2", // Mutable
		}
		require.NoError(t, dao.UpdateCourse(ctx, newCourse))

		courseResult := &models.Course{Base: models.Base{ID: originalCourse.ID}}
		require.NoError(t, dao.GetById(ctx, courseResult))
		require.Equal(t, originalCourse.ID, courseResult.ID)                     // No change
		require.Equal(t, originalCourse.Title, courseResult.Title)               // No change
		require.Equal(t, originalCourse.Path, courseResult.Path)                 // No change
		require.True(t, courseResult.CreatedAt.Equal(originalCourse.CreatedAt))  // No change
		require.False(t, courseResult.Available)                                 // Changed
		require.Equal(t, newCourse.CardPath, courseResult.CardPath)              // Changed
		require.False(t, courseResult.UpdatedAt.Equal(originalCourse.UpdatedAt)) // Changed
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		// Empty ID
		course.ID = ""
		require.ErrorIs(t, dao.UpdateCourse(ctx, course), utils.ErrInvalidId)

		// Nil Model
		require.ErrorIs(t, dao.UpdateCourse(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ClassifyCoursePaths(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			c := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, dao.CreateCourse(ctx, c))
			courses = append(courses, c)
		}

		path1 := string(filepath.Separator)              // ancestor
		path2 := string(filepath.Separator) + "test"     // none
		path3 := courses[2].Path                         // course
		path4 := filepath.Join(courses[2].Path + "test") // descendant

		result, err := dao.ClassifyCoursePaths(ctx, []string{path1, path2, path3, path4})
		require.Nil(t, err)

		require.Equal(t, types.PathClassificationAncestor, result[path1])
		require.Equal(t, types.PathClassificationNone, result[path2])
		require.Equal(t, types.PathClassificationCourse, result[path3])
		require.Equal(t, types.PathClassificationDescendant, result[path4])
	})

	t.Run("no paths", func(t *testing.T) {
		dao, ctx := setup(t)

		result, err := dao.ClassifyCoursePaths(ctx, []string{})
		require.Nil(t, err)
		require.Empty(t, result)
	})

	t.Run("empty path", func(t *testing.T) {
		dao, ctx := setup(t)

		result, err := dao.ClassifyCoursePaths(ctx, []string{"", "", ""})
		require.Nil(t, err)
		require.Empty(t, result)
	})

	t.Run("db error", func(t *testing.T) {
		dao, ctx := setup(t)

		_, err := dao.db.Exec("DROP TABLE IF EXISTS " + (&models.Course{}).Table())
		require.Nil(t, err)

		result, err := dao.ClassifyCoursePaths(ctx, []string{"/"})
		require.ErrorContains(t, err, "no such table: "+(&models.Course{}).Table())
		require.Empty(t, result)
	})
}
