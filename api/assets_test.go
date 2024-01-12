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

		// Create 10 assets
		models.NewTestData(t, db, 2, false, 5, 0)

		status, body, err := assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := assetsUnmarshalHelper(t, body)
		require.Equal(t, 10, int(paginationResp.TotalItems))
		require.Len(t, assetsResp, 10)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// Create 10 assets
		workingData := models.NewTestData(t, db, 2, false, 5, 0)

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		status, body, err := assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/?orderBy=created_at%20asc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := assetsUnmarshalHelper(t, body)
		require.Equal(t, 10, int(paginationResp.TotalItems))
		require.Len(t, assetsResp, 10)
		assert.Equal(t, workingData[0].Assets[0].ID, assetsResp[0].ID)
		assert.Equal(t, workingData[0].Assets[0].Title, assetsResp[0].Title)
		assert.Equal(t, workingData[0].Assets[0].Path, assetsResp[0].Path)

		// ----------------------------
		// CREATED_AT DESC
		// ----------------------------
		status, body, err = assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/?orderBy=created_at%20desc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp = assetsUnmarshalHelper(t, body)
		require.Equal(t, 10, int(paginationResp.TotalItems))
		require.Len(t, assetsResp, 10)
		assert.Equal(t, workingData[1].Assets[4].ID, assetsResp[0].ID)
		assert.Equal(t, workingData[1].Assets[4].Title, assetsResp[0].Title)
		assert.Equal(t, workingData[1].Assets[4].Path, assetsResp[0].Path)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		// 18 assets
		workingData := models.NewTestData(t, db, 3, false, 6, 0)

		// ----------------------------
		// Get the first page (10 assets)
		// ----------------------------
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
		assert.Equal(t, workingData[1].Assets[3].ID, assetsResp[9].ID)

		// ----------------------------
		// Get the second page (8 assets)
		// ----------------------------
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
		assert.Equal(t, workingData[2].Assets[5].ID, assetsResp[7].ID)
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

		workingData := models.NewTestData(t, db, 2, false, 5, 2)

		status, body, err := assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/"+workingData[1].Assets[3].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp assetResponse
		err = json.Unmarshal(body, &assetResp)
		require.Nil(t, err)
		assert.Equal(t, workingData[1].Assets[3].ID, assetResp.ID)
		assert.Equal(t, workingData[1].Assets[3].Title, assetResp.Title)
		assert.Equal(t, workingData[1].Assets[3].Path, assetResp.Path)
		assert.Equal(t, workingData[1].Assets[3].CourseID, assetResp.CourseID)
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

		workingData := models.NewTestData(t, db, 1, false, 1, 0)

		// ----------------------------
		// Update the video position
		// ----------------------------
		workingData[0].Assets[0].VideoPos = 45

		// Marshal the asset for the request
		data, err := json.Marshal(assetResponseHelper([]*models.Asset{workingData[0].Assets[0]})[0])
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/assets/"+workingData[0].Assets[0].ID, strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		status, body, err := assetsRequestHelper(t, appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp1 assetResponse
		err = json.Unmarshal(body, &assetResp1)
		require.Nil(t, err)
		assert.Equal(t, workingData[0].Assets[0].ID, assetResp1.ID)
		assert.Equal(t, 45, assetResp1.VideoPos)
		assert.False(t, assetResp1.Completed)
		assert.True(t, assetResp1.CompletedAt.IsZero())

		// ----------------------------
		// Set completed to true
		// ----------------------------
		assetResp1.Completed = true

		// Marshal the asset for the request
		data, err = json.Marshal(assetResp1)
		require.Nil(t, err)

		req = httptest.NewRequest(http.MethodPut, "/api/assets/"+workingData[0].Assets[0].ID, strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		status, body, err = assetsRequestHelper(t, appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp2 assetResponse
		err = json.Unmarshal(body, &assetResp2)
		require.Nil(t, err)
		assert.Equal(t, workingData[0].Assets[0].ID, assetResp2.ID)
		assert.Equal(t, 45, assetResp2.VideoPos)
		assert.True(t, assetResp2.Completed)
		assert.False(t, assetResp2.CompletedAt.IsZero())

		// ----------------------------
		// Set video position to 10 completed to false
		// ----------------------------
		assetResp2.VideoPos = 10
		assetResp2.Completed = false

		// Marshal the asset for the request
		data, err = json.Marshal(assetResp2)
		require.Nil(t, err)

		req = httptest.NewRequest(http.MethodPut, "/api/assets/"+workingData[0].Assets[0].ID, strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		status, body, err = assetsRequestHelper(t, appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp3 assetResponse
		err = json.Unmarshal(body, &assetResp3)
		require.Nil(t, err)
		assert.Equal(t, workingData[0].Assets[0].ID, assetResp3.ID)
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

		workingData := models.NewTestData(t, db, 1, false, 2, 0)

		// Create asset
		require.Nil(t, appFs.Fs.MkdirAll(filepath.Dir(workingData[0].Assets[1].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(appFs.Fs, workingData[0].Assets[1].Path, []byte("video"), os.ModePerm))

		status, body, err := assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/"+workingData[0].Assets[1].ID+"/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "video", string(body))
	})

	t.Run("200 (stream video)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 1, false, 2, 0)

		// Create asset
		require.Nil(t, appFs.Fs.MkdirAll(filepath.Dir(workingData[0].Assets[1].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(appFs.Fs, workingData[0].Assets[1].Path, []byte("video"), os.ModePerm))

		req := httptest.NewRequest(http.MethodGet, "/api/assets/"+workingData[0].Assets[1].ID+"/serve", nil)
		req.Header.Add("Range", "bytes=0-")

		status, body, err := assetsRequestHelper(t, appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusPartialContent, status)
		assert.Equal(t, "video", string(body))
	})

	t.Run("400 (invalid path)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 1, false, 2, 0)

		status, body, err := assetsRequestHelper(t, appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/"+workingData[0].Assets[1].ID+"/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		assert.Contains(t, string(body), "asset does not exist")
	})

	t.Run("400 (invalid video range)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 1, false, 2, 0)

		// Create asset
		require.Nil(t, appFs.Fs.MkdirAll(filepath.Dir(workingData[0].Assets[1].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(appFs.Fs, workingData[0].Assets[1].Path, []byte("video"), os.ModePerm))

		req := httptest.NewRequest(http.MethodGet, "/api/assets/"+workingData[0].Assets[1].ID+"/serve", nil)
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
