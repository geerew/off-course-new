package daos

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/types"
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

		testData := NewTestBuilder(t).Db(db).Courses(1).Build()
		s := &models.Scan{CourseID: testData[0].Course.ID}

		err := dao.Create(s)
		require.Nil(t, err, "Failed to create scan")

		newS, err := dao.Get(s.CourseID)
		require.Nil(t, err)
		require.Equal(t, s.ID, newS.ID)
		require.True(t, newS.Status.IsWaiting())
		require.False(t, newS.CreatedAt.IsZero())
		require.False(t, newS.UpdatedAt.IsZero())

	})

	t.Run("duplicate course id", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Build()
		s := &models.Scan{CourseID: testData[0].Course.ID}

		err := dao.Create(s)
		require.Nil(t, err)

		err = dao.Create(s)
		require.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.course_id", dao.Table))
	})

	t.Run("constraint errors", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Build()

		// Missing course ID
		s := &models.Scan{}
		require.ErrorContains(t, dao.Create(s), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", dao.Table))
		s.CourseID = ""
		require.ErrorContains(t, dao.Create(s), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", dao.Table))
		s.CourseID = "1234"

		// Invalid Course ID
		require.ErrorContains(t, dao.Create(s), "FOREIGN KEY constraint failed")
		s.CourseID = testData[0].Course.ID

		// Success
		require.Nil(t, dao.Create(s))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScan_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Scan().Build()

		s, err := dao.Get(testData[0].Course.ID)
		require.Nil(t, err)
		require.Equal(t, testData[0].Scan.ID, s.ID)
		require.Equal(t, testData[0].Course.Path, s.CoursePath)
	})

	t.Run("not found", func(t *testing.T) {
		_, dao, _ := scanSetup(t)

		s, err := dao.Get("1234")
		require.ErrorIs(t, err, sql.ErrNoRows)
		require.Nil(t, s)
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, _ := scanSetup(t)

		s, err := dao.Get("")
		require.ErrorIs(t, err, sql.ErrNoRows)
		require.Nil(t, s)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table)
		require.Nil(t, err)

		_, err = dao.Get("1234")
		require.ErrorContains(t, err, "no such table: "+dao.Table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScan_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Scan().Build()
		require.True(t, testData[0].Scan.Status.IsWaiting())

		// ----------------------------
		// Set to Processing
		// ----------------------------
		testData[0].Scan.Status = types.NewScanStatus(types.ScanStatusProcessing)
		require.Nil(t, dao.Update(testData[0].Scan))

		updatedScan, err := dao.Get(testData[0].Course.ID)
		require.Nil(t, err)
		require.False(t, updatedScan.Status.IsWaiting())

		// ----------------------------
		// Set to waiting
		// ----------------------------
		time.Sleep(1 * time.Millisecond)
		testData[0].Scan.Status = types.NewScanStatus(types.ScanStatusWaiting)
		require.Nil(t, dao.Update(testData[0].Scan))

		updatedScan, err = dao.Get(testData[0].Course.ID)
		require.Nil(t, err)
		require.True(t, updatedScan.Status.IsWaiting())
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, _ := scanSetup(t)

		err := dao.Update(&models.Scan{})
		require.ErrorIs(t, err, ErrEmptyId)
	})

	t.Run("invalid id", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Scan().Build()
		testData[0].Scan.ID = "1234"

		err := dao.Update(testData[0].Scan)
		require.Nil(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table)
		require.Nil(t, err)

		testData := NewTestBuilder(t).Courses(1).Scan().Build()
		err = dao.Update(testData[0].Scan)
		require.ErrorContains(t, err, "no such table: "+dao.Table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScan_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Scan().Build()
		err := dao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": testData[0].ID}}, nil)
		require.Nil(t, err)
	})

	t.Run("no db params", func(t *testing.T) {
		_, dao, _ := scanSetup(t)

		err := dao.Delete(nil, nil)
		require.ErrorIs(t, err, ErrMissingWhere)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table)
		require.Nil(t, err)

		err = dao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": "1234"}}, nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScan_DeleteCascade(t *testing.T) {
	_, dao, db := scanSetup(t)

	testData := NewTestBuilder(t).Db(db).Courses(1).Scan().Build()

	// Delete the course
	courseDao := NewCourseDao(db)
	err := courseDao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": testData[0].ID}}, nil)
	require.Nil(t, err)

	// Check the scan was deleted
	s, err := dao.Get(testData[0].ID)
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.Nil(t, s)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScan_NextScan(t *testing.T) {
	t.Run("first", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Scan().Build()

		s, err := dao.Next()
		require.Nil(t, err)
		require.Equal(t, testData[0].Scan.ID, s.ID)
		require.Equal(t, testData[0].Path, s.CoursePath)
	})

	t.Run("next", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(3).Scan().Build()

		// Update the the first scan to processing
		testData[0].Scan.Status = types.NewScanStatus(types.ScanStatusProcessing)
		require.Nil(t, dao.Update(testData[0].Scan))

		s, err := dao.Next()
		require.Nil(t, err)
		require.Equal(t, testData[1].Scan.ID, s.ID)
		require.Equal(t, testData[1].Path, s.CoursePath)

	})

	t.Run("empty", func(t *testing.T) {
		_, dao, _ := scanSetup(t)

		scan, err := dao.Next()
		require.Nil(t, err)
		require.Nil(t, scan)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table)
		require.Nil(t, err)

		_, err = dao.Next()
		require.ErrorContains(t, err, "no such table: "+dao.Table)
	})
}
