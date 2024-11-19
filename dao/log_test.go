package dao

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setupLog(tb testing.TB) (*DAO, context.Context) {
	tb.Helper()

	// Logger
	var logs []*logger.Log
	var logsMux sync.Mutex
	logger, _, err := logger.InitLogger(&logger.BatchOptions{
		BatchSize: 1,
		WriteFn:   logger.TestWriteFn(&logs, &logsMux),
	})
	require.NoError(tb, err, "Failed to initialize logger")

	// Filesystem
	appFs := appFs.NewAppFs(afero.NewMemMapFs(), logger)

	// DB
	dbManager, err := database.NewSqliteDBManager(&database.DatabaseConfig{
		IsDebug:  false,
		DataDir:  "./oc_data",
		AppFs:    appFs,
		InMemory: true,
	})

	require.NoError(tb, err)
	require.NotNil(tb, dbManager)

	dao := &DAO{db: dbManager.LogsDb}

	return dao, context.Background()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_WriteLog(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setupLog(t)
		log := &models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", 1)}
		require.NoError(t, dao.WriteLog(ctx, log))
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setupLog(t)
		require.ErrorIs(t, dao.WriteLog(ctx, nil), utils.ErrNilPtr)
	})

}

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestLog_List(t *testing.T) {
// 	t.Run("no entries", func(t *testing.T) {
// 		dao, _ := logSetup(t)

// 		courses, err := dao.List(nil, nil)
// 		require.Nil(t, err)
// 		require.Zero(t, courses)
// 	})

// 	t.Run("found", func(t *testing.T) {
// 		dao, _ := logSetup(t)

// 		for i := range 5 {
// 			require.Nil(t, dao.Write(&models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", i+1)}, nil))
// 			time.Sleep(1 * time.Millisecond)
// 		}

// 		result, err := dao.List(nil, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 5)
// 		require.Equal(t, "log 5", result[0].Message)
// 		require.Equal(t, "log 1", result[4].Message)
// 	})

// 	t.Run("where", func(t *testing.T) {
// 		dao, _ := logSetup(t)

// 		for i := range 5 {
// 			require.Nil(t, dao.Write(&models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", i+1)}, nil))
// 			time.Sleep(1 * time.Millisecond)
// 		}

// 		// ----------------------------
// 		// EQUALS log 2 or log 3
// 		// ----------------------------
// 		result, err := dao.List(
// 			&database.DatabaseParams{Where: squirrel.Or{
// 				squirrel.Eq{dao.Table() + ".message": "log 2"},
// 				squirrel.Eq{dao.Table() + ".message": "log 3"}}},
// 			nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 2)
// 		require.Equal(t, "log 3", result[0].Message)
// 		require.Equal(t, "log 2", result[1].Message)

// 		// ----------------------------
// 		// ERROR
// 		// ----------------------------
// 		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Eq{"": ""}}, nil)
// 		require.ErrorContains(t, err, "syntax error")
// 		require.Nil(t, result)
// 	})

// 	t.Run("pagination", func(t *testing.T) {
// 		dao, _ := logSetup(t)

// 		for i := range 17 {
// 			require.Nil(t, dao.Write(&models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", i+1)}, nil))
// 			time.Sleep(1 * time.Millisecond)
// 		}

// 		// ----------------------------
// 		// Page 1 with 10 items
// 		// ----------------------------
// 		p := pagination.New(1, 10)

// 		result, err := dao.List(&database.DatabaseParams{Pagination: p}, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 10)
// 		require.Equal(t, 17, p.TotalItems())
// 		require.Equal(t, "log 17", result[0].Message)
// 		require.Equal(t, "log 8", result[9].Message)

// 		// ----------------------------
// 		// Page 2 with 7 items
// 		// ----------------------------
// 		p = pagination.New(2, 10)

// 		result, err = dao.List(&database.DatabaseParams{Pagination: p}, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 7)
// 		require.Equal(t, 17, p.TotalItems())
// 		require.Equal(t, "log 7", result[0].Message)
// 		require.Equal(t, "log 1", result[6].Message)
// 	})

// 	t.Run("db error", func(t *testing.T) {
// 		dao, db := logSetup(t)

// 		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
// 		require.Nil(t, err)

// 		_, err = dao.List(nil, nil)
// 		require.ErrorContains(t, err, "no such table: "+dao.Table())
// 	})
// }
