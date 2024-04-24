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

	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssets_GetAssets(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		status, body, err := assetsRequestHelper(appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := assetsUnmarshalHelper(t, body)
		assert.Zero(t, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 0)
	})

	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		// Create 10 assets
		daos.NewTestBuilder(t).Db(db).Courses(2).Assets(5).Attachments(2).Build()

		status, body, err := assetsRequestHelper(appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := assetsUnmarshalHelper(t, body)
		require.Equal(t, 10, int(paginationResp.TotalItems))
		require.Len(t, assetsResp, 10)
		require.Nil(t, assetsResp[0].Attachments)

		// ----------------------------
		// Attachments
		// ----------------------------
		status, body, err = assetsRequestHelper(appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/?expand=true", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp = assetsUnmarshalHelper(t, body)
		require.Equal(t, 10, int(paginationResp.TotalItems))
		require.Len(t, assetsResp, 10)
		require.Len(t, assetsResp[0].Attachments, 2)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		// Create 10 assets
		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Assets(5).Attachments(2).Build()

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		status, body, err := assetsRequestHelper(appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/?orderBy=created_at%20asc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := assetsUnmarshalHelper(t, body)
		require.Equal(t, 10, int(paginationResp.TotalItems))
		require.Len(t, assetsResp, 10)
		assert.Equal(t, testData[0].Assets[0].ID, assetsResp[0].ID)

		// ----------------------------
		// CREATED_AT DESC
		// ----------------------------
		status, body, err = assetsRequestHelper(appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/?orderBy=created_at%20desc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp = assetsUnmarshalHelper(t, body)
		require.Equal(t, 10, int(paginationResp.TotalItems))
		require.Len(t, assetsResp, 10)
		assert.Equal(t, testData[1].Assets[4].ID, assetsResp[0].ID)

		// ----------------------------
		// CREATED_AT ASC + ATTACHMENTS.TITLE DESC
		// ----------------------------
		attDao := daos.NewAttachmentDao(db)

		status, body, err = assetsRequestHelper(appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/?expand=true&orderBy=created_at%20asc,"+attDao.Table+".created_at%20desc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp = assetsUnmarshalHelper(t, body)
		require.Equal(t, 10, int(paginationResp.TotalItems))
		require.Len(t, assetsResp, 10)
		assert.Equal(t, testData[0].Assets[0].ID, assetsResp[0].ID)
		require.Len(t, assetsResp[0].Attachments, 2)
		assert.Equal(t, testData[0].Assets[0].Attachments[1].ID, assetsResp[0].Attachments[0].ID)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		// 18 assets
		testData := daos.NewTestBuilder(t).Db(db).Courses(3).Assets(6).Build()

		// ----------------------------
		// Get the first page (10 assets)
		// ----------------------------
		params := url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}

		status, body, err := assetsRequestHelper(appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := assetsUnmarshalHelper(t, body)
		assert.Equal(t, 18, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 10)

		// Check the last asset in the paginated response
		assert.Equal(t, testData[1].Assets[3].ID, assetsResp[9].ID)

		// ----------------------------
		// Get the second page (8 assets)
		// ----------------------------
		params = url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}

		status, body, err = assetsRequestHelper(appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp = assetsUnmarshalHelper(t, body)
		assert.Equal(t, 18, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 8)

		// Check the last asset in the paginated response
		assert.Equal(t, testData[2].Assets[5].ID, assetsResp[7].ID)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		// Drop the table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewAssetDao(db).Table)
		require.Nil(t, err)

		status, _, err := assetsRequestHelper(appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssets_GetAsset(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Assets(5).Attachments(2).Build()

		status, body, err := assetsRequestHelper(appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/"+testData[1].Assets[3].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp assetResponse
		err = json.Unmarshal(body, &assetResp)
		require.Nil(t, err)
		assert.Equal(t, testData[1].Assets[3].ID, assetResp.ID)
		assert.Equal(t, testData[1].Assets[3].Title, assetResp.Title)
		assert.Equal(t, testData[1].Assets[3].Path, assetResp.Path)
		assert.Equal(t, testData[1].Assets[3].CourseID, assetResp.CourseID)
		assert.Nil(t, assetResp.Attachments)

		// ----------------------------
		// Attachments
		// ----------------------------
		status, body, err = assetsRequestHelper(appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/"+testData[1].Assets[3].ID+"/?expand=true", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		err = json.Unmarshal(body, &assetResp)
		require.Nil(t, err)
		assert.Equal(t, testData[1].Assets[3].ID, assetResp.ID)
		assert.Equal(t, testData[1].Assets[3].Title, assetResp.Title)
		assert.Equal(t, testData[1].Assets[3].Path, assetResp.Path)
		assert.Equal(t, testData[1].Assets[3].CourseID, assetResp.CourseID)
		assert.Len(t, assetResp.Attachments, 2)
	})

	t.Run("404 (not found)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		status, _, err := assetsRequestHelper(appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		// Drop the table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewAssetDao(db).Table)
		require.Nil(t, err)

		status, _, err := assetsRequestHelper(appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssets_UpdateAsset(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Assets(1).Build()

		// ----------------------------
		// Update the video position
		// ----------------------------
		testData[0].Assets[0].VideoPos = 45

		// Marshal the asset for the request
		data, err := json.Marshal(assetResponseHelper([]*models.Asset{testData[0].Assets[0]})[0])
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/assets/"+testData[0].Assets[0].ID, strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		status, body, err := assetsRequestHelper(appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp1 assetResponse
		err = json.Unmarshal(body, &assetResp1)
		require.Nil(t, err)
		assert.Equal(t, testData[0].Assets[0].ID, assetResp1.ID)
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

		req = httptest.NewRequest(http.MethodPut, "/api/assets/"+testData[0].Assets[0].ID, strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		status, body, err = assetsRequestHelper(appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp2 assetResponse
		err = json.Unmarshal(body, &assetResp2)
		require.Nil(t, err)
		assert.Equal(t, testData[0].Assets[0].ID, assetResp2.ID)
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

		req = httptest.NewRequest(http.MethodPut, "/api/assets/"+testData[0].Assets[0].ID, strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		status, body, err = assetsRequestHelper(appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp3 assetResponse
		err = json.Unmarshal(body, &assetResp3)
		require.Nil(t, err)
		assert.Equal(t, testData[0].Assets[0].ID, assetResp3.ID)
		assert.Equal(t, 10, assetResp3.VideoPos)
		assert.False(t, assetResp3.Completed)
		assert.True(t, assetResp3.CompletedAt.IsZero())
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		req := httptest.NewRequest(http.MethodPut, "/api/assets/test", strings.NewReader(`bob`))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := assetsRequestHelper(appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("404 (not found)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		req := httptest.NewRequest(http.MethodPut, "/api/assets/test", strings.NewReader(`{"id": "1234567"}`))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := assetsRequestHelper(appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		// Drop the table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewAssetDao(db).Table)
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/assets/test", strings.NewReader(`{"id": "1234567"}`))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := assetsRequestHelper(appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssets_ServeAsset(t *testing.T) {
	t.Run("200 (full video)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Assets(2).Build()

		// Create asset
		require.Nil(t, appFs.Fs.MkdirAll(filepath.Dir(testData[0].Assets[1].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(appFs.Fs, testData[0].Assets[1].Path, []byte("video"), os.ModePerm))

		status, body, err := assetsRequestHelper(appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/"+testData[0].Assets[1].ID+"/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "video", string(body))
	})

	t.Run("200 (stream video)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Assets(2).Build()

		// Create asset
		require.Nil(t, appFs.Fs.MkdirAll(filepath.Dir(testData[0].Assets[1].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(appFs.Fs, testData[0].Assets[1].Path, []byte("video"), os.ModePerm))

		req := httptest.NewRequest(http.MethodGet, "/api/assets/"+testData[0].Assets[1].ID+"/serve", nil)
		req.Header.Add("Range", "bytes=0-")

		status, body, err := assetsRequestHelper(appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusPartialContent, status)
		assert.Equal(t, "video", string(body))
	})

	t.Run("200 (html)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		testData := daos.NewTestBuilder(t).Courses(1).Assets(1).Build()

		// Change asset to an html file
		testData[0].Assets[0].Path = strings.Replace(testData[0].Assets[0].Path, ".mp4", ".html", 1)
		testData[0].Assets[0].Type = *types.NewAsset("html")

		coursesDao := daos.NewCourseDao(db)
		assetsDao := daos.NewAssetDao(db)

		require.Nil(t, coursesDao.Create(testData[0].Course))
		require.Nil(t, assetsDao.Create(testData[0].Assets[0]))

		// Create asset
		require.Nil(t, appFs.Fs.MkdirAll(filepath.Dir(testData[0].Assets[0].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(appFs.Fs, testData[0].Assets[0].Path, []byte("html"), os.ModePerm))

		req := httptest.NewRequest(http.MethodGet, "/api/assets/"+testData[0].Assets[0].ID+"/serve", nil)

		status, body, err := assetsRequestHelper(appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "html", string(body))
	})

	t.Run("400 (invalid path)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Assets(2).Build()

		status, body, err := assetsRequestHelper(appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/"+testData[0].Assets[1].ID+"/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		assert.Contains(t, string(body), "asset does not exist")
	})

	t.Run("400 (invalid video range)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Assets(2).Build()

		// Create asset
		require.Nil(t, appFs.Fs.MkdirAll(filepath.Dir(testData[0].Assets[1].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(appFs.Fs, testData[0].Assets[1].Path, []byte("video"), os.ModePerm))

		req := httptest.NewRequest(http.MethodGet, "/api/assets/"+testData[0].Assets[1].ID+"/serve", nil)
		req.Header.Add("Range", "bytes=10-1")

		status, body, err := assetsRequestHelper(appFs, db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		assert.Contains(t, string(body), "range start cannot be greater than end")
	})

	t.Run("404 (not found)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		status, _, err := assetsRequestHelper(appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/test/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewAssetDao(db).Table)
		require.Nil(t, err)

		status, _, err := assetsRequestHelper(appFs, db, httptest.NewRequest(http.MethodGet, "/api/assets/test/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPERS
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func assetsRequestHelper(appFs *appFs.AppFs, db database.Database, req *http.Request) (int, []byte, error) {
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
