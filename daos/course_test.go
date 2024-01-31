package daos

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func courseSetup(t *testing.T) (*appFs.AppFs, *CourseDao, database.Database) {
	appFs, db := setup(t)
	courseDao := NewCourseDao(db)
	return appFs, courseDao, db
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourse_Count(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, dao, _ := courseSetup(t)

		count, err := dao.Count(nil)
		require.Nil(t, err)
		assert.Zero(t, count)
	})

	t.Run("entries", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		NewTestData(t, db, 5, false, 0, 0)

		count, err := dao.Count(nil)
		require.Nil(t, err)
		assert.Equal(t, count, 5)
	})

	t.Run("where", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		workingData := NewTestData(t, db, 3, false, 0, 0)

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		count, err := dao.Count(&database.DatabaseParams{Where: squirrel.Eq{dao.table + ".id": workingData[2].ID}})
		require.Nil(t, err)
		assert.Equal(t, 1, count)

		// ----------------------------
		// NOT EQUALS ID
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.NotEq{dao.table + ".id": workingData[2].ID}})
		require.Nil(t, err)
		assert.Equal(t, 2, count)

		// ----------------------------
		// ERROR
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		assert.Equal(t, 0, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		_, err = dao.Count(nil)
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourse_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, _ := courseSetup(t)

		workingData := NewTestData(t, nil, 1, false, 0, 0)

		err := dao.Create(workingData[0].Course)
		require.Nil(t, err)

		newC, err := dao.Get(workingData[0].ID)
		require.Nil(t, err)
		assert.NotEmpty(t, newC.ID)
		assert.Equal(t, workingData[0].Title, newC.Title)
		assert.Equal(t, workingData[0].Path, newC.Path)
		assert.Empty(t, newC.CardPath)
		assert.False(t, newC.Available)
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
		_, dao, _ := courseSetup(t)

		workingData := NewTestData(t, nil, 1, false, 0, 0)

		err := dao.Create(workingData[0].Course)
		require.Nil(t, err)

		err = dao.Create(workingData[0].Course)
		require.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.path", dao.table))
	})

	t.Run("constraints", func(t *testing.T) {
		_, dao, _ := courseSetup(t)

		// No title
		c := &models.Course{}
		require.ErrorContains(t, dao.Create(c), fmt.Sprintf("NOT NULL constraint failed: %s.title", dao.table))
		c.Title = ""
		require.ErrorContains(t, dao.Create(c), fmt.Sprintf("NOT NULL constraint failed: %s.title", dao.table))
		c.Title = "Course 1"

		// No path
		require.ErrorContains(t, dao.Create(c), fmt.Sprintf("NOT NULL constraint failed: %s.path", dao.table))
		c.Path = ""
		require.ErrorContains(t, dao.Create(c), fmt.Sprintf("NOT NULL constraint failed: %s.path", dao.table))
		c.Path = "/course 1"

		// Success
		require.Nil(t, dao.Create(c))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourse_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		workingData := NewTestData(t, db, 2, false, 1, 0)

		c, err := dao.Get(workingData[1].ID)
		require.Nil(t, err)
		assert.Equal(t, workingData[1].ID, c.ID)
		assert.Empty(t, workingData[1].ScanStatus)

		// ----------------------------
		// scan
		// ----------------------------
		newTestScan(t, db, workingData[1].ID)

		c, err = dao.Get(workingData[1].ID)
		require.Nil(t, err)
		assert.Equal(t, string(types.ScanStatusWaiting), c.ScanStatus)

		// ----------------------------
		// Availability
		// ----------------------------
		require.False(t, c.Available)

		// Set to started
		workingData[1].Available = true
		require.Nil(t, dao.Update(workingData[1].Course))

		c, err = dao.Get(workingData[1].ID)
		require.Nil(t, err)
		require.True(t, c.Available)

		// ----------------------------
		// Progress
		// ----------------------------
		require.False(t, c.Started)
		require.True(t, c.StartedAt.IsZero())
		require.Zero(t, c.Percent)
		require.True(t, c.CompletedAt.IsZero())

		// Get course progress
		// cpDao := NewCourseProgressDao(db)
		// origCp, err := cpDao.Get(workingData[1].ID)
		// require.Nil(t, err)

		// Update
		// err = cpDao.Update(origCp)
		// require.Nil(t, err)

		// updatedCp, err :=  cpDao.Get(workingData[1].ID)
		// require.Nil(t, err)
		// assert.True(t, updatedCp.Started)
		// assert.False(t, updatedCp.StartedAt.IsZero())

		// 		// Set to started
		// 		_, err = UpdateCourseProgressStarted(db, workingData[1].ID, true)
		// 		require.Nil(t, err)

		// 		c, err = GetCourse(db, workingData[1].ID)
		// 		require.Nil(t, err)
		// 		require.True(t, c.Started)
		// 		require.False(t, c.StartedAt.IsZero())
		// 		require.Zero(t, c.Percent)
		// 		require.True(t, c.CompletedAt.IsZero())

		// 		// Mark asset as completed (only 1 asset so the course will be 100%)
		// 		_, err = UpdateAssetProgressCompleted(db, workingData[1].Assets[0].ID, true)
		// 		require.Nil(t, err)

		// 		c, err = GetCourse(db, workingData[1].ID)
		// 		require.Nil(t, err)
		// 		require.True(t, c.Started)
		// 		require.False(t, c.StartedAt.IsZero())
		// 		require.Equal(t, 100, c.Percent)
		// 		require.False(t, c.CompletedAt.IsZero())
	})

	t.Run("not found", func(t *testing.T) {
		_, dao, _ := courseSetup(t)

		c, err := dao.Get("1234")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, c)
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, _ := courseSetup(t)

		c, err := dao.Get("")
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, c)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		_, err = dao.Get("1234")
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourse_List(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, dao, _ := courseSetup(t)

		courses, err := dao.List(nil)
		require.Nil(t, err)
		require.Zero(t, courses)
	})

	t.Run("found", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		workingData := NewTestData(t, db, 5, false, 2, 0)

		result, err := dao.List(nil)
		require.Nil(t, err)
		require.Len(t, result, 5)

		// ----------------------------
		// Scan
		// ----------------------------
		assert.Empty(t, result[1].ScanStatus)
		newTestScan(t, db, workingData[1].ID)

		result, err = dao.List(nil)
		require.Nil(t, err)
		assert.Equal(t, workingData[1].ID, result[1].ID)
		assert.Equal(t, string(types.ScanStatusWaiting), result[1].ScanStatus)

		// ----------------------------
		// Availability
		// ----------------------------
		for _, c := range result {
			require.False(t, c.Available)
		}

		workingData[0].Available = true

		// Set course 1 as available
		require.Nil(t, dao.Update(workingData[0].Course))

		// Find available courses
		result, err = dao.List(&database.DatabaseParams{Where: squirrel.And{squirrel.Eq{dao.table + ".available": true}}})
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, workingData[0].ID, result[0].ID)

		// ----------------------------
		// Progress
		// ----------------------------
		apDao := NewAssetProgressDao(db)

		for _, c := range result {
			require.False(t, c.Started)
			require.True(t, c.StartedAt.IsZero())
			require.Zero(t, c.Percent)
			require.True(t, c.CompletedAt.IsZero())
		}

		// Set asset progress for asset 1 as started
		ap1 := newTestAssetsProgress(t, db, workingData[0].Assets[0].ID, workingData[0].ID)
		ap1.VideoPos = 50
		require.Nil(t, apDao.Update(ap1))

		// Find started courses (not completed)
		dbParams := &database.DatabaseParams{
			Where: squirrel.And{
				squirrel.Eq{TableCoursesProgress() + ".started": true},
				squirrel.NotEq{TableCoursesProgress() + ".percent": 100},
			},
		}

		result, err = dao.List(dbParams)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, workingData[0].ID, result[0].ID)

		// Set both assets as completed
		ap1.Completed = true
		require.Nil(t, apDao.Update(ap1))

		ap2 := newTestAssetsProgress(t, db, workingData[0].Assets[1].ID, workingData[0].ID)
		ap2.Completed = true
		require.Nil(t, apDao.Update(ap2))

		// Find completed courses
		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Eq{TableCoursesProgress() + ".percent": 100}})
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, workingData[0].ID, result[0].ID)
	})

	t.Run("orderby", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		workingData := NewTestData(t, db, 3, false, 0, 0)

		// ----------------------------
		// CREATED_AT DESC
		// ----------------------------
		dbParams := &database.DatabaseParams{OrderBy: []string{"created_at desc"}}
		result, err := dao.List(dbParams)
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, workingData[2].ID, result[0].ID)

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{"created_at asc"}})
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, workingData[0].ID, result[0].ID)

		// ----------------------------
		// SCAN_STATUS DESC
		// ----------------------------

		// Create a scan for course 2 and 3
		workingData[1].Scan = newTestScan(t, db, workingData[1].ID)
		workingData[2].Scan = newTestScan(t, db, workingData[2].ID)

		// Set course 3 to processing
		scanDao := NewScanDao(db)
		workingData[2].Scan.Status = types.NewScanStatus(types.ScanStatusProcessing)
		require.Nil(t, scanDao.Update(workingData[2].Scan))

		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{TableScans() + ".status desc"}})
		require.Nil(t, err)
		require.Len(t, result, 3)

		assert.Equal(t, workingData[0].ID, result[2].ID)
		assert.Equal(t, workingData[1].ID, result[1].ID)
		assert.Equal(t, workingData[2].ID, result[0].ID)

		// ----------------------------
		// SCAN_STATUS ASC
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{"scan_status asc"}})
		require.Nil(t, err)
		require.Len(t, result, 3)

		assert.Equal(t, workingData[0].ID, result[0].ID)
		assert.Equal(t, workingData[1].ID, result[1].ID)
		assert.Equal(t, workingData[2].ID, result[2].ID)

		// ----------------------------
		// Error
		// ----------------------------
		dbParams = &database.DatabaseParams{OrderBy: []string{"unit_test asc"}}
		result, err = dao.List(dbParams)
		require.ErrorContains(t, err, "no such column")
		assert.Nil(t, result)
	})

	t.Run("where", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		workingData := NewTestData(t, db, 3, false, 0, 0)

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		result, err := dao.List(&database.DatabaseParams{Where: squirrel.Eq{dao.table + ".id": workingData[2].ID}})
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, workingData[2].ID, result[0].ID)

		// ----------------------------
		// EQUALS ID OR ID
		// ----------------------------
		dbParams := &database.DatabaseParams{
			Where:   squirrel.Or{squirrel.Eq{dao.table + ".id": workingData[1].ID}, squirrel.Eq{dao.table + ".id": workingData[2].ID}},
			OrderBy: []string{"created_at asc"},
		}
		result, err = dao.List(dbParams)
		require.Nil(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, workingData[1].ID, result[0].ID)
		assert.Equal(t, workingData[2].ID, result[1].ID)

		// ----------------------------
		// ERROR
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		assert.Nil(t, result)
	})

	t.Run("pagination", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		workingData := NewTestData(t, db, 17, false, 0, 0)

		// ----------------------------
		// Page 1 with 10 items
		// ----------------------------
		p := pagination.New(1, 10)

		result, err := dao.List(&database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, workingData[0].ID, result[0].ID)
		assert.Equal(t, workingData[9].ID, result[9].ID)

		// ----------------------------
		// Page 2 with 7 items
		// ----------------------------
		p = pagination.New(2, 10)

		result, err = dao.List(&database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 7)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, workingData[10].ID, result[0].ID)
		assert.Equal(t, workingData[16].ID, result[6].ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		_, err = dao.List(nil)
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourse_Update(t *testing.T) {
	t.Run("card path", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)
		require.Empty(t, workingData[0].CardPath)

		// Update the card path
		workingData[0].CardPath = "/path/to/card.jpg"
		require.Nil(t, dao.Update(workingData[0].Course))

		c, err := dao.Get(workingData[0].ID)
		require.Nil(t, err)
		assert.Equal(t, workingData[0].CardPath, c.CardPath)
	})

	t.Run("available", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)
		require.False(t, workingData[0].Available)

		// Update the availability
		workingData[0].Available = true
		require.Nil(t, dao.Update(workingData[0].Course))

		c, err := dao.Get(workingData[0].ID)
		require.Nil(t, err)
		assert.True(t, c.Available)
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, _ := courseSetup(t)

		err := dao.Update(&models.Course{})
		assert.ErrorIs(t, err, ErrEmptyId)
	})

	t.Run("invalid id", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		workingData := NewTestData(t, db, 1, true, 0, 0)
		workingData[0].ID = "1234"

		require.Nil(t, dao.Update(workingData[0].Course))
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		workingData := NewTestData(t, nil, 1, false, 0, 0)

		err = dao.Update(workingData[0].Course)
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourse_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)
		err := dao.Delete(workingData[0].ID)
		require.Nil(t, err)
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, _ := courseSetup(t)

		err := dao.Delete("")
		assert.ErrorContains(t, err, "id cannot be empty")
	})

	t.Run("invalid id", func(t *testing.T) {
		_, dao, _ := courseSetup(t)

		err := dao.Delete("1234")
		assert.Nil(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		err = dao.Delete("1234")
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}
