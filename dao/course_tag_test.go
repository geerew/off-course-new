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

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestCourseTag_ListCourseIdsByTags(t *testing.T) {
// 	t.Run("no entries", func(t *testing.T) {
// 		dao, _ := courseTagSetup(t)

// 		tags, err := dao.ListCourseIdsByTags([]string{"1234"}, nil, nil)
// 		require.Nil(t, err)
// 		require.Zero(t, tags)
// 	})

// 	t.Run("found", func(t *testing.T) {
// 		dao, _ := courseTagSetup(t)

// 		course1 := NewTestBuilder(t).Db(dao.db).Courses([]string{"course 1"}).Tags([]string{"Go", "Data Structures"}).Build()[0]
// 		course2 := NewTestBuilder(t).Db(dao.db).Courses([]string{"course 2"}).Tags([]string{"Data Structures", "TypeScript", "PHP"}).Build()[0]
// 		course3 := NewTestBuilder(t).Db(dao.db).Courses([]string{"course 3"}).Tags([]string{"Go", "Data Structures", "PHP"}).Build()[0]

// 		// Order by title (asc)
// 		dbParams := &database.DatabaseParams{OrderBy: []string{NewCourseDao(dao.db).Table() + ".title asc"}}

// 		// Go
// 		result, err := dao.ListCourseIdsByTags([]string{"Go"}, dbParams, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 2)
// 		require.Equal(t, course1.ID, result[0])
// 		require.Equal(t, course3.ID, result[1])

// 		// Go, Data Structures
// 		result, err = dao.ListCourseIdsByTags([]string{"Go", "Data Structures"}, dbParams, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 2)
// 		require.Equal(t, course1.ID, result[0])
// 		require.Equal(t, course3.ID, result[1])

// 		// Go, Data Structures, PHP
// 		result, err = dao.ListCourseIdsByTags([]string{"Go", "Data Structures", "PHP"}, dbParams, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 1)
// 		require.Equal(t, course3.ID, result[0])

// 		// Go, Data Structures, PHP, TypeScript
// 		result, err = dao.ListCourseIdsByTags([]string{"Go", "Data Structures", "PHP", "TypeScript"}, dbParams, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 0)

// 		// Data Structures
// 		result, err = dao.ListCourseIdsByTags([]string{"Data Structures"}, dbParams, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 3)
// 		require.Equal(t, course1.ID, result[0])
// 		require.Equal(t, course2.ID, result[1])
// 		require.Equal(t, course3.ID, result[2])

// 	})

// 	t.Run("db error", func(t *testing.T) {
// 		dao, db := courseTagSetup(t)

// 		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
// 		require.Nil(t, err)

// 		_, err = dao.ListCourseIdsByTags([]string{"1234"}, nil, nil)
// 		require.ErrorContains(t, err, "no such table: "+dao.Table())
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestCourseTag_List(t *testing.T) {
// 	t.Run("no entries", func(t *testing.T) {
// 		dao, _ := courseTagSetup(t)

// 		tags, err := dao.List(nil, nil)
// 		require.Nil(t, err)
// 		require.Zero(t, tags)
// 	})

// 	t.Run("found", func(t *testing.T) {
// 		dao, _ := courseTagSetup(t)

// 		NewTestBuilder(t).Db(dao.db).Courses(2).Tags(5).Build()

// 		result, err := dao.List(nil, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 10)
// 	})

// 	t.Run("orderby", func(t *testing.T) {
// 		dao, _ := courseTagSetup(t)

// 		NewTestBuilder(t).Db(dao.db).Courses(2).Tags([]string{"PHP", "Go", "Java", "TypeScript", "JavaScript"}).Build()
// 		tagDao := NewTagDao(dao.db)

// 		// ----------------------------
// 		// TAG DESC
// 		// ----------------------------
// 		result, err := dao.List(&database.DatabaseParams{OrderBy: []string{tagDao.Table() + ".tag desc"}}, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 10)
// 		require.Equal(t, "TypeScript", result[0].Tag)

// 		// ----------------------------
// 		// TAG ASC
// 		// ----------------------------
// 		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{tagDao.Table() + ".tag asc"}}, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 10)
// 		require.Equal(t, "Go", result[0].Tag)
// 	})

// 	t.Run("where", func(t *testing.T) {
// 		dao, _ := courseTagSetup(t)

// 		testData := NewTestBuilder(t).Db(dao.db).Courses(2).Tags([]string{"PHP", "Go", "Java", "TypeScript", "JavaScript"}).Build()

// 		courseDao := NewCourseDao(dao.db)
// 		tagDao := NewTagDao(dao.db)

// 		// ----------------------------
// 		// EQUALS (course title)
// 		// ----------------------------
// 		result, err := dao.List(&database.DatabaseParams{Where: squirrel.Eq{courseDao.Table() + ".title": testData[0].Course.Title}}, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 5)

// 		// ----------------------------
// 		// Like (Java%)
// 		// ----------------------------
// 		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Like{tagDao.Table() + ".tag": "Java%"}}, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 4)

// 		// ----------------------------
// 		// ERROR
// 		// ----------------------------
// 		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Eq{"": ""}}, nil)
// 		require.ErrorContains(t, err, "syntax error")
// 		require.Nil(t, result)
// 	})

// 	t.Run("pagination", func(t *testing.T) {
// 		dao, _ := courseTagSetup(t)

// 		NewTestBuilder(t).Db(dao.db).Courses(1).Tags(20).Build()

// 		// ----------------------------
// 		// Page 1 with 10 items
// 		// ----------------------------
// 		p := pagination.New(1, 10)

// 		result, err := dao.List(&database.DatabaseParams{Pagination: p, OrderBy: []string{"tags.tag asc"}}, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 10)
// 		require.Equal(t, 20, p.TotalItems())
// 		require.Equal(t, "C", result[0].Tag)

// 		// ----------------------------
// 		// Page 2 with 10 items
// 		// ----------------------------
// 		p = pagination.New(2, 10)

// 		result, err = dao.List(&database.DatabaseParams{Pagination: p, OrderBy: []string{"tags.tag asc"}}, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 10)
// 		require.Equal(t, 20, p.TotalItems())
// 		require.Equal(t, "Perl", result[0].Tag)
// 	})

// 	t.Run("db error", func(t *testing.T) {
// 		dao, db := courseTagSetup(t)

// 		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
// 		require.Nil(t, err)

// 		_, err = dao.List(nil, nil)
// 		require.ErrorContains(t, err, "no such table: "+dao.Table())
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestCourseTag_Delete(t *testing.T) {
// 	t.Run("success", func(t *testing.T) {
// 		dao, _ := courseTagSetup(t)

// 		testData := NewTestBuilder(t).Db(dao.db).Courses(1).Tags([]string{"C", "Go", "JavaScript", "Perl"}).Build()

// 		err := dao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": testData[0].Tags[1].ID}}, nil)
// 		require.Nil(t, err)

// 		tags, err := dao.List(nil, nil)
// 		require.Nil(t, err)
// 		require.Len(t, tags, 3)
// 	})

// 	t.Run("no db params", func(t *testing.T) {
// 		dao, _ := courseTagSetup(t)

// 		err := dao.Delete(nil, nil)
// 		require.ErrorIs(t, err, ErrMissingWhere)
// 	})

// 	t.Run("db error", func(t *testing.T) {
// 		dao, db := courseTagSetup(t)

// 		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
// 		require.Nil(t, err)

// 		err = dao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"tag": "1234"}}, nil)
// 		require.ErrorContains(t, err, "no such table: "+dao.Table())
// 	})
// }
