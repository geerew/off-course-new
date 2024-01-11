package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachments_GetAttachments(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		status, body, err := attachmentsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/attachments/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 0, int(respData.TotalItems))
		assert.Len(t, respData.Items, 0)
	})

	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Create 2 courses, 2 assets, and 2 attachments (8 attachments total)
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 2)
		attachments := models.NewTestAttachments(t, db, assets, 2)

		status, body, err := attachmentsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/attachments/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		_, attachmentsResp := attachmentsUnmarshalHelper(t, body)
		require.Len(t, attachmentsResp, 8)
		assert.Equal(t, attachments[7].ID, attachmentsResp[0].ID)
		assert.Equal(t, attachments[7].Title, attachmentsResp[0].Title)
		assert.Equal(t, attachments[7].Path, attachmentsResp[0].Path)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Create 2 courses, 2 assets, and 2 attachments (8 attachments total)
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 2)
		attachments := models.NewTestAttachments(t, db, assets, 2)

		req := httptest.NewRequest(http.MethodGet, "/api/attachments/?orderBy=created_at%20asc", nil)
		status, body, err := attachmentsRequestHelper(t, appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		_, attachmentsResp := attachmentsUnmarshalHelper(t, body)
		require.Len(t, attachmentsResp, 8)
		assert.Equal(t, attachments[0].ID, attachmentsResp[0].ID)
		assert.Equal(t, attachments[0].Title, attachmentsResp[0].Title)
		assert.Equal(t, attachments[0].Path, attachmentsResp[0].Path)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Create 1 course, 2 assets, and 8 attachments (16 attachments total)
		courses := models.NewTestCourses(t, db, 1)
		assets := models.NewTestAssets(t, db, courses, 2)
		attachments := models.NewTestAttachments(t, db, assets, 8)

		// ------------------------
		// Get the first page (10 attachments)
		// ------------------------
		params := url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}

		req := httptest.NewRequest(http.MethodGet, "/api/attachments/?"+params.Encode(), nil)

		status, body, err := attachmentsRequestHelper(t, appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentsResp := attachmentsUnmarshalHelper(t, body)
		assert.Equal(t, 16, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 10)

		// Check the last attachment in the paginated response
		assert.Equal(t, attachments[9].ID, attachmentsResp[9].ID)
		assert.Equal(t, attachments[9].Title, attachmentsResp[9].Title)
		assert.Equal(t, attachments[9].Path, attachmentsResp[9].Path)

		// ------------------------
		// Get the next page (6 attachments)
		// ------------------------
		params = url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}

		req = httptest.NewRequest(http.MethodGet, "/api/attachments/?"+params.Encode(), nil)

		status, body, err = attachmentsRequestHelper(t, appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentsResp = attachmentsUnmarshalHelper(t, body)
		assert.Equal(t, 16, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 6)

		// Check the last attachment in the paginated response
		assert.Equal(t, attachments[15].ID, attachmentsResp[5].ID)
		assert.Equal(t, attachments[15].Title, attachmentsResp[5].Title)
		assert.Equal(t, attachments[15].Path, attachmentsResp[5].Path)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Drop the table
		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableAttachments())
		require.Nil(t, err)

		status, _, err := attachmentsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/attachments/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachments_GetAttachment(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Create 2 courses, 2 assets, and 2 attachments (8 attachments total)
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 2)
		attachments := models.NewTestAttachments(t, db, assets, 2)

		status, body, err := attachmentsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/attachments/"+attachments[6].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var respData attachmentResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, attachments[6].ID, respData.ID)
		assert.Equal(t, attachments[6].Title, respData.Title)
		assert.Equal(t, attachments[6].Path, respData.Path)
	})

	t.Run("404 (not found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		status, _, err := attachmentsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/attachments/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Drop the table
		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableAttachments())
		require.Nil(t, err)

		status, _, err := attachmentsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/attachments/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachments_ServeAttachment(t *testing.T) {
	t.Run("200 (ok)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Create 2 courses, 2 assets, and 2 attachments (8 attachments total)
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 2)
		attachments := models.NewTestAttachments(t, db, assets, 2)

		// Create attachment file
		require.Nil(t, appFs.Fs.MkdirAll(filepath.Dir(attachments[2].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(appFs.Fs, attachments[2].Path, []byte("hello"), os.ModePerm))

		status, body, err := attachmentsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/attachments/"+attachments[2].ID+"/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "hello", string(body))
	})

	t.Run("400 (invalid path)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Create 2 courses, 2 assets, and 2 attachments (8 attachments total)
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 2)
		attachments := models.NewTestAttachments(t, db, assets, 2)

		status, body, err := attachmentsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/attachments/"+attachments[2].ID+"/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		assert.Contains(t, string(body), "attachment does not exist")
	})

	t.Run("404 (not found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		status, _, err := attachmentsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/attachments/test/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Drop the table
		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableAttachments())
		require.Nil(t, err)

		status, _, err := attachmentsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/attachments/test/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPERS
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func attachmentsRequestHelper(t *testing.T, appFs *appFs.AppFs, db database.Database, req *http.Request) (int, []byte, error) {
	f := fiber.New()
	bindAttachmentsApi(f.Group("/api"), appFs, db)

	resp, err := f.Test(req)
	if err != nil {
		return -1, nil, err
	}

	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, b, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func attachmentsUnmarshalHelper(t *testing.T, body []byte) (pagination.PaginationResult, []attachmentResponse) {
	var respData pagination.PaginationResult
	err := json.Unmarshal(body, &respData)
	require.Nil(t, err)

	var attachmentsResp []attachmentResponse
	for _, item := range respData.Items {
		var attachment attachmentResponse
		require.Nil(t, json.Unmarshal(item, &attachment))
		attachmentsResp = append(attachmentsResp, attachment)
	}

	return respData, attachmentsResp
}
