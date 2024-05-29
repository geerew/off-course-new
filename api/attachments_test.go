package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachments_GetAttachments(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		router := setup(t)

		status, body, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/attachments/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := unmarshalHelper[attachmentResponse](t, body)
		require.Zero(t, int(paginationResp.TotalItems))
		require.Zero(t, len(paginationResp.Items))
	})

	t.Run("200 (found)", func(t *testing.T) {
		router := setup(t)

		daos.NewTestBuilder(t).Db(router.config.DbManager.DataDb).Courses(2).Assets(2).Attachments(2).Build()

		status, body, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/attachments/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentsResp := unmarshalHelper[attachmentResponse](t, body)
		require.Equal(t, 8, int(paginationResp.TotalItems))
		require.Len(t, attachmentsResp, 8)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		router := setup(t)

		testData := daos.NewTestBuilder(t).Db(router.config.DbManager.DataDb).Courses(2).Assets(2).Attachments(4).Build()

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		req := httptest.NewRequest(http.MethodGet, "/api/attachments/?orderBy=created_at%20asc", nil)
		status, body, err := requestHelper(router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentsResp := unmarshalHelper[attachmentResponse](t, body)
		require.Equal(t, 16, int(paginationResp.TotalItems))
		require.Len(t, attachmentsResp, 16)
		require.Equal(t, testData[0].Assets[0].Attachments[0].ID, attachmentsResp[0].ID)
		require.Equal(t, testData[0].Assets[0].Attachments[0].Title, attachmentsResp[0].Title)
		require.Equal(t, testData[0].Assets[0].Attachments[0].Path, attachmentsResp[0].Path)

		// ----------------------------
		// CREATED_AT DESC
		// ----------------------------
		req = httptest.NewRequest(http.MethodGet, "/api/attachments/?orderBy=created_at%20desc", nil)
		status, body, err = requestHelper(router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentsResp = unmarshalHelper[attachmentResponse](t, body)
		require.Equal(t, 16, int(paginationResp.TotalItems))
		require.Len(t, attachmentsResp, 16)
		require.Equal(t, testData[1].Assets[1].Attachments[3].ID, attachmentsResp[0].ID)
		require.Equal(t, testData[1].Assets[1].Attachments[3].Title, attachmentsResp[0].Title)
		require.Equal(t, testData[1].Assets[1].Attachments[3].Path, attachmentsResp[0].Path)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		router := setup(t)

		testData := daos.NewTestBuilder(t).Db(router.config.DbManager.DataDb).Courses(2).Assets(2).Attachments(4).Build()

		// ----------------------------
		// Get the first page (10 attachments)
		// ----------------------------
		params := url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}

		status, body, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/attachments/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentsResp := unmarshalHelper[attachmentResponse](t, body)
		require.Equal(t, 16, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 10)

		// Check the last attachment in the paginated response
		require.Equal(t, testData[1].Assets[0].Attachments[1].ID, attachmentsResp[9].ID)

		// ----------------------------
		// Get the next page (6 attachments)
		// ----------------------------
		params = url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}

		status, body, err = requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/attachments/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentsResp = unmarshalHelper[attachmentResponse](t, body)
		require.Equal(t, 16, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 6)

		// Check the last attachment in the paginated response
		require.Equal(t, testData[1].Assets[1].Attachments[3].ID, attachmentsResp[5].ID)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router := setup(t)

		// Drop the table
		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + daos.NewAttachmentDao(router.config.DbManager.DataDb).Table())
		require.Nil(t, err)

		status, _, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/attachments/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachments_GetAttachment(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		router := setup(t)

		testData := daos.NewTestBuilder(t).Db(router.config.DbManager.DataDb).Courses(2).Assets(2).Attachments(2).Build()

		req := httptest.NewRequest(http.MethodGet, "/api/attachments/"+testData[1].Assets[1].Attachments[0].ID, nil)
		status, body, err := requestHelper(router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var attachmentResp attachmentResponse
		err = json.Unmarshal(body, &attachmentResp)
		require.Nil(t, err)
		require.Equal(t, testData[1].Assets[1].Attachments[0].ID, attachmentResp.ID)
		require.Equal(t, testData[1].Assets[1].Attachments[0].Title, attachmentResp.Title)
		require.Equal(t, testData[1].Assets[1].Attachments[0].Path, attachmentResp.Path)
	})

	t.Run("404 (not found)", func(t *testing.T) {
		router := setup(t)

		status, _, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/attachments/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router := setup(t)

		// Drop the table
		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + daos.NewAttachmentDao(router.config.DbManager.DataDb).Table())
		require.Nil(t, err)

		status, _, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/attachments/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachments_ServeAttachment(t *testing.T) {
	t.Run("200 (ok)", func(t *testing.T) {
		router := setup(t)

		testData := daos.NewTestBuilder(t).Db(router.config.DbManager.DataDb).Courses(2).Assets(2).Attachments(2).Build()

		// Create attachment file
		require.Nil(t, router.config.AppFs.Fs.MkdirAll(filepath.Dir(testData[1].Assets[0].Attachments[0].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(router.config.AppFs.Fs, testData[1].Assets[0].Attachments[0].Path, []byte("hello"), os.ModePerm))

		status, body, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/attachments/"+testData[1].Assets[0].Attachments[0].ID+"/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, "hello", string(body))
	})

	t.Run("400 (invalid path)", func(t *testing.T) {
		router := setup(t)

		testData := daos.NewTestBuilder(t).Db(router.config.DbManager.DataDb).Courses(2).Assets(2).Attachments(2).Build()

		status, body, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/attachments/"+testData[1].Assets[0].Attachments[0].ID+"/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Attachment does not exist")
	})

	t.Run("404 (not found)", func(t *testing.T) {
		router := setup(t)

		status, _, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/attachments/test/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router := setup(t)

		// Drop the table
		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + daos.NewAttachmentDao(router.config.DbManager.DataDb).Table())
		require.Nil(t, err)

		status, _, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/attachments/test/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}
