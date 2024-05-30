package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssets_GetAssets(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		router := setup(t)

		status, body, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/assets/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := unmarshalHelper[assetResponse](t, body)
		require.Zero(t, int(paginationResp.TotalItems))
		require.Zero(t, len(paginationResp.Items))
	})

	t.Run("200 (found)", func(t *testing.T) {
		router := setup(t)

		// Create 10 assets
		daos.NewTestBuilder(t).Db(router.config.DbManager.DataDb).Courses(2).Assets(5).Attachments(2).Build()

		status, body, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/assets/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := unmarshalHelper[assetResponse](t, body)
		require.Equal(t, 10, int(paginationResp.TotalItems))
		require.Len(t, assetsResp, 10)
		require.Nil(t, assetsResp[0].Attachments)

		// ----------------------------
		// Attachments
		// ----------------------------
		status, body, err = requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/assets/?expand=true", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp = unmarshalHelper[assetResponse](t, body)
		require.Equal(t, 10, int(paginationResp.TotalItems))
		require.Len(t, assetsResp, 10)
		require.Len(t, assetsResp[0].Attachments, 2)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		router := setup(t)

		// Create 10 assets
		testData := daos.NewTestBuilder(t).Db(router.config.DbManager.DataDb).Courses(2).Assets(5).Attachments(2).Build()

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		status, body, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/assets/?orderBy=created_at%20asc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := unmarshalHelper[assetResponse](t, body)
		require.Equal(t, 10, int(paginationResp.TotalItems))
		require.Len(t, assetsResp, 10)
		require.Equal(t, testData[0].Assets[0].ID, assetsResp[0].ID)

		// ----------------------------
		// CREATED_AT DESC
		// ----------------------------
		status, body, err = requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/assets/?orderBy=created_at%20desc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp = unmarshalHelper[assetResponse](t, body)
		require.Equal(t, 10, int(paginationResp.TotalItems))
		require.Len(t, assetsResp, 10)
		require.Equal(t, testData[1].Assets[4].ID, assetsResp[0].ID)

		// ----------------------------
		// CREATED_AT ASC + ATTACHMENTS.TITLE DESC
		// ----------------------------
		attDao := daos.NewAttachmentDao(router.config.DbManager.DataDb)

		status, body, err = requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/assets/?expand=true&orderBy=created_at%20asc,"+attDao.Table()+".created_at%20desc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp = unmarshalHelper[assetResponse](t, body)
		require.Equal(t, 10, int(paginationResp.TotalItems))
		require.Len(t, assetsResp, 10)
		require.Equal(t, testData[0].Assets[0].ID, assetsResp[0].ID)
		require.Len(t, assetsResp[0].Attachments, 2)
		require.Equal(t, testData[0].Assets[0].Attachments[1].ID, assetsResp[0].Attachments[0].ID)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		router := setup(t)

		// 18 assets
		testData := daos.NewTestBuilder(t).Db(router.config.DbManager.DataDb).Courses(3).Assets(6).Build()

		// ----------------------------
		// Get the first page (10 assets)
		// ----------------------------
		params := url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}

		status, body, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/assets/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := unmarshalHelper[assetResponse](t, body)
		require.Equal(t, 18, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 10)

		// Check the last asset in the paginated response
		require.Equal(t, testData[1].Assets[3].ID, assetsResp[9].ID)

		// ----------------------------
		// Get the second page (8 assets)
		// ----------------------------
		params = url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}

		status, body, err = requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/assets/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp = unmarshalHelper[assetResponse](t, body)
		require.Equal(t, 18, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 8)

		// Check the last asset in the paginated response
		require.Equal(t, testData[2].Assets[5].ID, assetsResp[7].ID)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router := setup(t)

		// Drop the table
		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + daos.NewAssetDao(router.config.DbManager.DataDb).Table())
		require.Nil(t, err)

		status, _, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/assets/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssets_GetAsset(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		router := setup(t)

		testData := daos.NewTestBuilder(t).Db(router.config.DbManager.DataDb).Courses(2).Assets(5).Attachments(2).Build()

		status, body, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/assets/"+testData[1].Assets[3].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp assetResponse
		err = json.Unmarshal(body, &assetResp)
		require.Nil(t, err)
		require.Equal(t, testData[1].Assets[3].ID, assetResp.ID)
		require.Equal(t, testData[1].Assets[3].Title, assetResp.Title)
		require.Equal(t, testData[1].Assets[3].Path, assetResp.Path)
		require.Equal(t, testData[1].Assets[3].CourseID, assetResp.CourseID)
		require.Nil(t, assetResp.Attachments)

		// ----------------------------
		// Attachments
		// ----------------------------
		status, body, err = requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/assets/"+testData[1].Assets[3].ID+"/?expand=true", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		err = json.Unmarshal(body, &assetResp)
		require.Nil(t, err)
		require.Equal(t, testData[1].Assets[3].ID, assetResp.ID)
		require.Equal(t, testData[1].Assets[3].Title, assetResp.Title)
		require.Equal(t, testData[1].Assets[3].Path, assetResp.Path)
		require.Equal(t, testData[1].Assets[3].CourseID, assetResp.CourseID)
		require.Len(t, assetResp.Attachments, 2)
	})

	t.Run("404 (not found)", func(t *testing.T) {
		router := setup(t)

		status, _, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/assets/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router := setup(t)

		// Drop the table
		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + daos.NewAssetDao(router.config.DbManager.DataDb).Table())
		require.Nil(t, err)

		status, _, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/assets/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssets_UpdateAsset(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		router := setup(t)

		testData := daos.NewTestBuilder(t).Db(router.config.DbManager.DataDb).Courses(1).Assets(1).Build()

		// ----------------------------
		// Update the video position
		// ----------------------------
		testData[0].Assets[0].VideoPos = 45

		// Marshal the asset for the request
		data, err := json.Marshal(assetResponseHelper([]*models.Asset{testData[0].Assets[0]})[0])
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/assets/"+testData[0].Assets[0].ID, strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		status, body, err := requestHelper(router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp1 assetResponse
		err = json.Unmarshal(body, &assetResp1)
		require.Nil(t, err)
		require.Equal(t, testData[0].Assets[0].ID, assetResp1.ID)
		require.Equal(t, 45, assetResp1.VideoPos)
		require.False(t, assetResp1.Completed)
		require.True(t, assetResp1.CompletedAt.IsZero())

		// ----------------------------
		// Set completed to true
		// ----------------------------
		assetResp1.Completed = true

		// Marshal the asset for the request
		data, err = json.Marshal(assetResp1)
		require.Nil(t, err)

		req = httptest.NewRequest(http.MethodPut, "/api/assets/"+testData[0].Assets[0].ID, strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		status, body, err = requestHelper(router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp2 assetResponse
		err = json.Unmarshal(body, &assetResp2)
		require.Nil(t, err)
		require.Equal(t, testData[0].Assets[0].ID, assetResp2.ID)
		require.Equal(t, 45, assetResp2.VideoPos)
		require.True(t, assetResp2.Completed)
		require.False(t, assetResp2.CompletedAt.IsZero())

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

		status, body, err = requestHelper(router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp3 assetResponse
		err = json.Unmarshal(body, &assetResp3)
		require.Nil(t, err)
		require.Equal(t, testData[0].Assets[0].ID, assetResp3.ID)
		require.Equal(t, 10, assetResp3.VideoPos)
		require.False(t, assetResp3.Completed)
		require.True(t, assetResp3.CompletedAt.IsZero())
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		router := setup(t)

		req := httptest.NewRequest(http.MethodPut, "/api/assets/test", strings.NewReader(`bob`))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := requestHelper(router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("400 (invalid course id)", func(t *testing.T) {
		router := setup(t)

		req := httptest.NewRequest(http.MethodPut, "/api/assets/test", strings.NewReader(`{"id": "1234567"}`))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := requestHelper(router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router := setup(t)

		// Drop the table
		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + daos.NewAssetDao(router.config.DbManager.DataDb).Table())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/assets/test", strings.NewReader(`{"id": "1234567"}`))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := requestHelper(router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssets_ServeAsset(t *testing.T) {
	t.Run("200 (full video)", func(t *testing.T) {
		router := setup(t)

		testData := daos.NewTestBuilder(t).Db(router.config.DbManager.DataDb).Courses(1).Assets(2).Build()

		// Create asset
		require.Nil(t, router.config.AppFs.Fs.MkdirAll(filepath.Dir(testData[0].Assets[1].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(router.config.AppFs.Fs, testData[0].Assets[1].Path, []byte("video"), os.ModePerm))

		status, body, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/assets/"+testData[0].Assets[1].ID+"/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, "video", string(body))
	})

	t.Run("200 (stream video)", func(t *testing.T) {
		router := setup(t)

		testData := daos.NewTestBuilder(t).Db(router.config.DbManager.DataDb).Courses(1).Assets(2).Build()

		// Create asset
		require.Nil(t, router.config.AppFs.Fs.MkdirAll(filepath.Dir(testData[0].Assets[1].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(router.config.AppFs.Fs, testData[0].Assets[1].Path, []byte("video"), os.ModePerm))

		req := httptest.NewRequest(http.MethodGet, "/api/assets/"+testData[0].Assets[1].ID+"/serve", nil)
		req.Header.Add("Range", "bytes=0-")

		status, body, err := requestHelper(router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusPartialContent, status)
		require.Equal(t, "video", string(body))
	})

	t.Run("200 (html)", func(t *testing.T) {
		router := setup(t)

		testData := daos.NewTestBuilder(t).Courses(1).Assets(1).Build()

		// Change asset to an html file
		testData[0].Assets[0].Path = strings.Replace(testData[0].Assets[0].Path, ".mp4", ".html", 1)
		testData[0].Assets[0].Type = *types.NewAsset("html")

		coursesDao := daos.NewCourseDao(router.config.DbManager.DataDb)
		assetsDao := daos.NewAssetDao(router.config.DbManager.DataDb)

		require.Nil(t, coursesDao.Create(testData[0].Course))
		require.Nil(t, assetsDao.Create(testData[0].Assets[0], nil))

		// Create asset
		require.Nil(t, router.config.AppFs.Fs.MkdirAll(filepath.Dir(testData[0].Assets[0].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(router.config.AppFs.Fs, testData[0].Assets[0].Path, []byte("html data"), os.ModePerm))

		req := httptest.NewRequest(http.MethodGet, "/api/assets/"+testData[0].Assets[0].ID+"/serve", nil)

		status, body, err := requestHelper(router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, "html data", string(body))
	})

	t.Run("400 (invalid path)", func(t *testing.T) {
		router := setup(t)

		testData := daos.NewTestBuilder(t).Db(router.config.DbManager.DataDb).Courses(1).Assets(2).Build()

		status, body, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/assets/"+testData[0].Assets[1].ID+"/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Asset does not exist")
	})

	t.Run("400 (invalid video range)", func(t *testing.T) {
		router := setup(t)

		testData := daos.NewTestBuilder(t).Db(router.config.DbManager.DataDb).Courses(1).Assets(2).Build()

		// Create asset
		require.Nil(t, router.config.AppFs.Fs.MkdirAll(filepath.Dir(testData[0].Assets[1].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(router.config.AppFs.Fs, testData[0].Assets[1].Path, []byte("video"), os.ModePerm))

		req := httptest.NewRequest(http.MethodGet, "/api/assets/"+testData[0].Assets[1].ID+"/serve", nil)
		req.Header.Add("Range", "bytes=10-1")

		status, body, err := requestHelper(router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Range start cannot be greater than end")
	})

	t.Run("404 (not found)", func(t *testing.T) {
		router := setup(t)

		status, _, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/assets/test/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router := setup(t)

		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + daos.NewAssetDao(router.config.DbManager.DataDb).Table())
		require.Nil(t, err)

		status, _, err := requestHelper(router, httptest.NewRequest(http.MethodGet, "/api/assets/test/serve", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}
