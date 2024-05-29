package daos

import (
	"fmt"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func logSetup(t *testing.T) (*LogDao, database.Database) {
	dbManager := setup(t)
	logDao := NewLogDao(dbManager.LogsDb)
	return logDao, dbManager.LogsDb
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestLog_Count(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		dao, _ := logSetup(t)

		count, err := dao.Count(nil)
		require.Nil(t, err)
		require.Zero(t, count)
	})

	t.Run("entries", func(t *testing.T) {
		dao, _ := logSetup(t)

		for i := range 2 {
			require.Nil(t, dao.Write(&models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", i+1)}, nil))
		}

		count, err := dao.Count(nil)
		require.Nil(t, err)
		require.Equal(t, count, 2)
	})

	t.Run("where", func(t *testing.T) {
		dao, _ := logSetup(t)

		for i := range 2 {
			require.Nil(t, dao.Write(&models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", i+1)}, nil))
		}

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		count, err := dao.Count(&database.DatabaseParams{Where: squirrel.Eq{dao.Table() + ".message": "log 1"}})
		require.Nil(t, err)
		require.Equal(t, 1, count)

		// ----------------------------
		// NOT EQUALS ID
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.NotEq{dao.Table() + ".message": "log 1"}})
		require.Nil(t, err)
		require.Equal(t, 1, count)

		// ----------------------------
		// ERROR
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		require.Zero(t, count)
	})

	t.Run("db error", func(t *testing.T) {
		dao, db := logSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
		require.Nil(t, err)

		_, err = dao.Count(nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestLog_Write(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, _ := logSetup(t)

		for i := range 2 {
			require.Nil(t, dao.Write(&models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", i+1)}, nil))
		}

		// newC, err := dao.Get(testData[0].ID, nil, nil)
		// require.Nil(t, err)
		// require.NotEmpty(t, newC.ID)
		// require.Equal(t, testData[0].Title, newC.Title)
		// require.Equal(t, testData[0].Path, newC.Path)
		// require.Empty(t, newC.CardPath)
		// require.False(t, newC.Available)
		// require.False(t, newC.CreatedAt.IsZero())
		// require.False(t, newC.UpdatedAt.IsZero())
		// //Scan status
		// require.Empty(t, newC.ScanStatus)
		// // Progress
		// require.False(t, newC.Started)
		// require.True(t, newC.StartedAt.IsZero())
		// require.Zero(t, newC.Percent)
		// require.True(t, newC.CompletedAt.IsZero())
	})

	t.Run("constraints", func(t *testing.T) {
		dao, _ := logSetup(t)

		// No message
		log := &models.Log{}
		require.ErrorContains(t, dao.Write(log, nil), fmt.Sprintf("NOT NULL constraint failed: %s.message", dao.Table()))
		log.Message = ""
		require.ErrorContains(t, dao.Write(log, nil), fmt.Sprintf("NOT NULL constraint failed: %s.message", dao.Table()))
		log.Message = "Log 1"

		// Success
		require.Nil(t, dao.Write(log, nil))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestLog_Get(t *testing.T) {
// 	t.Run("found", func(t *testing.T) {
// 		dao, db := logSetup(t)

// 		testData := NewTestBuilder(t).Db(db).Courses(2).Assets(1).Build()

// 		c, err := dao.Get(testData[1].ID, nil, nil)
// 		require.Nil(t, err)
// 		require.Equal(t, testData[1].ID, c.ID)
// 		require.Empty(t, testData[1].ScanStatus)

// 		// ----------------------------
// 		// scan
// 		// ----------------------------
// 		scanDao := NewScanDao(db)
// 		require.Nil(t, scanDao.Create(&models.Scan{CourseID: testData[1].ID}))

// 		c, err = dao.Get(testData[1].ID, nil, nil)
// 		require.Nil(t, err)
// 		require.Equal(t, string(types.ScanStatusWaiting), c.ScanStatus)

// 		// ----------------------------
// 		// Availability
// 		// ----------------------------
// 		require.False(t, c.Available)

// 		// Set to started
// 		testData[1].Available = true
// 		require.Nil(t, dao.Update(testData[1].Course))

// 		c, err = dao.Get(testData[1].ID, nil, nil)
// 		require.Nil(t, err)
// 		require.True(t, c.Available)

// 		// ----------------------------
// 		// Progress
// 		// ----------------------------
// 		require.False(t, c.Started)
// 		require.True(t, c.StartedAt.IsZero())
// 		require.Zero(t, c.Percent)
// 		require.True(t, c.CompletedAt.IsZero())

// 		// Get course progress
// 		// cpDao := NewCourseProgressDao(db)
// 		// origCp, err := cpDao.Get(testData[1].ID, nil)
// 		// require.Nil(t, err)

// 		// Update
// 		// err = cpDao.Update(origCp)
// 		// require.Nil(t, err)

// 		// updatedCp, err :=  cpDao.Get(testData[1].ID, nil)
// 		// require.Nil(t, err)
// 		// require.True(t, updatedCp.Started)
// 		// require.False(t, updatedCp.StartedAt.IsZero())

// 		// 		// Set to started
// 		// 		_, err = UpdateCourseProgressStarted(db, testData[1].ID, true)
// 		// 		require.Nil(t, err)

// 		// 		c, err = GetCourse(db, testData[1].ID)
// 		// 		require.Nil(t, err)
// 		// 		require.True(t, c.Started)
// 		// 		require.False(t, c.StartedAt.IsZero())
// 		// 		require.Zero(t, c.Percent)
// 		// 		require.True(t, c.CompletedAt.IsZero())

// 		// 		// Mark asset as completed (only 1 asset so the course will be 100%)
// 		// 		_, err = UpdateAssetProgressCompleted(db, testData[1].Assets[0].ID, true)
// 		// 		require.Nil(t, err)

// 		// 		c, err = GetCourse(db, testData[1].ID)
// 		// 		require.Nil(t, err)
// 		// 		require.True(t, c.Started)
// 		// 		require.False(t, c.StartedAt.IsZero())
// 		// 		require.Equal(t, 100, c.Percent)
// 		// 		require.False(t, c.CompletedAt.IsZero())
// 	})

// 	t.Run("not found", func(t *testing.T) {
// 		dao, _ := logSetup(t)

// 		c, err := dao.Get("1234", nil, nil)
// 		require.ErrorIs(t, err, sql.ErrNoRows)
// 		require.Nil(t, c)
// 	})

// 	t.Run("empty id", func(t *testing.T) {
// 		dao, _ := logSetup(t)

// 		c, err := dao.Get("", nil, nil)
// 		require.ErrorIs(t, err, sql.ErrNoRows)
// 		require.Nil(t, c)
// 	})

// 	t.Run("db error", func(t *testing.T) {
// 		dao, db := logSetup(t)

// 		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
// 		require.Nil(t, err)

// 		_, err = dao.Get("1234", nil, nil)
// 		require.ErrorContains(t, err, "no such table: "+dao.Table())
// 	})
// }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestLog_List(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		dao, _ := logSetup(t)

		courses, err := dao.List(nil, nil)
		require.Nil(t, err)
		require.Zero(t, courses)
	})

	t.Run("found", func(t *testing.T) {
		dao, _ := logSetup(t)

		for i := range 5 {
			require.Nil(t, dao.Write(&models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", i+1)}, nil))
			time.Sleep(1 * time.Millisecond)
		}

		result, err := dao.List(nil, nil)
		require.Nil(t, err)
		require.Len(t, result, 5)
		require.Equal(t, "log 5", result[0].Message)
		require.Equal(t, "log 1", result[4].Message)
	})

	t.Run("where", func(t *testing.T) {
		dao, _ := logSetup(t)

		for i := range 5 {
			require.Nil(t, dao.Write(&models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", i+1)}, nil))
			time.Sleep(1 * time.Millisecond)
		}

		// ----------------------------
		// EQUALS log 2 or log 3
		// ----------------------------
		result, err := dao.List(
			&database.DatabaseParams{Where: squirrel.Or{
				squirrel.Eq{dao.Table() + ".message": "log 2"},
				squirrel.Eq{dao.Table() + ".message": "log 3"}}},
			nil)
		require.Nil(t, err)
		require.Len(t, result, 2)
		require.Equal(t, "log 3", result[0].Message)
		require.Equal(t, "log 2", result[1].Message)

		// ----------------------------
		// ERROR
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Eq{"": ""}}, nil)
		require.ErrorContains(t, err, "syntax error")
		require.Nil(t, result)
	})

	t.Run("pagination", func(t *testing.T) {
		dao, _ := logSetup(t)

		for i := range 17 {
			require.Nil(t, dao.Write(&models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", i+1)}, nil))
			time.Sleep(1 * time.Millisecond)
		}

		// ----------------------------
		// Page 1 with 10 items
		// ----------------------------
		p := pagination.New(1, 10)

		result, err := dao.List(&database.DatabaseParams{Pagination: p}, nil)
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 17, p.TotalItems())
		require.Equal(t, "log 17", result[0].Message)
		require.Equal(t, "log 8", result[9].Message)

		// ----------------------------
		// Page 2 with 7 items
		// ----------------------------
		p = pagination.New(2, 10)

		result, err = dao.List(&database.DatabaseParams{Pagination: p}, nil)
		require.Nil(t, err)
		require.Len(t, result, 7)
		require.Equal(t, 17, p.TotalItems())
		require.Equal(t, "log 7", result[0].Message)
		require.Equal(t, "log 1", result[6].Message)
	})

	t.Run("db error", func(t *testing.T) {
		dao, db := logSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
		require.Nil(t, err)

		_, err = dao.List(nil, nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestLog_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, _ := logSetup(t)

		for i := range 3 {
			require.Nil(t, dao.Write(&models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", i+1)}, nil))
			time.Sleep(1 * time.Millisecond)
		}

		err := dao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"message": "log 2"}}, nil)
		require.Nil(t, err)
	})

	t.Run("no db params", func(t *testing.T) {
		dao, _ := logSetup(t)

		err := dao.Delete(nil, nil)
		require.ErrorIs(t, err, ErrMissingWhere)
	})

	t.Run("db error", func(t *testing.T) {
		dao, db := logSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
		require.Nil(t, err)

		err = dao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": "1234"}}, nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table())
	})
}
