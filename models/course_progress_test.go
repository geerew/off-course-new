package models

import (
	"database/sql"
	"fmt"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CountCoursesProgress(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		count, err := CountCoursesProgress(db, nil)
		require.Nil(t, err)
		assert.Zero(t, count)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		NewTestData(t, db, 5, false, 0, 0)

		count, err := CountCoursesProgress(db, nil)
		require.Nil(t, err)
		assert.Equal(t, count, 5)
	})

	t.Run("where", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		NewTestData(t, db, 3, false, 0, 0)

		// Get the courses progress
		cps, err := GetCoursesProgress(db, nil)
		require.Nil(t, err)

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		count, err := CountCoursesProgress(db, &database.DatabaseParams{Where: sq.Eq{TableCoursesProgress() + ".id": cps[2].ID}})
		require.Nil(t, err)
		assert.Equal(t, 1, count)

		// ----------------------------
		// NOT EQUALS ID
		// ----------------------------
		count, err = CountCoursesProgress(db, &database.DatabaseParams{Where: sq.NotEq{TableCoursesProgress() + ".id": cps[2].ID}})
		require.Nil(t, err)
		assert.Equal(t, 2, count)

		// ----------------------------
		// ERROR
		// ----------------------------
		count, err = CountCoursesProgress(db, &database.DatabaseParams{Where: sq.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		assert.Equal(t, -1, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableCoursesProgress())
		require.Nil(t, err)

		_, err = CountCoursesProgress(db, nil)
		require.ErrorContains(t, err, "no such table: "+TableCoursesProgress())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetCoursesProgress(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		courses, err := GetCoursesProgress(db, nil)
		require.Nil(t, err)
		assert.Zero(t, courses)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		NewTestData(t, db, 5, false, 0, 0)

		result, err := GetCoursesProgress(db, nil)
		require.Nil(t, err)
		assert.Len(t, result, 5)
	})

	t.Run("orderby", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		NewTestData(t, db, 3, false, 0, 0)

		// Get the courses progress
		cps, err := GetCoursesProgress(db, nil)
		require.Nil(t, err)

		// ----------------------------
		// Descending
		// ----------------------------
		dbParams := &database.DatabaseParams{OrderBy: []string{"created_at desc"}}
		result, err := GetCoursesProgress(db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, cps[2].ID, result[0].ID)

		// ----------------------------
		// Ascending
		// ----------------------------
		result, err = GetCoursesProgress(db, &database.DatabaseParams{OrderBy: []string{"created_at asc"}})
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, cps[0].ID, result[0].ID)

		// ----------------------------
		// Error
		// ----------------------------
		dbParams = &database.DatabaseParams{OrderBy: []string{"unit_test asc"}}
		result, err = GetCoursesProgress(db, dbParams)
		require.ErrorContains(t, err, "no such column")
		assert.Nil(t, result)
	})

	t.Run("where", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		NewTestData(t, db, 3, false, 0, 0)

		// Get the courses progress
		cps, err := GetCoursesProgress(db, nil)
		require.Nil(t, err)

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		result, err := GetCoursesProgress(db, &database.DatabaseParams{Where: sq.Eq{TableCoursesProgress() + ".id": cps[2].ID}})
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, cps[2].ID, result[0].ID)

		// ----------------------------
		// EQUALS ID OR ID
		// ----------------------------
		dbParams := &database.DatabaseParams{
			Where:   sq.Or{sq.Eq{TableCoursesProgress() + ".id": cps[1].ID}, sq.Eq{TableCoursesProgress() + ".id": cps[2].ID}},
			OrderBy: []string{"created_at asc"},
		}
		result, err = GetCoursesProgress(db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, cps[1].ID, result[0].ID)
		assert.Equal(t, cps[2].ID, result[1].ID)

		// ----------------------------
		// ERROR
		// ----------------------------
		result, err = GetCoursesProgress(db, &database.DatabaseParams{Where: sq.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		assert.Nil(t, result)
	})

	t.Run("pagination", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		NewTestData(t, db, 17, false, 0, 0)

		// Get the courses progress
		cps, err := GetCoursesProgress(db, nil)
		require.Nil(t, err)

		// ----------------------------
		// Page 1 with 10 items
		// ----------------------------
		p := pagination.New(1, 10)

		result, err := GetCoursesProgress(db, &database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, cps[0].ID, result[0].ID)
		assert.Equal(t, cps[9].ID, result[9].ID)

		// ----------------------------
		// Page 2 with 7 items
		// ----------------------------
		p = pagination.New(2, 10)

		result, err = GetCoursesProgress(db, &database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 7)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, cps[10].ID, result[0].ID)
		assert.Equal(t, cps[16].ID, result[6].ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableCoursesProgress())
		require.Nil(t, err)

		_, err = GetCoursesProgress(db, nil)
		require.ErrorContains(t, err, "no such table: "+TableCoursesProgress())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetCourseProgress(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		cp, err := GetCourseProgress(db, "1234")
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, cp)
	})

	t.Run("found", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 2, false, 0, 0)

		result, err := GetCourseProgress(db, workingData[1].ID)
		require.Nil(t, err)
		assert.Equal(t, workingData[1].ID, result.CourseID)
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		result, err := GetCourseProgress(db, "")
		require.EqualError(t, err, "id cannot be empty")
		assert.Nil(t, result)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableCoursesProgress())
		require.Nil(t, err)

		_, err = GetCourseProgress(db, "1234")
		require.ErrorContains(t, err, "no such table: "+TableCoursesProgress())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateCourseProgress(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		// The courses progress is created when the course is created
		workingData := NewTestData(t, db, 1, false, 0, 0)

		cp, err := GetCourseProgress(db, workingData[0].ID)
		require.Nil(t, err)

		require.Nil(t, err)
		require.NotEmpty(t, cp.ID)
		assert.Equal(t, workingData[0].ID, cp.CourseID)
		assert.False(t, cp.Started)
		assert.True(t, cp.StartedAt.IsZero())
		assert.Zero(t, cp.Percent)
		assert.True(t, cp.CompletedAt.IsZero())
		assert.False(t, cp.CreatedAt.IsZero())
		assert.False(t, cp.UpdatedAt.IsZero())
	})

	t.Run("duplicate course id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		// The courses progress is created when the course is created
		workingData := NewTestData(t, db, 1, false, 0, 0)

		cp, err := GetCourseProgress(db, workingData[0].ID)
		require.Nil(t, err)

		err = CreateCourseProgress(db, cp)
		require.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.course_id", TableCoursesProgress()))
	})

	t.Run("constraint errors", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)

		// Delete the courses_progress row using squirrel
		query, args, err := sq.StatementBuilder.Delete(TableCoursesProgress()).Where(sq.Eq{"course_id": workingData[0].ID}).ToSql()
		require.Nil(t, err)

		_, err = db.Exec(query, args...)
		require.Nil(t, err)

		// Course ID
		cp := &CourseProgress{}
		require.ErrorContains(t, CreateCourseProgress(db, cp), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", TableCoursesProgress()))
		cp.CourseID = ""
		require.ErrorContains(t, CreateCourseProgress(db, cp), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", TableCoursesProgress()))
		cp.CourseID = "1234"

		// Invalid Course ID
		require.ErrorContains(t, CreateCourseProgress(db, cp), "FOREIGN KEY constraint failed")
		cp.CourseID = workingData[0].ID

		// Success
		require.Nil(t, CreateCourseProgress(db, cp))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateCourseProgressStarted(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)

		origCp, err := GetCourseProgress(db, workingData[0].ID)
		require.Nil(t, err)
		require.False(t, origCp.Started)
		assert.True(t, origCp.StartedAt.IsZero())

		// ----------------------------
		// Set to true
		// ----------------------------
		updatedCp1, err := UpdateCourseProgressStarted(db, origCp.CourseID, true)
		require.Nil(t, err)
		require.True(t, updatedCp1.Started)
		assert.False(t, updatedCp1.StartedAt.IsZero())
		assert.NotEqual(t, origCp.UpdatedAt, updatedCp1.UpdatedAt)

		// ----------------------------
		// Set to false
		// ----------------------------
		updatedCp2, err := UpdateCourseProgressStarted(db, origCp.CourseID, false)
		require.Nil(t, err)
		require.False(t, updatedCp2.Started)
		assert.True(t, updatedCp2.StartedAt.IsZero())
		assert.NotEqual(t, updatedCp1.UpdatedAt, updatedCp2.UpdatedAt)
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		cp, err := UpdateCourseProgressStarted(db, "", true)
		assert.EqualError(t, err, "id cannot be empty")
		assert.Nil(t, cp)
	})

	t.Run("no change", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)

		origCp, err := GetCourseProgress(db, workingData[0].ID)
		require.Nil(t, err)
		require.False(t, origCp.Started)
		assert.True(t, origCp.StartedAt.IsZero())

		updatedCp, err := UpdateCourseProgressStarted(db, origCp.CourseID, false)
		require.Nil(t, err)
		assert.False(t, updatedCp.Started)
		assert.True(t, updatedCp.StartedAt.IsZero())
		assert.Equal(t, origCp.UpdatedAt.String(), updatedCp.UpdatedAt.String())
	})

	t.Run("invalid course id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		cp, err := UpdateCourseProgressStarted(db, "1234", true)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, cp)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableCoursesProgress())
		require.Nil(t, err)

		_, err = UpdateCourseProgressStarted(db, "1234", true)
		require.ErrorContains(t, err, "no such table: "+TableCoursesProgress())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateCourseProgressPercent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)
		require.Zero(t, workingData[0].Percent)
		require.True(t, workingData[0].CompletedAt.IsZero())

		// ----------------------------
		// Set to 33
		// ----------------------------
		updatedCp1, err := UpdateCourseProgressPercent(db, workingData[0].ID, 33)
		require.Nil(t, err)
		assert.Equal(t, 33, updatedCp1.Percent)
		assert.True(t, updatedCp1.CompletedAt.IsZero())
		assert.NotEqual(t, workingData[0].UpdatedAt, updatedCp1.UpdatedAt)

		// ----------------------------
		// Set to 100
		// ----------------------------
		updatedCp2, err := UpdateCourseProgressPercent(db, workingData[0].ID, 100)
		require.Nil(t, err)
		assert.Equal(t, 100, updatedCp2.Percent)
		assert.False(t, updatedCp2.CompletedAt.IsZero())
		assert.NotEqual(t, updatedCp1.UpdatedAt, updatedCp2.UpdatedAt)

		// ----------------------------
		// Set to 50
		// ----------------------------
		updatedCp3, err := UpdateCourseProgressPercent(db, workingData[0].ID, 50)
		require.Nil(t, err)
		require.Equal(t, 50, updatedCp3.Percent)
		assert.True(t, updatedCp3.CompletedAt.IsZero())
		assert.NotEqual(t, updatedCp2.UpdatedAt, updatedCp3.UpdatedAt)
	})

	t.Run("create new", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 3, false, 0, 0)

		for i := range workingData {
			require.Zero(t, workingData[i].Percent)
			require.True(t, workingData[i].CompletedAt.IsZero())
		}

		// ----------------------------
		// Set to 10
		// ----------------------------
		cp1, err := UpdateCourseProgressPercent(db, workingData[0].ID, 10)
		require.Nil(t, err)
		assert.Equal(t, 10, cp1.Percent)
		assert.True(t, cp1.CompletedAt.IsZero())

		// ----------------------------
		// Set to 100
		// ----------------------------
		cp2, err := UpdateCourseProgressPercent(db, workingData[1].ID, 100)
		require.Nil(t, err)
		assert.Equal(t, 100, cp2.Percent)
		assert.False(t, cp2.CompletedAt.IsZero())

		// ----------------------------
		// Set to 0
		// ----------------------------
		cp3, err := UpdateCourseProgressPercent(db, workingData[2].ID, 0)
		require.Nil(t, err)
		assert.Zero(t, cp3.Percent)
		assert.True(t, cp3.CompletedAt.IsZero())
	})

	t.Run("normalize percent", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)

		origCp, err := GetCourseProgress(db, workingData[0].ID)
		require.Nil(t, err)
		require.Zero(t, origCp.Percent)
		require.True(t, origCp.CompletedAt.IsZero())

		// ----------------------------
		// Set to 101
		// ----------------------------
		updatedCp1, err := UpdateCourseProgressPercent(db, origCp.CourseID, 101)
		require.Nil(t, err)
		assert.Equal(t, 100, updatedCp1.Percent)
		assert.False(t, updatedCp1.CompletedAt.IsZero())
		assert.NotEqual(t, origCp.UpdatedAt, updatedCp1.UpdatedAt)

		// ----------------------------
		// Set to -1
		// ----------------------------
		updatedCp2, err := UpdateCourseProgressPercent(db, origCp.CourseID, -1)
		require.Nil(t, err)
		assert.Zero(t, updatedCp2.Percent)
		assert.True(t, updatedCp2.CompletedAt.IsZero())
		assert.NotEqual(t, updatedCp1.UpdatedAt, updatedCp2.UpdatedAt)
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		cp, err := UpdateCourseProgressPercent(db, "", 33)
		require.EqualError(t, err, "id cannot be empty")
		assert.Nil(t, cp)
	})

	t.Run("no change", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)

		origCp, err := GetCourseProgress(db, workingData[0].ID)
		require.Nil(t, err)
		require.Zero(t, origCp.Percent)
		require.True(t, origCp.CompletedAt.IsZero())

		updatedCp, err := UpdateCourseProgressPercent(db, origCp.CourseID, 0)
		require.Nil(t, err)
		assert.Zero(t, updatedCp.Percent)
		assert.True(t, origCp.CompletedAt.IsZero())
		assert.Equal(t, origCp.UpdatedAt.String(), updatedCp.UpdatedAt.String())
	})

	t.Run("invalid course id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		cp, err := UpdateCourseProgressPercent(db, "1234", 33)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, cp)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableCoursesProgress())
		require.Nil(t, err)

		_, err = UpdateCourseProgressPercent(db, "1234", 33)
		require.ErrorContains(t, err, "no such table: "+TableCoursesProgress())
	})
}
