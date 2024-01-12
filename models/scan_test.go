package models

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetScan(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		s, err := GetScan(db, "1234")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, s)
	})

	t.Run("found", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, true, 0, 0)

		s, err := GetScan(db, workingData[0].Course.ID)
		require.Nil(t, err)
		assert.Equal(t, workingData[0].Scan.ID, s.ID)
		assert.Equal(t, workingData[0].Course.Path, s.CoursePath)
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		result, err := GetScan(db, "")
		require.EqualError(t, err, "id cannot be empty")
		assert.Nil(t, result)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableScans())
		require.Nil(t, err)

		_, err = GetScan(db, "1234")
		require.ErrorContains(t, err, "no such table: "+TableScans())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateScan(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)
		s := newTestScan(t, nil, workingData[0].Course.ID)

		err := CreateScan(db, s)
		require.Nil(t, err)

		newS, err := GetScan(db, s.CourseID)
		require.Nil(t, err)
		assert.Equal(t, s.ID, newS.ID)
		assert.True(t, newS.Status.IsWaiting())
		assert.False(t, newS.CreatedAt.IsZero())
		assert.False(t, newS.UpdatedAt.IsZero())
	})

	t.Run("missing status", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)
		s := newTestScan(t, nil, workingData[0].Course.ID)
		s.Status = types.ScanStatus{}

		err := CreateScan(db, s)
		require.Nil(t, err)
		assert.True(t, s.Status.IsWaiting())
	})

	t.Run("duplicate course id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)
		s := newTestScan(t, nil, workingData[0].Course.ID)

		err := CreateScan(db, s)
		require.Nil(t, err)

		err = CreateScan(db, s)
		assert.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.course_id", TableScans()))
	})

	t.Run("constraint errors", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		course := NewTestData(t, db, 1, false, 0, 0)[0]

		// Missing course
		s := &Scan{}
		assert.ErrorContains(t, CreateScan(db, s), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", TableScans()))
		s.CourseID = ""
		assert.ErrorContains(t, CreateScan(db, s), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", TableScans()))
		s.CourseID = "1234"

		// Invalid Course ID
		assert.ErrorContains(t, CreateScan(db, s), "FOREIGN KEY constraint failed")
		s.CourseID = course.ID

		// Success
		assert.Nil(t, CreateScan(db, s))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateScanStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, true, 0, 0)
		require.True(t, workingData[0].Scan.Status.IsWaiting())

		// ----------------------------
		// Set to Processing
		// ----------------------------
		updatedScan1, err := UpdateScanStatus(db, workingData[0].ID, types.ScanStatusProcessing)
		require.Nil(t, err)
		require.False(t, updatedScan1.Status.IsWaiting())
		assert.NotEqual(t, workingData[0].Scan.UpdatedAt, updatedScan1.UpdatedAt)

		// ----------------------------
		// Set to waiting
		// ----------------------------
		updatedScan2, err := UpdateScanStatus(db, workingData[0].ID, types.ScanStatusWaiting)
		require.Nil(t, err)
		require.True(t, updatedScan2.Status.IsWaiting())
		assert.NotEqual(t, updatedScan1.UpdatedAt, updatedScan2.UpdatedAt)
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		updatedScan, err := UpdateScanStatus(db, "", types.ScanStatusProcessing)
		assert.EqualError(t, err, "id cannot be empty")
		assert.Nil(t, updatedScan)
	})

	t.Run("no change", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, true, 0, 0)
		require.True(t, workingData[0].Scan.Status.IsWaiting())

		updatedScan, err := UpdateScanStatus(db, workingData[0].ID, types.ScanStatusWaiting)
		require.Nil(t, err)
		assert.True(t, updatedScan.Status.IsWaiting())
		assert.Equal(t, workingData[0].Scan.UpdatedAt.String(), updatedScan.UpdatedAt.String())
	})

	t.Run("no course with id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		updatedScan, err := UpdateScanStatus(db, "1234", types.ScanStatusProcessing)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, updatedScan)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableScans())
		require.Nil(t, err)

		_, err = UpdateScanStatus(db, "1234", types.ScanStatusProcessing)
		require.ErrorContains(t, err, "no such table: "+TableScans())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteScan(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, true, 0, 0)
		err := DeleteScan(db, workingData[0].ID)
		require.Nil(t, err)
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		err := DeleteScan(db, "")
		assert.ErrorContains(t, err, "id cannot be empty")
	})

	t.Run("invalid id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		err := DeleteScan(db, "1234")
		assert.Nil(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableScans())
		require.Nil(t, err)

		err = DeleteScan(db, "1234")
		require.ErrorContains(t, err, "no such table: "+TableScans())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteScanCascade(t *testing.T) {
	_, db, teardown := setup(t)
	defer teardown(t)

	workingData := NewTestData(t, db, 1, true, 0, 0)

	err := DeleteCourse(db, workingData[0].ID)
	require.Nil(t, err)

	s, err := GetScan(db, workingData[0].ID)
	require.ErrorIs(t, err, sql.ErrNoRows)
	assert.Nil(t, s)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_NextScan(t *testing.T) {
	t.Run("first", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, true, 0, 0)

		s, err := NextScan(db)
		require.Nil(t, err)
		assert.Equal(t, workingData[0].Scan.ID, s.ID)
		assert.Equal(t, workingData[0].Path, s.CoursePath)

	})

	t.Run("next", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 3, true, 0, 0)

		// Update the the first scan to processing
		_, err := UpdateScanStatus(db, workingData[0].ID, types.ScanStatusProcessing)
		require.Nil(t, err)

		// Get the next scan
		s, err := NextScan(db)
		require.Nil(t, err)
		assert.Equal(t, workingData[1].Scan.ID, s.ID)
		assert.Equal(t, workingData[1].Path, s.CoursePath)

	})

	t.Run("empty", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		scan, err := NextScan(db)
		require.Nil(t, err)
		assert.Nil(t, scan)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableScans())
		require.Nil(t, err)

		_, err = NextScan(db)
		require.ErrorContains(t, err, "no such table: "+TableScans())
	})
}
