package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/geerew/off-course/models"
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

		f := fiber.New()
		bindAttachmentsApi(f.Group("/api"), appFs, db)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 0, int(respData.TotalItems))
		assert.Len(t, respData.Items, 0)
	})

	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAttachmentsApi(f.Group("/api"), appFs, db)

		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 2)
		attachments := models.NewTestAttachments(t, db, assets, 2)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 8, int(respData.TotalItems))

		// Unmarshal
		var attachmentsResp []attachmentResponse
		for _, item := range respData.Items {
			var attachment attachmentResponse
			require.Nil(t, json.Unmarshal(item, &attachment))
			attachmentsResp = append(attachmentsResp, attachment)
		}

		assert.Equal(t, attachments[7].ID, attachmentsResp[0].ID)
		assert.Equal(t, attachments[7].Title, attachmentsResp[0].Title)
		assert.Equal(t, attachments[7].Path, attachmentsResp[0].Path)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAttachmentsApi(f.Group("/api"), appFs, db)

		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 2)
		attachments := models.NewTestAttachments(t, db, assets, 2)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/?orderBy=created_at%20asc", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 8, int(respData.TotalItems))

		// Unmarshal
		var attachmentsResp []attachmentResponse
		for _, item := range respData.Items {
			var attachment attachmentResponse
			require.Nil(t, json.Unmarshal(item, &attachment))
			attachmentsResp = append(attachmentsResp, attachment)
		}

		assert.Equal(t, attachments[0].ID, attachmentsResp[0].ID)
		assert.Equal(t, attachments[0].Title, attachmentsResp[0].Title)
		assert.Equal(t, attachments[0].Path, attachmentsResp[0].Path)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAttachmentsApi(f.Group("/api"), appFs, db)

		courses := models.NewTestCourses(t, db, 1)
		assets := models.NewTestAssets(t, db, courses, 2)
		attachments := models.NewTestAttachments(t, db, assets, 8)

		// Get the first 10 courses
		params := url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}
		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/?"+params.Encode(), nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 16, int(respData.TotalItems))
		assert.Len(t, respData.Items, 10)

		// Unmarshal
		var attachmentsResp []attachmentResponse
		for _, item := range respData.Items {
			var attachment attachmentResponse
			require.Nil(t, json.Unmarshal(item, &attachment))
			attachmentsResp = append(attachmentsResp, attachment)
		}

		// Assert the last attachment in the paginated response
		assert.Equal(t, attachments[9].ID, attachmentsResp[9].ID)
		assert.Equal(t, attachments[9].Title, attachmentsResp[9].Title)
		assert.Equal(t, attachments[9].Path, attachmentsResp[9].Path)

		// Get the next 8 courses
		params = url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}
		resp, err = f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/?"+params.Encode(), nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ = io.ReadAll(resp.Body)

		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 16, int(respData.TotalItems))
		assert.Len(t, respData.Items, 6)

		// Unmarshal
		attachmentsResp = []attachmentResponse{}
		for _, item := range respData.Items {
			var attachment attachmentResponse
			require.Nil(t, json.Unmarshal(item, &attachment))
			attachmentsResp = append(attachmentsResp, attachment)
		}

		// Assert the last attachment in the paginated response
		assert.Equal(t, attachments[15].ID, attachmentsResp[5].ID)
		assert.Equal(t, attachments[15].Title, attachmentsResp[5].Title)
		assert.Equal(t, attachments[15].Path, attachmentsResp[5].Path)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAttachmentsApi(f.Group("/api"), appFs, db)

		// Drop the attachments table
		_, err := db.DB().NewDropTable().Model(&models.Attachment{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachments_GetAttachment(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAttachmentsApi(f.Group("/api"), appFs, db)

		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)
		attachments := models.NewTestAttachments(t, db, assets, 2)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/"+attachments[6].ID, nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

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

		f := fiber.New()
		bindAttachmentsApi(f.Group("/api"), appFs, db)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/test", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAttachmentsApi(f.Group("/api"), appFs, db)

		// Drop the table
		_, err := db.DB().NewDropTable().Model(&models.Attachment{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/test", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachments_DownloadAttachment(t *testing.T) {
	t.Run("200 (ok)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAttachmentsApi(f.Group("/api"), appFs, db)

		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)
		attachments := models.NewTestAttachments(t, db, assets, 2)

		// Create the attachment path
		require.Nil(t, appFs.Fs.MkdirAll(filepath.Dir(attachments[2].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(appFs.Fs, attachments[2].Path, []byte("hello"), os.ModePerm))

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/"+attachments[2].ID+"/download", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "hello", string(body))
	})

	t.Run("404 (not found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAttachmentsApi(f.Group("/api"), appFs, db)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/test/download", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAttachmentsApi(f.Group("/api"), appFs, db)

		// Drop the table
		_, err := db.DB().NewDropTable().Model(&models.Attachment{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/test/download", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("400 (invalid path)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAttachmentsApi(f.Group("/api"), appFs, db)

		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)
		attachments := models.NewTestAttachments(t, db, assets, 2)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/"+attachments[6].ID+"/download", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "attachment does not exist")
	})
}
