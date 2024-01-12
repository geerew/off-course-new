package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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

func TestAssets_GetAssets(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		status, body, err := assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := assetsUnmarshalHelper(t, body)
		assert.Zero(t, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 0)
	})

	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Create 2 courses, 5 assets, 2 attachments (10 assets and 20 attachments total)
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)
		models.NewTestAttachments(t, db, assets, 2)

		status, body, err := assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := assetsUnmarshalHelper(t, body)
		require.Equal(t, 10, int(paginationResp.TotalItems))
		require.Len(t, assetsResp, 10)
		assert.Equal(t, assets[9].ID, assetsResp[0].ID)
		assert.Equal(t, assets[9].Title, assetsResp[0].Title)
		assert.Equal(t, assets[9].Path, assetsResp[0].Path)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Create 2 courses, 5 assets (10 assets total)
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)

		status, body, err := assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/?orderBy=created_at%20asc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := assetsUnmarshalHelper(t, body)
		require.Equal(t, 10, int(paginationResp.TotalItems))
		require.Len(t, assetsResp, 10)
		assert.Equal(t, assets[0].ID, assetsResp[0].ID)
		assert.Equal(t, assets[0].Title, assetsResp[0].Title)
		assert.Equal(t, assets[0].Path, assetsResp[0].Path)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Create 3 courses, 6 assets (18 assets total)
		courses := models.NewTestCourses(t, db, 3)
		assets := models.NewTestAssets(t, db, courses, 6)

		// ------------------------
		// Get the first page (10 assets)
		// ------------------------
		params := url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}

		status, body, err := assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := assetsUnmarshalHelper(t, body)
		assert.Equal(t, 18, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 10)

		// Check the last asset in the paginated response
		assert.Equal(t, assets[9].ID, assetsResp[9].ID)
		assert.Equal(t, assets[9].Title, assetsResp[9].Title)
		assert.Equal(t, assets[9].Path, assetsResp[9].Path)

		// ------------------------
		// Get the second page (8 assets)
		// ------------------------
		params = url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}

		status, body, err = assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp = assetsUnmarshalHelper(t, body)
		assert.Equal(t, 18, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 8)

		// Check the last asset in the paginated response
		assert.Equal(t, assets[17].ID, assetsResp[7].ID)
		assert.Equal(t, assets[17].Title, assetsResp[7].Title)
		assert.Equal(t, assets[17].Path, assetsResp[7].Path)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Drop the table
		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableAssets())
		require.Nil(t, err)

		status, _, err := assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssets_GetAsset(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Create 2 courses, 5 assets, 2 attachments
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)
		models.NewTestAttachments(t, db, assets, 2)

		status, body, err := assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/"+assets[6].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp assetResponse
		err = json.Unmarshal(body, &assetResp)
		require.Nil(t, err)
		assert.Equal(t, assets[6].ID, assetResp.ID)
		assert.Equal(t, assets[6].Title, assetResp.Title)
		assert.Equal(t, assets[6].Path, assetResp.Path)
		assert.Equal(t, assets[6].CourseID, assetResp.CourseID)
		assert.Len(t, assetResp.Attachments, 2)
	})

	t.Run("404 (not found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		status, _, err := assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Drop the table
		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableAssets())
		require.Nil(t, err)

		status, _, err := assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssets_UpdateAsset(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		// Create 1 courses, 1 asset
		course := models.NewTestCourses(t, db, 1)[0]
		asset := models.NewTestAssets(t, db, []*models.Course{course}, 1)[0]

		// ------------------------
		// Update the video position
		// ------------------------
		asset.VideoPos = 45

		// Marshal the asset for the request
		data, err := json.Marshal(assetResponseHelper([]*models.Asset{asset})[0])
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/assets/"+asset.ID, strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		status, body, err := assetsRequestHelper(t, appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp1 assetResponse
		err = json.Unmarshal(body, &assetResp1)
		require.Nil(t, err)
		assert.Equal(t, asset.ID, assetResp1.ID)
		assert.Equal(t, 45, assetResp1.VideoPos)
		assert.False(t, assetResp1.Completed)
		assert.True(t, assetResp1.CompletedAt.IsZero())

		// ------------------------
		// Set completed to true
		// ------------------------
		assetResp1.Completed = true

		// Marshal the asset for the request
		data, err = json.Marshal(assetResp1)
		require.Nil(t, err)

		req = httptest.NewRequest(http.MethodPut, "/api/assets/"+asset.ID, strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		status, body, err = assetsRequestHelper(t, appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp2 assetResponse
		err = json.Unmarshal(body, &assetResp2)
		require.Nil(t, err)
		assert.Equal(t, asset.ID, assetResp2.ID)
		assert.Equal(t, 45, assetResp2.VideoPos)
		assert.True(t, assetResp2.Completed)
		assert.False(t, assetResp2.CompletedAt.IsZero())

		// ------------------------
		// Set video position to 10 completed to false
		// ------------------------
		assetResp2.VideoPos = 10
		assetResp2.Completed = false

		// Marshal the asset for the request
		data, err = json.Marshal(assetResp2)
		require.Nil(t, err)

		req = httptest.NewRequest(http.MethodPut, "/api/assets/"+asset.ID, strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		status, body, err = assetsRequestHelper(t, appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp3 assetResponse
		err = json.Unmarshal(body, &assetResp3)
		require.Nil(t, err)
		assert.Equal(t, asset.ID, assetResp3.ID)
		assert.Equal(t, 10, assetResp3.VideoPos)
		assert.False(t, assetResp3.Completed)
		assert.True(t, assetResp3.CompletedAt.IsZero())
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		req := httptest.NewRequest(http.MethodPut, "/api/assets/test", strings.NewReader(`bob`))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := assetsRequestHelper(t, appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("404 (not found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		req := httptest.NewRequest(http.MethodPut, "/api/assets/test", strings.NewReader(`{"id": "1234567"}`))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := assetsRequestHelper(t, appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		// Drop the table
		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableAssets())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/assets/test", strings.NewReader(`{"id": "1234567"}`))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := assetsRequestHelper(t, appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssets_ServeAsset(t *testing.T) {
	t.Run("200 (full video)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Create 1 course and 2 assets (2 assets total)
		courses := models.NewTestCourses(t, db, 1)
		assets := models.NewTestAssets(t, db, courses, 2)

		// Create asset
		require.Nil(t, appFs.Fs.MkdirAll(filepath.Dir(assets[1].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(appFs.Fs, assets[1].Path, []byte("video"), os.ModePerm))

		status, body, err := assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/"+assets[1].ID+"/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "video", string(body))
	})

	t.Run("200 (stream video)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Create 1 course and 2 assets (2 assets total)
		courses := models.NewTestCourses(t, db, 1)
		assets := models.NewTestAssets(t, db, courses, 2)

		// Create asset
		require.Nil(t, appFs.Fs.MkdirAll(filepath.Dir(assets[1].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(appFs.Fs, assets[1].Path, []byte("video"), os.ModePerm))

		req := httptest.NewRequest(http.MethodGet, "/api/assets/"+assets[1].ID+"/serve", nil)
		req.Header.Add("Range", "bytes=0-")

		status, body, err := assetsRequestHelper(t, appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusPartialContent, status)
		assert.Equal(t, "video", string(body))
	})

	t.Run("400 (invalid path)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Create 1 course and 2 assets (2 assets total)
		courses := models.NewTestCourses(t, db, 1)
		assets := models.NewTestAssets(t, db, courses, 2)

		status, body, err := assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/"+assets[1].ID+"/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		assert.Contains(t, string(body), "asset does not exist")
	})

	t.Run("400 (invalid video range)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Create 1 course and 2 assets (2 assets total)
		courses := models.NewTestCourses(t, db, 1)
		assets := models.NewTestAssets(t, db, courses, 2)

		// Create asset
		require.Nil(t, appFs.Fs.MkdirAll(filepath.Dir(assets[1].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(appFs.Fs, assets[1].Path, []byte("video"), os.ModePerm))

		req := httptest.NewRequest(http.MethodGet, "/api/assets/"+assets[1].ID+"/serve", nil)
		req.Header.Add("Range", "bytes=10-1")

		status, body, err := assetsRequestHelper(t, appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		assert.Contains(t, string(body), "range start cannot be greater than end")
	})

	t.Run("404 (not found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		status, _, err := assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/test/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableAssets())
		require.Nil(t, err)

		status, _, err := assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/test/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPERS
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func assetsRequestHelper(t *testing.T, appFs *appFs.AppFs, db database.Database, req *http.Request) (int, []byte, error) {
	f := fiber.New()
	bindAssetsApi(f.Group("/api"), appFs, db)

	resp, err := f.Test(req)
	if err != nil {
		return -1, nil, err
	}

	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, b, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func assetsUnmarshalHelper(t *testing.T, body []byte) (pagination.PaginationResult, []assetResponse) {
	var respData pagination.PaginationResult
	err := json.Unmarshal(body, &respData)
	require.Nil(t, err)

	var assetsResponse []assetResponse
	for _, item := range respData.Items {
		var asset assetResponse
		require.Nil(t, json.Unmarshal(item, &asset))
		assetsResponse = append(assetsResponse, asset)
	}

	return respData, assetsResponse
}
