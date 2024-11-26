package dao

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/stretchr/testify/require"
)

func Test_CreateCourseTag(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 2 {
			course := &models.Course{Title: fmt.Sprintf("Course %d", i), Path: fmt.Sprintf("/course-%d", i)}
			require.NoError(t, dao.CreateCourse(ctx, course))
			courses = append(courses, course)
		}

		tag := &models.Tag{Tag: "Go"}
		require.NoError(t, dao.CreateTag(ctx, tag))

		// Using ID
		courseTagByID := &models.CourseTag{TagID: tag.ID, CourseID: courses[0].ID}
		require.Nil(t, dao.CreateCourseTag(ctx, courseTagByID))

		// Using Tag
		courseTagByTag := &models.CourseTag{CourseID: courses[1].ID, Tag: "Go"}
		require.Nil(t, dao.CreateCourseTag(ctx, courseTagByTag))

		// Create (tag does not exist)
		courseTagCreated := &models.CourseTag{CourseID: courses[1].ID, Tag: "TypeScript"}
		require.Nil(t, dao.CreateCourseTag(ctx, courseTagCreated))
	})

	t.Run("nil pointer", func(t *testing.T) {
		dao, ctx := setup(t)

		require.ErrorIs(t, dao.CreateCourseTag(ctx, nil), utils.ErrNilPtr)
	})

	t.Run("invalid tag ID", func(t *testing.T) {
		dao, ctx := setup(t)

		courseTag := &models.CourseTag{TagID: "invalid", CourseID: "invalid"}
		require.ErrorContains(t, dao.CreateCourseTag(ctx, courseTag), "FOREIGN KEY constraint failed")
	})

	t.Run("invalid course ID", func(t *testing.T) {
		dao, ctx := setup(t)

		tag := &models.Tag{Tag: "Go"}
		require.NoError(t, dao.CreateTag(ctx, tag))

		courseTag := &models.CourseTag{TagID: tag.ID, CourseID: "invalid"}
		require.ErrorContains(t, dao.CreateCourseTag(ctx, courseTag), "FOREIGN KEY constraint failed")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CourseTagDeleteCascade(t *testing.T) {
	dao, ctx := setup(t)

	course := &models.Course{Title: "Course", Path: "/course"}
	require.NoError(t, dao.CreateCourse(ctx, course))

	courseTag := &models.CourseTag{CourseID: course.ID, Tag: "Tag 1"}
	require.NoError(t, dao.CreateCourseTag(ctx, courseTag))

	require.Nil(t, dao.Delete(ctx, course, nil))

	err := dao.GetById(ctx, courseTag)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseTag_PluckForTags(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		dao, ctx := setup(t)

		courseIDs, err := dao.PluckForTags(ctx, []string{"1234"})
		require.NoError(t, err)
		require.Empty(t, courseIDs)
	})

	t.Run("found", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 2 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course %d", i)}
			require.NoError(t, dao.CreateCourse(ctx, course))
			courses = append(courses, course)
		}

		// Add Go and C to course 1
		require.NoError(t, dao.CreateCourseTag(ctx, &models.CourseTag{CourseID: courses[0].ID, Tag: "Go"}))
		require.NoError(t, dao.CreateCourseTag(ctx, &models.CourseTag{CourseID: courses[0].ID, Tag: "C"}))

		// Add Go and JavaScript to course 2
		require.NoError(t, dao.CreateCourseTag(ctx, &models.CourseTag{CourseID: courses[1].ID, Tag: "Go"}))
		require.NoError(t, dao.CreateCourseTag(ctx, &models.CourseTag{CourseID: courses[1].ID, Tag: "JavaScript"}))

		fmt.Println(courses[0].ID, courses[1].ID)

		// Go
		courseIDs, err := dao.PluckForTags(ctx, []string{"Go"})
		require.NoError(t, err)
		require.Len(t, courseIDs, 2)
		require.Equal(t, courses[0].ID, courseIDs[0])
		require.Equal(t, courses[1].ID, courseIDs[1])

		// Go, JavaScript
		courseIDs, err = dao.PluckForTags(ctx, []string{"Go", "JavaScript"})
		require.NoError(t, err)
		require.Len(t, courseIDs, 1)
		require.Equal(t, courses[1].ID, courseIDs[0])

		// Go, JavaScript, C
		courseIDs, err = dao.PluckForTags(ctx, []string{"Go", "JavaScript", "C"})
		require.NoError(t, err)
		require.Len(t, courseIDs, 0)
	})
}
