package daos

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func scanSetup(t *testing.T) (*appFs.AppFs, *ScanDao, database.Database) {
	appFs, db := setup(t)
	scanDao := NewScanDao(db)
	return appFs, scanDao, db
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScan_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)
		s := newTestScan(t, nil, workingData[0].Course.ID)

		err := dao.Create(s)
		require.Nil(t, err, "Failed to create scan")

		newS, err := dao.Get(s.CourseID)
		require.Nil(t, err)
		assert.Equal(t, s.ID, newS.ID)
		assert.True(t, newS.Status.IsWaiting())
		assert.False(t, newS.CreatedAt.IsZero())
		assert.False(t, newS.UpdatedAt.IsZero())

	})

	t.Run("duplicate course id", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)
		s := newTestScan(t, nil, workingData[0].Course.ID)

		err := dao.Create(s)
		require.Nil(t, err)

		err = dao.Create(s)
		require.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.course_id", dao.table))
	})

	t.Run("constraint errors", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		course := NewTestData(t, db, 1, false, 0, 0)[0]

		// Missing course ID
		s := &models.Scan{}
		require.ErrorContains(t, dao.Create(s), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", dao.table))
		s.CourseID = ""
		require.ErrorContains(t, dao.Create(s), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", dao.table))
		s.CourseID = "1234"

		// Invalid Course ID
		require.ErrorContains(t, dao.Create(s), "FOREIGN KEY constraint failed")
		s.CourseID = course.ID

		// Success
		require.Nil(t, dao.Create(s))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScan_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		workingData := NewTestData(t, db, 1, true, 0, 0)

		s, err := dao.Get(workingData[0].Course.ID)
		require.Nil(t, err)
		assert.Equal(t, workingData[0].Scan.ID, s.ID)
		assert.Equal(t, workingData[0].Course.Path, s.CoursePath)
	})

	t.Run("not found", func(t *testing.T) {
		_, dao, _ := scanSetup(t)

		s, err := dao.Get("1234")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, s)
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, _ := scanSetup(t)

		s, err := dao.Get("")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, s)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		_, err = dao.Get("1234")
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScan_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		workingData := NewTestData(t, db, 1, true, 0, 0)
		require.True(t, workingData[0].Scan.Status.IsWaiting())

		// ----------------------------
		// Set to Processing
		// ----------------------------
		workingData[0].Scan.Status = types.NewScanStatus(types.ScanStatusProcessing)
		require.Nil(t, dao.Update(workingData[0].Scan))

		updatedScan, err := dao.Get(workingData[0].Course.ID)
		require.Nil(t, err)
		require.False(t, updatedScan.Status.IsWaiting())

		// ----------------------------
		// Set to waiting
		// ----------------------------
		time.Sleep(1 * time.Millisecond)
		workingData[0].Scan.Status = types.NewScanStatus(types.ScanStatusWaiting)
		require.Nil(t, dao.Update(workingData[0].Scan))

		updatedScan, err = dao.Get(workingData[0].Course.ID)
		require.Nil(t, err)
		require.True(t, updatedScan.Status.IsWaiting())
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, _ := scanSetup(t)

		err := dao.Update(&models.Scan{})
		assert.ErrorIs(t, err, ErrEmptyId)
	})

	t.Run("invalid id", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		workingData := NewTestData(t, db, 1, true, 0, 0)
		workingData[0].Scan.ID = "1234"

		err := dao.Update(workingData[0].Scan)
		assert.Nil(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		workingData := NewTestData(t, nil, 1, true, 0, 0)
		err = dao.Update(workingData[0].Scan)
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScan_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		workingData := NewTestData(t, db, 1, true, 0, 0)
		err := dao.Delete(workingData[0].ID)
		require.Nil(t, err)
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, _ := scanSetup(t)

		err := dao.Delete("")
		assert.ErrorContains(t, err, "id cannot be empty")
	})

	t.Run("invalid id", func(t *testing.T) {
		_, dao, _ := scanSetup(t)

		err := dao.Delete("1234")
		assert.Nil(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		err = dao.Delete("1234")
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScan_DeleteCascade(t *testing.T) {
	_, dao, db := scanSetup(t)

	workingData := NewTestData(t, db, 1, true, 0, 0)

	// Delete the course
	courseDao := NewCourseDao(db)
	err := courseDao.Delete(workingData[0].ID)
	require.Nil(t, err)

	// Check the scan was deleted
	s, err := dao.Get(workingData[0].ID)
	require.ErrorIs(t, err, sql.ErrNoRows)
	assert.Nil(t, s)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScan_NextScan(t *testing.T) {
	t.Run("first", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		workingData := NewTestData(t, db, 1, true, 0, 0)

		s, err := dao.Next()
		require.Nil(t, err)
		assert.Equal(t, workingData[0].Scan.ID, s.ID)
		assert.Equal(t, workingData[0].Path, s.CoursePath)
	})

	t.Run("next", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		workingData := NewTestData(t, db, 3, true, 0, 0)

		// Update the the first scan to processing
		workingData[0].Scan.Status = types.NewScanStatus(types.ScanStatusProcessing)
		require.Nil(t, dao.Update(workingData[0].Scan))

		s, err := dao.Next()
		require.Nil(t, err)
		assert.Equal(t, workingData[1].Scan.ID, s.ID)
		assert.Equal(t, workingData[1].Path, s.CoursePath)

	})

	t.Run("empty", func(t *testing.T) {
		_, dao, _ := scanSetup(t)

		scan, err := dao.Next()
		require.Nil(t, err)
		assert.Nil(t, scan)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		_, err = dao.Next()
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}
