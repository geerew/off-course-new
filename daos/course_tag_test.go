package daos

import (
	"fmt"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func courseTagSetup(t *testing.T) (*CourseTagDao, database.Database) {
	_, db := setup(t)
	courseTagDao := NewCourseTagDao(db)
	return courseTagDao, db
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseTag_Count(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		dao, _ := courseTagSetup(t)

		count, err := dao.Count(nil)
		require.Nil(t, err)
		assert.Zero(t, count)
	})

	t.Run("entries", func(t *testing.T) {
		dao, db := courseTagSetup(t)

		NewTestBuilder(t).Db(db).Courses(2).Tags(6).Build()

		count, err := dao.Count(nil)
		require.Nil(t, err)
		assert.Equal(t, count, 12)
	})

	t.Run("where", func(t *testing.T) {
		dao, db := courseTagSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(2).Tags([]string{"C", "Go", "Java", "TypeScript", "JavaScript"}).Build()

		courseDao := NewCourseDao(dao.db)
		tagDao := NewTagDao(dao.db)

		// ----------------------------
		// EQUALS
		// ----------------------------
		count, err := dao.Count(&database.DatabaseParams{Where: squirrel.Eq{courseDao.Table + ".title": testData[1].Course.Title}})
		require.Nil(t, err)
		assert.Equal(t, 5, count)

		// ----------------------------
		// NOT EQUALS
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.NotEq{tagDao.Table + ".tag": "Go"}})
		require.Nil(t, err)
		assert.Equal(t, 8, count)

		// ----------------------------
		//  STARTS WITH (Java%)
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.Like{tagDao.Table + ".tag": "Java%"}})
		require.Nil(t, err)
		assert.Equal(t, 4, count)

		// ----------------------------
		// ERROR
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		assert.Equal(t, 0, count)
	})

	t.Run("db error", func(t *testing.T) {
		dao, db := courseTagSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table)
		require.Nil(t, err)

		_, err = dao.Count(nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseTag_Create(t *testing.T) {
	t.Run("success (new tag)", func(t *testing.T) {
		dao, db := courseTagSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Build()

		ct := &models.CourseTag{
			CourseId: testData[0].Course.ID,
			Tag:      test_tags[0],
		}

		// Create the course-tag. This will also create the tag
		err := dao.Create(ct, nil)
		require.Nil(t, err)
	})

	t.Run("success (existing tag)", func(t *testing.T) {
		dao, db := courseTagSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Build()

		// Create the tag
		tagDao := NewTagDao(db)
		tag := &models.Tag{
			Tag: test_tags[0],
		}
		require.Nil(t, tagDao.Create(tag, nil))

		// Create the course-tag
		ct := &models.CourseTag{
			TagId:    tag.ID,
			CourseId: testData[0].Course.ID,
			Tag:      tag.Tag,
		}

		err := dao.Create(ct, nil)
		require.Nil(t, err)
	})

	t.Run("duplicate", func(t *testing.T) {
		dao, db := courseTagSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Build()

		ct := &models.CourseTag{
			CourseId: testData[0].Course.ID,
			Tag:      test_tags[0],
		}

		// Create the course-tag. This will also create the tag
		require.Nil(t, dao.Create(ct, nil))

		// Create the course-tag (again)
		require.ErrorContains(t, dao.Create(ct, nil), fmt.Sprintf("UNIQUE constraint failed: %s.tag_id, %s.course_id", dao.Table, dao.Table))
	})

	t.Run("constraints", func(t *testing.T) {
		dao, db := courseTagSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Build()
		// tag := "test"

		// Tag
		ct := &models.CourseTag{}
		require.ErrorIs(t, dao.Create(ct, nil), ErrMissingTag)
		ct.Tag = "test"

		// Course ID
		require.ErrorIs(t, dao.Create(ct, nil), ErrMissingCourseId)
		ct.CourseId = "1234"
		require.ErrorContains(t, dao.Create(ct, nil), "constraint failed: FOREIGN KEY constraint failed")
		ct.CourseId = testData[0].Course.ID

		// Success
		assert.Nil(t, dao.Create(ct, nil))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseTag_List(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		dao, _ := courseTagSetup(t)

		tags, err := dao.List(nil, nil)
		require.Nil(t, err)
		require.Zero(t, tags)
	})

	t.Run("found", func(t *testing.T) {
		dao, _ := courseTagSetup(t)

		NewTestBuilder(t).Db(dao.db).Courses(2).Tags(5).Build()

		result, err := dao.List(nil, nil)
		require.Nil(t, err)
		require.Len(t, result, 10)
	})

	t.Run("orderby", func(t *testing.T) {
		dao, _ := courseTagSetup(t)

		NewTestBuilder(t).Db(dao.db).Courses(2).Tags([]string{"PHP", "Go", "Java", "TypeScript", "JavaScript"}).Build()
		tagDao := NewTagDao(dao.db)

		// ----------------------------
		// TAG DESC
		// ----------------------------
		result, err := dao.List(&database.DatabaseParams{OrderBy: []string{tagDao.Table + ".tag desc"}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 10)
		assert.Equal(t, "TypeScript", result[0].Tag)

		// ----------------------------
		// TAG ASC
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{tagDao.Table + ".tag asc"}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 10)
		assert.Equal(t, "Go", result[0].Tag)

		// ----------------------------
		// Error
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{"unit_test asc"}}, nil)
		require.ErrorContains(t, err, "no such column")
		assert.Nil(t, result)
	})

	t.Run("where", func(t *testing.T) {
		dao, _ := courseTagSetup(t)

		testData := NewTestBuilder(t).Db(dao.db).Courses(2).Tags([]string{"PHP", "Go", "Java", "TypeScript", "JavaScript"}).Build()

		courseDao := NewCourseDao(dao.db)
		tagDao := NewTagDao(dao.db)

		// ----------------------------
		// EQUALS (course title)
		// ----------------------------
		result, err := dao.List(&database.DatabaseParams{Where: squirrel.Eq{courseDao.Table + ".title": testData[0].Course.Title}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 5)

		// ----------------------------
		// Like (Java%)
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Like{tagDao.Table + ".tag": "Java%"}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 4)

		// ----------------------------
		// ERROR
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Eq{"": ""}}, nil)
		require.ErrorContains(t, err, "syntax error")
		assert.Nil(t, result)
	})

	t.Run("pagination", func(t *testing.T) {
		dao, _ := courseTagSetup(t)

		NewTestBuilder(t).Db(dao.db).Courses(1).Tags(20).Build()

		// ----------------------------
		// Page 1 with 10 items
		// ----------------------------
		p := pagination.New(1, 10)

		result, err := dao.List(&database.DatabaseParams{Pagination: p, OrderBy: []string{"tag asc"}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 20, p.TotalItems())
		assert.Equal(t, "C", result[0].Tag)

		// ----------------------------
		// Page 2 with 10 items
		// ----------------------------
		p = pagination.New(2, 10)

		result, err = dao.List(&database.DatabaseParams{Pagination: p, OrderBy: []string{"tag asc"}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 20, p.TotalItems())
		assert.Equal(t, "Perl", result[0].Tag)
	})

	t.Run("db error", func(t *testing.T) {
		dao, db := courseTagSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table)
		require.Nil(t, err)

		_, err = dao.List(nil, nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseTag_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, _ := courseTagSetup(t)

		testData := NewTestBuilder(t).Db(dao.db).Courses(1).Tags([]string{"C", "Go", "JavaScript", "Perl"}).Build()

		err := dao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": testData[0].Tags[1].ID}}, nil)
		require.Nil(t, err)

		tags, err := dao.List(nil, nil)
		require.Nil(t, err)
		require.Len(t, tags, 3)
	})

	t.Run("no db params", func(t *testing.T) {
		_, dao, _ := scanSetup(t)

		err := dao.Delete(nil, nil)
		assert.ErrorIs(t, err, ErrMissingWhere)
	})

	t.Run("db error", func(t *testing.T) {
		dao, db := courseTagSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table)
		require.Nil(t, err)

		err = dao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"tag": "1234"}}, nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table)
	})
}
