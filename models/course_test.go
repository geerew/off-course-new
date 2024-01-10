package models

import (
	"database/sql"
	"fmt"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CountCourses(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		count, err := CountCourses(db, nil)
		require.Nil(t, err)
		assert.Zero(t, count)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		NewTestCourses(t, db, 5)

		count, err := CountCourses(db, nil)
		require.Nil(t, err)
		assert.Equal(t, count, 5)
	})

	t.Run("where", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 3)

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		count, err := CountCourses(db, &database.DatabaseParams{Where: sq.Eq{TableCourses() + ".id": courses[2].ID}})
		require.Nil(t, err)
		assert.Equal(t, 1, count)

		// ----------------------------
		// NOT EQUALS ID
		// ----------------------------
		count, err = CountCourses(db, &database.DatabaseParams{Where: sq.NotEq{TableCourses() + ".id": courses[2].ID}})
		require.Nil(t, err)
		assert.Equal(t, 2, count)

		// ----------------------------
		// ERROR
		// ----------------------------
		count, err = CountCourses(db, &database.DatabaseParams{Where: sq.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		assert.Equal(t, -1, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableCourses())
		require.Nil(t, err)

		_, err = CountCourses(db, nil)
		require.ErrorContains(t, err, "no such table: "+TableCourses())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetCourses(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		courses, err := GetCourses(db, nil)
		require.Nil(t, err)
		require.Zero(t, courses)
	})

	t.Run("found", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)

		result, err := GetCourses(db, nil)
		require.Nil(t, err)
		require.Len(t, result, 5)

		// ----------------------------
		// Scan
		// ----------------------------
		assert.Empty(t, result[1].ScanStatus)
		NewTestScans(t, db, []*Course{courses[1]})

		result, err = GetCourses(db, nil)
		require.Nil(t, err)
		assert.Equal(t, courses[1].ID, result[1].ID)
		assert.Equal(t, string(types.ScanStatusWaiting), result[1].ScanStatus)

		// ----------------------------
		// Progress
		// ----------------------------
		for _, c := range result {
			require.False(t, c.Started)
			require.True(t, c.StartedAt.IsZero())
			require.Zero(t, c.Percent)
			require.True(t, c.CompletedAt.IsZero())
		}

		// Set course 1 as started and course 3 as completed
		_, err = UpdateCourseProgressStarted(db, courses[0].ID, true)
		require.Nil(t, err)
		_, err = UpdateCourseProgressStarted(db, courses[2].ID, true)
		require.Nil(t, err)
		_, err = UpdateCourseProgressPercent(db, courses[2].ID, 100)
		require.Nil(t, err)

		// Find started courses (not completed)
		dbParams := &database.DatabaseParams{
			Where: sq.And{sq.Eq{TableCoursesProgress() + ".started": true}, sq.NotEq{TableCoursesProgress() + ".percent": 100}},
		}
		result, err = GetCourses(db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, courses[0].ID, result[0].ID)

		// Find completed courses
		result, err = GetCourses(db, &database.DatabaseParams{Where: sq.Eq{TableCoursesProgress() + ".percent": 100}})
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, courses[2].ID, result[0].ID)
	})

	t.Run("orderby", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 3)

		// ----------------------------
		// Descending
		// ----------------------------
		dbParams := &database.DatabaseParams{OrderBy: []string{"created_at desc"}}
		result, err := GetCourses(db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, courses[2].ID, result[0].ID)

		// ----------------------------
		// Ascending
		// ----------------------------
		result, err = GetCourses(db, &database.DatabaseParams{OrderBy: []string{"created_at asc"}})
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, courses[0].ID, result[0].ID)

		// ----------------------------
		// Error
		// ----------------------------
		dbParams = &database.DatabaseParams{OrderBy: []string{"unit_test asc"}}
		result, err = GetCourses(db, dbParams)
		require.ErrorContains(t, err, "no such column")
		assert.Nil(t, result)
	})

	t.Run("where", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 3)

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		result, err := GetCourses(db, &database.DatabaseParams{Where: sq.Eq{TableCourses() + ".id": courses[2].ID}})
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, courses[2].ID, result[0].ID)

		// ----------------------------
		// EQUALS ID OR ID
		// ----------------------------
		dbParams := &database.DatabaseParams{
			Where:   sq.Or{sq.Eq{TableCourses() + ".id": courses[1].ID}, sq.Eq{TableCourses() + ".id": courses[2].ID}},
			OrderBy: []string{"created_at asc"},
		}
		result, err = GetCourses(db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, courses[1].ID, result[0].ID)
		assert.Equal(t, courses[2].ID, result[1].ID)

		// ----------------------------
		// ERROR
		// ----------------------------
		result, err = GetCourses(db, &database.DatabaseParams{Where: sq.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		assert.Nil(t, result)
	})

	t.Run("pagination", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 17)

		// ----------------------------
		// Page 1 with 10 items
		// ----------------------------
		p := pagination.New(1, 10)

		result, err := GetCourses(db, &database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, courses[0].ID, result[0].ID)
		assert.Equal(t, courses[9].ID, result[9].ID)

		// ----------------------------
		// Page 2 with 7 items
		// ----------------------------
		p = pagination.New(2, 10)

		result, err = GetCourses(db, &database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 7)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, courses[10].ID, result[0].ID)
		assert.Equal(t, courses[16].ID, result[6].ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableCourses())
		require.Nil(t, err)

		_, err = GetCourses(db, nil)
		require.ErrorContains(t, err, "no such table: "+TableCourses())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetCourse(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		c, err := GetCourse(db, "1234")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, c)
	})

	t.Run("found", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 2)

		c, err := GetCourse(db, courses[1].ID)
		require.Nil(t, err)
		assert.Equal(t, courses[1].ID, c.ID)
		assert.Empty(t, courses[1].ScanStatus)

		// ----------------------------
		// scan
		// ----------------------------
		NewTestScans(t, db, []*Course{courses[1]})

		c, err = GetCourse(db, courses[1].ID)
		require.Nil(t, err)
		assert.Equal(t, string(types.ScanStatusWaiting), c.ScanStatus)

		// ----------------------------
		// Progress
		// ----------------------------
		require.False(t, c.Started)
		require.True(t, c.StartedAt.IsZero())
		require.Zero(t, c.Percent)
		require.True(t, c.CompletedAt.IsZero())

		// Set to started
		_, err = UpdateCourseProgressStarted(db, courses[1].ID, true)
		require.Nil(t, err)

		c, err = GetCourse(db, courses[1].ID)
		require.Nil(t, err)
		require.True(t, c.Started)
		require.False(t, c.StartedAt.IsZero())
		require.Zero(t, c.Percent)
		require.True(t, c.CompletedAt.IsZero())

		// Set to completed
		_, err = UpdateCourseProgressPercent(db, courses[1].ID, 100)
		require.Nil(t, err)

		c, err = GetCourse(db, courses[1].ID)
		require.Nil(t, err)
		require.True(t, c.Started)
		require.False(t, c.StartedAt.IsZero())
		require.Equal(t, 100, c.Percent)
		require.False(t, c.CompletedAt.IsZero())
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		c, err := GetCourse(db, "")
		require.EqualError(t, err, "id cannot be empty")
		assert.Nil(t, c)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableCourses())
		require.Nil(t, err)

		_, err = GetCourse(db, "1234")
		require.ErrorContains(t, err, "no such table: "+TableCourses())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateCourse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		c := NewTestCourses(t, nil, 1)[0]

		err := CreateCourse(db, c)
		require.Nil(t, err)

		newC, err := GetCourse(db, c.ID)
		require.Nil(t, err)
		assert.NotEmpty(t, newC.ID)
		assert.Equal(t, c.Title, newC.Title)
		assert.Equal(t, c.Path, newC.Path)
		assert.Empty(t, newC.CardPath)
		assert.False(t, newC.CreatedAt.IsZero())
		assert.False(t, newC.UpdatedAt.IsZero())
		//Scan status
		assert.Empty(t, newC.ScanStatus)
		// Progress
		assert.False(t, newC.Started)
		assert.True(t, newC.StartedAt.IsZero())
		assert.Zero(t, newC.Percent)
		assert.True(t, newC.CompletedAt.IsZero())
	})

	t.Run("duplicate paths", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		c := NewTestCourses(t, nil, 1)[0]

		err := CreateCourse(db, c)
		require.Nil(t, err)

		err = CreateCourse(db, c)
		assert.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.path", TableCourses()))
	})

	t.Run("constraint errors", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		// No title
		c := &Course{}
		assert.ErrorContains(t, CreateCourse(db, c), fmt.Sprintf("NOT NULL constraint failed: %s.title", TableCourses()))
		c.Title = ""
		assert.ErrorContains(t, CreateCourse(db, c), fmt.Sprintf("NOT NULL constraint failed: %s.title", TableCourses()))
		c.Title = "Course 1"

		// No path
		assert.ErrorContains(t, CreateCourse(db, c), fmt.Sprintf("NOT NULL constraint failed: %s.path", TableCourses()))
		c.Path = ""
		assert.ErrorContains(t, CreateCourse(db, c), fmt.Sprintf("NOT NULL constraint failed: %s.path", TableCourses()))
		c.Path = "/course 1"

		// Success
		assert.Nil(t, CreateCourse(db, c))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateCourseCardPath(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		origCourse := NewTestCourses(t, db, 1)[0]
		require.Empty(t, origCourse.CardPath)

		updatedCourse, err := UpdateCourseCardPath(db, origCourse.ID, "/path/to/card.jpg")
		require.Nil(t, err)
		require.Equal(t, "/path/to/card.jpg", updatedCourse.CardPath)
		assert.NotEqual(t, origCourse.UpdatedAt, updatedCourse.UpdatedAt)
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		updatedCourse, err := UpdateCourseCardPath(db, "", "")
		assert.EqualError(t, err, "id cannot be empty")
		assert.Nil(t, updatedCourse)
	})

	t.Run("no change", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		origCourse := NewTestCourses(t, db, 1)[0]
		require.Empty(t, origCourse.CardPath)

		updatedCourse, err := UpdateCourseCardPath(db, origCourse.ID, "")
		require.Nil(t, err)
		assert.Empty(t, updatedCourse.CardPath)
		assert.Equal(t, origCourse.UpdatedAt.String(), updatedCourse.UpdatedAt.String())
	})

	t.Run("no course with id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		updatedCourse, err := UpdateCourseCardPath(db, "1234", "/path/to/card.jpg")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, updatedCourse)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableCourses())
		require.Nil(t, err)

		_, err = UpdateCourseCardPath(db, "1234", "/path/to/card.jpg")
		require.ErrorContains(t, err, "no such table: "+TableCourses())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateCourseUpdatedAt(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		origCourse := NewTestCourses(t, db, 1)[0]

		updatedCourse, err := UpdateCourseUpdatedAt(db, origCourse.ID)
		require.Nil(t, err)
		assert.NotEqual(t, origCourse.UpdatedAt, updatedCourse.UpdatedAt)
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		updatedCourse, err := UpdateCourseUpdatedAt(db, "")
		assert.EqualError(t, err, "id cannot be empty")
		assert.Nil(t, updatedCourse)
	})

	t.Run("no course with id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		updatedCourse, err := UpdateCourseUpdatedAt(db, "1234")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, updatedCourse)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableCourses())
		require.Nil(t, err)

		_, err = UpdateCourseUpdatedAt(db, "1234")
		require.ErrorContains(t, err, "no such table: "+TableCourses())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteCourse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]

		err := DeleteCourse(db, course.ID)
		require.Nil(t, err)
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		err := DeleteCourse(db, "")
		assert.ErrorContains(t, err, "id cannot be empty")
	})

	t.Run("invalid id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		err := DeleteCourse(db, "1234")
		assert.Nil(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableCourses())
		require.Nil(t, err)

		err = DeleteCourse(db, "1234")
		require.ErrorContains(t, err, "no such table: "+TableCourses())
	})
}
