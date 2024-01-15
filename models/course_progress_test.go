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

		workingData := NewTestData(t, db, 1, false, 2, 0)
		require.False(t, workingData[0].Started)
		require.True(t, workingData[0].StartedAt.IsZero())
		require.Zero(t, workingData[0].Percent)
		require.True(t, workingData[0].CompletedAt.IsZero())

		// ----------------------------
		// Asset 1 completed
		// ----------------------------
		ap1, err := UpdateAssetProgressCompleted(db, workingData[0].Assets[0].ID, true)
		require.Nil(t, err)
		require.True(t, ap1.Completed)
		require.False(t, ap1.CompletedAt.IsZero())

		cp1, err := GetCourse(db, workingData[0].ID)
		require.Nil(t, err)
		require.True(t, cp1.Started)
		require.False(t, cp1.StartedAt.IsZero())
		require.Equal(t, 50, cp1.Percent)
		require.True(t, cp1.CompletedAt.IsZero())

		// ----------------------------
		// Asset 2 completed
		// ----------------------------
		ap2, err := UpdateAssetProgressCompleted(db, workingData[0].Assets[1].ID, true)
		require.Nil(t, err)
		require.True(t, ap2.Completed)
		require.False(t, ap2.CompletedAt.IsZero())

		cp2, err := GetCourse(db, workingData[0].ID)
		require.Nil(t, err)
		require.True(t, cp2.Started)
		require.Equal(t, cp1.StartedAt.String(), cp2.StartedAt.String())
		require.Equal(t, 100, cp2.Percent)
		require.False(t, cp2.CompletedAt.IsZero())

		// ----------------------------
		// Asset 1 not completed
		// ----------------------------
		ap3, err := UpdateAssetProgressCompleted(db, workingData[0].Assets[0].ID, false)
		require.Nil(t, err)
		require.False(t, ap3.Completed)
		require.True(t, ap3.CompletedAt.IsZero())

		cp3, err := GetCourse(db, workingData[0].ID)
		require.Nil(t, err)
		require.True(t, cp3.Started)
		require.Equal(t, cp1.StartedAt.String(), cp2.StartedAt.String())
		require.Equal(t, 50, cp3.Percent)
		require.True(t, cp3.CompletedAt.IsZero())
	})

	t.Run("no assets", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)
		require.False(t, workingData[0].Started)
		require.True(t, workingData[0].StartedAt.IsZero())
		require.Zero(t, workingData[0].Percent)
		require.True(t, workingData[0].CompletedAt.IsZero())

		cp, err := UpdateCourseProgressPercent(db, workingData[0].ID)
		require.Nil(t, err)
		require.False(t, cp.Started)
		require.True(t, cp.StartedAt.IsZero())
		assert.Zero(t, cp.Percent)
		assert.True(t, cp.CompletedAt.IsZero())
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		cp, err := UpdateCourseProgressPercent(db, "")
		require.EqualError(t, err, "id cannot be empty")
		assert.Nil(t, cp)
	})

	t.Run("invalid course id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		cp, err := UpdateCourseProgressPercent(db, "1234")
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, cp)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableCoursesProgress())
		require.Nil(t, err)

		_, err = UpdateCourseProgressPercent(db, "1234")
		require.ErrorContains(t, err, "no such table: "+TableCoursesProgress())
	})
}
