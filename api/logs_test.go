package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestLogs_GetLogs(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		router := setup(t)

		status, body, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/logs/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := unmarshalHelper[logResponse](t, body)
		require.Zero(t, int(paginationResp.TotalItems))
		require.Zero(t, len(paginationResp.Items))
	})

	t.Run("200 (found)", func(t *testing.T) {
		router := setup(t)

		logDao := daos.NewLogDao(router.config.DbManager.LogsDb)
		for i := range 5 {
			require.Nil(t, logDao.Write(&models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", i+1)}, nil))
		}

		status, body, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/logs/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses := unmarshalHelper[logResponse](t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, logResponses, 5)
	})

	t.Run("200 (min level)", func(t *testing.T) {
		router := setup(t)

		logDao := daos.NewLogDao(router.config.DbManager.LogsDb)

		require.Nil(t, logDao.Write(&models.Log{Data: map[string]any{}, Level: -4, Message: "debug log"}, nil))
		require.Nil(t, logDao.Write(&models.Log{Data: map[string]any{}, Level: 0, Message: "info log"}, nil))
		require.Nil(t, logDao.Write(&models.Log{Data: map[string]any{}, Level: 4, Message: "warn log"}, nil))
		require.Nil(t, logDao.Write(&models.Log{Data: map[string]any{}, Level: 8, Message: "error log"}, nil))

		// ----------------------------
		// All
		// ----------------------------
		status, body, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/logs/?level=-4", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses := unmarshalHelper[logResponse](t, body)
		require.Equal(t, 4, int(paginationResp.TotalItems))
		require.Len(t, logResponses, 4)

		// ----------------------------
		// Info and above
		// ----------------------------
		status, body, err = requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/logs/?level=0", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses = unmarshalHelper[logResponse](t, body)
		require.Equal(t, 3, int(paginationResp.TotalItems))
		require.Len(t, logResponses, 3)

		// ----------------------------
		// Warn and above
		// ----------------------------
		status, body, err = requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/logs/?level=4", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses = unmarshalHelper[logResponse](t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, logResponses, 2)

		// ----------------------------
		// Error only
		// ----------------------------
		status, body, err = requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/logs/?level=8", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses = unmarshalHelper[logResponse](t, body)
		require.Equal(t, 1, int(paginationResp.TotalItems))
		require.Len(t, logResponses, 1)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		router := setup(t)

		logDao := daos.NewLogDao(router.config.DbManager.LogsDb)
		for i := range 17 {
			require.Nil(t, logDao.Write(&models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", i+1)}, nil))
			time.Sleep(1 * time.Millisecond)
		}

		// ----------------------------
		// Get the first page (10 logs)
		// ----------------------------
		params := url.Values{
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}

		status, body, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/logs/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses := unmarshalHelper[logResponse](t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 10)

		require.Equal(t, "log 17", logResponses[0].Message)
		require.Equal(t, "log 8", logResponses[9].Message)

		// ----------------------------
		// Get the second page (7 logs)
		// ----------------------------
		params = url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}
		status, body, err = requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/logs/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses = unmarshalHelper[logResponse](t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 7)

		require.Equal(t, "log 7", logResponses[0].Message)
		require.Equal(t, "log 1", logResponses[6].Message)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router := setup(t)

		// Drop the courses table
		_, err := router.config.DbManager.LogsDb.Exec("DROP TABLE IF EXISTS " + daos.NewLogDao(router.config.DbManager.LogsDb).Table())
		require.Nil(t, err)

		status, _, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/logs/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}
