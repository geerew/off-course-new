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

		NewTestBuilder(t).Db(db).Courses(5).Build()

		count, err := dao.Count(nil)
		require.Nil(t, err)
		assert.Equal(t, count, 5)
	})

	t.Run("where", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(3).Build()

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		count, err := dao.Count(&database.DatabaseParams{Where: squirrel.Eq{dao.Table + ".id": testData[2].ID}})
		require.Nil(t, err)
		assert.Equal(t, 1, count)

		// ----------------------------
		// NOT EQUALS ID
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.NotEq{dao.Table + ".id": testData[2].ID}})
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

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table)
		require.Nil(t, err)

		_, err = dao.Count(nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourse_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, _ := courseSetup(t)

		testData := NewTestBuilder(t).Courses(1).Build()

		err := dao.Create(testData[0].Course)
		require.Nil(t, err)

		newC, err := dao.Get(testData[0].ID, nil)
		require.Nil(t, err)
		assert.NotEmpty(t, newC.ID)
		assert.Equal(t, testData[0].Title, newC.Title)
		assert.Equal(t, testData[0].Path, newC.Path)
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

		testData := NewTestBuilder(t).Courses(1).Build()

		err := dao.Create(testData[0].Course)
		require.Nil(t, err)

		err = dao.Create(testData[0].Course)
		require.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.path", dao.Table))
	})

	t.Run("constraints", func(t *testing.T) {
		_, dao, _ := courseSetup(t)

		// No title
		c := &models.Course{}
		require.ErrorContains(t, dao.Create(c), fmt.Sprintf("NOT NULL constraint failed: %s.title", dao.Table))
		c.Title = ""
		require.ErrorContains(t, dao.Create(c), fmt.Sprintf("NOT NULL constraint failed: %s.title", dao.Table))
		c.Title = "Course 1"

		// No path
		require.ErrorContains(t, dao.Create(c), fmt.Sprintf("NOT NULL constraint failed: %s.path", dao.Table))
		c.Path = ""
		require.ErrorContains(t, dao.Create(c), fmt.Sprintf("NOT NULL constraint failed: %s.path", dao.Table))
		c.Path = "/course 1"

		// Success
		require.Nil(t, dao.Create(c))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourse_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(2).Assets(1).Build()

		c, err := dao.Get(testData[1].ID, nil)
		require.Nil(t, err)
		assert.Equal(t, testData[1].ID, c.ID)
		assert.Empty(t, testData[1].ScanStatus)

		// ----------------------------
		// scan
		// ----------------------------
		scanDao := NewScanDao(db)
		require.Nil(t, scanDao.Create(&models.Scan{CourseID: testData[1].ID}))

		c, err = dao.Get(testData[1].ID, nil)
		require.Nil(t, err)
		assert.Equal(t, string(types.ScanStatusWaiting), c.ScanStatus)

		// ----------------------------
		// Availability
		// ----------------------------
		require.False(t, c.Available)

		// Set to started
		testData[1].Available = true
		require.Nil(t, dao.Update(testData[1].Course))

		c, err = dao.Get(testData[1].ID, nil)
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
		// origCp, err := cpDao.Get(testData[1].ID, nil)
		// require.Nil(t, err)

		// Update
		// err = cpDao.Update(origCp)
		// require.Nil(t, err)

		// updatedCp, err :=  cpDao.Get(testData[1].ID, nil)
		// require.Nil(t, err)
		// assert.True(t, updatedCp.Started)
		// assert.False(t, updatedCp.StartedAt.IsZero())

		// 		// Set to started
		// 		_, err = UpdateCourseProgressStarted(db, testData[1].ID, true)
		// 		require.Nil(t, err)

		// 		c, err = GetCourse(db, testData[1].ID)
		// 		require.Nil(t, err)
		// 		require.True(t, c.Started)
		// 		require.False(t, c.StartedAt.IsZero())
		// 		require.Zero(t, c.Percent)
		// 		require.True(t, c.CompletedAt.IsZero())

		// 		// Mark asset as completed (only 1 asset so the course will be 100%)
		// 		_, err = UpdateAssetProgressCompleted(db, testData[1].Assets[0].ID, true)
		// 		require.Nil(t, err)

		// 		c, err = GetCourse(db, testData[1].ID)
		// 		require.Nil(t, err)
		// 		require.True(t, c.Started)
		// 		require.False(t, c.StartedAt.IsZero())
		// 		require.Equal(t, 100, c.Percent)
		// 		require.False(t, c.CompletedAt.IsZero())
	})

	t.Run("not found", func(t *testing.T) {
		_, dao, _ := courseSetup(t)

		c, err := dao.Get("1234", nil)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, c)
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, _ := courseSetup(t)

		c, err := dao.Get("", nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, c)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table)
		require.Nil(t, err)

		_, err = dao.Get("1234", nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourse_List(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, dao, _ := courseSetup(t)

		courses, err := dao.List(nil, nil)
		require.Nil(t, err)
		require.Zero(t, courses)
	})

	t.Run("found", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(5).Assets(2).Build()

		result, err := dao.List(nil, nil)
		require.Nil(t, err)
		require.Len(t, result, 5)

		// ----------------------------
		// Scan
		// ----------------------------
		assert.Empty(t, result[1].ScanStatus)

		// Create a scan for course 2
		scanDao := NewScanDao(db)
		require.Nil(t, scanDao.Create(&models.Scan{CourseID: testData[1].ID}))

		result, err = dao.List(nil, nil)
		require.Nil(t, err)
		assert.Equal(t, testData[1].ID, result[1].ID)
		assert.Equal(t, string(types.ScanStatusWaiting), result[1].ScanStatus)

		// ----------------------------
		// Availability
		// ----------------------------
		for _, c := range result {
			require.False(t, c.Available)
		}

		// Set course 1 as available
		testData[0].Available = true
		require.Nil(t, dao.Update(testData[0].Course))

		// Find available courses
		result, err = dao.List(&database.DatabaseParams{Where: squirrel.And{squirrel.Eq{dao.Table + ".available": true}}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, testData[0].ID, result[0].ID)

		// ----------------------------
		// Progress
		// ----------------------------
		apDao := NewAssetProgressDao(db)
		cpDao := NewCourseProgressDao(db)

		for _, c := range result {
			require.False(t, c.Started)
			require.True(t, c.StartedAt.IsZero())
			require.Zero(t, c.Percent)
			require.True(t, c.CompletedAt.IsZero())
		}

		// Create progress for asset 1 in course 1 and set the video position to 50
		ap1 := &models.AssetProgress{AssetID: testData[0].Assets[0].ID, CourseID: testData[0].ID, VideoPos: 50}

		require.Nil(t, apDao.Create(ap1, nil))

		// Find in-progress courses
		dbParams := &database.DatabaseParams{
			Where: squirrel.And{
				squirrel.Eq{cpDao.Table + ".started": true},
				squirrel.NotEq{cpDao.Table + ".percent": 100},
			},
		}

		result, err = dao.List(dbParams, nil)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, testData[0].ID, result[0].ID)

		// Set progress for asset 1 in course 1 as complete
		ap1.Completed = true
		require.Nil(t, apDao.Update(ap1, nil))

		// Create progress for asset 2 in course 1 and set completed to true
		ap2 := &models.AssetProgress{AssetID: testData[0].Assets[1].ID, CourseID: testData[0].ID, Completed: true}
		require.Nil(t, apDao.Create(ap2, nil))

		// Find completed courses
		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Eq{cpDao.Table + ".percent": 100}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, testData[0].ID, result[0].ID)
	})

	t.Run("orderby", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(3).Build()

		// ----------------------------
		// CREATED_AT DESC
		// ----------------------------
		dbParams := &database.DatabaseParams{OrderBy: []string{"created_at desc"}}
		result, err := dao.List(dbParams, nil)
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, testData[2].ID, result[0].ID)

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{"created_at asc"}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, testData[0].ID, result[0].ID)

		// ----------------------------
		// SCAN_STATUS DESC
		// ----------------------------

		// Create a scan for course 2 and 3
		scanDao := NewScanDao(db)

		testData[1].Scan = &models.Scan{CourseID: testData[1].ID}
		require.Nil(t, scanDao.Create(testData[1].Scan))
		testData[2].Scan = &models.Scan{CourseID: testData[2].ID}
		require.Nil(t, scanDao.Create(testData[2].Scan))

		// Set course 3 to processing
		testData[2].Scan.Status = types.NewScanStatus(types.ScanStatusProcessing)
		require.Nil(t, scanDao.Update(testData[2].Scan))

		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{scanDao.Table + ".status desc"}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 3)

		assert.Equal(t, testData[0].ID, result[2].ID)
		assert.Equal(t, testData[1].ID, result[1].ID)
		assert.Equal(t, testData[2].ID, result[0].ID)

		// ----------------------------
		// SCAN_STATUS ASC
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{scanDao.Table + ".status asc"}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 3)

		assert.Equal(t, testData[0].ID, result[0].ID)
		assert.Equal(t, testData[1].ID, result[1].ID)
		assert.Equal(t, testData[2].ID, result[2].ID)

		// ----------------------------
		// Error
		// ----------------------------
		dbParams = &database.DatabaseParams{OrderBy: []string{"unit_test asc"}}
		result, err = dao.List(dbParams, nil)
		require.ErrorContains(t, err, "no such column")
		assert.Nil(t, result)
	})

	t.Run("where", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(3).Build()

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		result, err := dao.List(&database.DatabaseParams{Where: squirrel.Eq{dao.Table + ".id": testData[2].ID}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, testData[2].ID, result[0].ID)

		// ----------------------------
		// EQUALS ID OR ID
		// ----------------------------
		dbParams := &database.DatabaseParams{
			Where:   squirrel.Or{squirrel.Eq{dao.Table + ".id": testData[1].ID}, squirrel.Eq{dao.Table + ".id": testData[2].ID}},
			OrderBy: []string{"created_at asc"},
		}
		result, err = dao.List(dbParams, nil)
		require.Nil(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, testData[1].ID, result[0].ID)
		assert.Equal(t, testData[2].ID, result[1].ID)

		// ----------------------------
		// ERROR
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Eq{"": ""}}, nil)
		require.ErrorContains(t, err, "syntax error")
		assert.Nil(t, result)
	})

	t.Run("pagination", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(17).Build()

		// ----------------------------
		// Page 1 with 10 items
		// ----------------------------
		p := pagination.New(1, 10)

		result, err := dao.List(&database.DatabaseParams{Pagination: p}, nil)
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, testData[0].ID, result[0].ID)
		assert.Equal(t, testData[9].ID, result[9].ID)

		// ----------------------------
		// Page 2 with 7 items
		// ----------------------------
		p = pagination.New(2, 10)

		result, err = dao.List(&database.DatabaseParams{Pagination: p}, nil)
		require.Nil(t, err)
		require.Len(t, result, 7)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, testData[10].ID, result[0].ID)
		assert.Equal(t, testData[16].ID, result[6].ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table)
		require.Nil(t, err)

		_, err = dao.List(nil, nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourse_Update(t *testing.T) {
	t.Run("card path", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Build()
		require.Empty(t, testData[0].CardPath)

		// Update the card path
		testData[0].CardPath = "/path/to/card.jpg"
		require.Nil(t, dao.Update(testData[0].Course))

		c, err := dao.Get(testData[0].ID, nil)
		require.Nil(t, err)
		assert.Equal(t, testData[0].CardPath, c.CardPath)
	})

	t.Run("available", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Build()
		require.False(t, testData[0].Available)

		// Update the availability
		testData[0].Available = true
		require.Nil(t, dao.Update(testData[0].Course))

		c, err := dao.Get(testData[0].ID, nil)
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

		testData := NewTestBuilder(t).Db(db).Courses(1).Build()
		testData[0].ID = "1234"
		require.Nil(t, dao.Update(testData[0].Course))
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table)
		require.Nil(t, err)

		testData := NewTestBuilder(t).Courses(1).Build()

		err = dao.Update(testData[0].Course)
		require.ErrorContains(t, err, "no such table: "+dao.Table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourse_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Build()

		err := dao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": testData[0].ID}}, nil)
		require.Nil(t, err)
	})

	t.Run("no db params", func(t *testing.T) {
		_, dao, _ := courseSetup(t)

		err := dao.Delete(nil, nil)
		assert.ErrorIs(t, err, ErrMissingWhere)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := courseSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table)
		require.Nil(t, err)

		err = dao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": "1234"}}, nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table)
	})
}
