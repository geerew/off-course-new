package api

// import (
// 	"encoding/json"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"net/url"
// 	"testing"

// 	"github.com/geerew/off-course/models"
// 	"github.com/geerew/off-course/utils/pagination"
// 	"github.com/gofiber/fiber/v2"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestAssets_GetAssets(t *testing.T) {
// 	t.Run("200 (empty)", func(t *testing.T) {
// 		_, db, _, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindAssetsApi(f.Group("/api"), db)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData pagination.PaginationResult
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 0, int(respData.TotalItems))
// 		assert.Len(t, respData.Items, 0)
// 	})

// 	t.Run("200 (found)", func(t *testing.T) {
// 		_, db, _, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindAssetsApi(f.Group("/api"), db)

// 		// Create 2 courses with 5 assets each with 2 attachments each
// 		courses := models.CreateTestCourses(t, db, 2)
// 		assets := models.CreateTestAssets(t, db, courses, 5)
// 		models.CreateTestAttachments(t, db, assets, 2)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData pagination.PaginationResult
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 10, int(respData.TotalItems))

// 		// Unmarshal
// 		assetsResponse := []assetResponse{}
// 		for _, item := range respData.Items {
// 			var asset assetResponse
// 			require.Nil(t, json.Unmarshal(item, &asset))
// 			assetsResponse = append(assetsResponse, asset)
// 		}

// 		// Assert values. By default orderBy is desc so the last inserted course should be first
// 		assert.Equal(t, assets[9].ID, assetsResponse[0].ID)
// 		assert.Equal(t, assets[9].Title, assetsResponse[0].Title)
// 		assert.Equal(t, assets[9].Path, assetsResponse[0].Path)
// 		assert.Len(t, assetsResponse[9].Attachments, 2)
// 	})

// 	t.Run("200 (orderBy)", func(t *testing.T) {
// 		_, db, _, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindAssetsApi(f.Group("/api"), db)

// 		// Create 2 courses with 5 assets each (10 assets total)
// 		courses := models.CreateTestCourses(t, db, 2)
// 		assets := models.CreateTestAssets(t, db, courses, 5)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/?orderBy=created_at%20asc", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData pagination.PaginationResult
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 10, int(respData.TotalItems))

// 		// Unmarshal
// 		assetsResponse := []assetResponse{}
// 		for _, item := range respData.Items {
// 			var asset assetResponse
// 			require.Nil(t, json.Unmarshal(item, &asset))
// 			assetsResponse = append(assetsResponse, asset)
// 		}

// 		assert.Equal(t, assets[0].ID, assetsResponse[0].ID)
// 		assert.Equal(t, assets[0].Title, assetsResponse[0].Title)
// 		assert.Equal(t, assets[0].Path, assetsResponse[0].Path)
// 	})

// 	t.Run("200 (pagination)", func(t *testing.T) {
// 		_, db, _, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindAssetsApi(f.Group("/api"), db)

// 		// Create 3 courses with 6 assets each (18 assets total)
// 		courses := models.CreateTestCourses(t, db, 3)
// 		assets := models.CreateTestAssets(t, db, courses, 6)

// 		// Get the first 10 assets
// 		params := url.Values{
// 			"orderBy":                    {"created_at asc"},
// 			pagination.PageQueryParam:    {"1"},
// 			pagination.PerPageQueryParam: {"10"},
// 		}
// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/?"+params.Encode(), nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData pagination.PaginationResult
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 18, int(respData.TotalItems))
// 		assert.Len(t, respData.Items, 10)

// 		// Unmarshal
// 		assetsResponse := []assetResponse{}
// 		for _, item := range respData.Items {
// 			var asset assetResponse
// 			require.Nil(t, json.Unmarshal(item, &asset))
// 			assetsResponse = append(assetsResponse, asset)
// 		}

// 		assert.Equal(t, assets[0].ID, assetsResponse[0].ID)
// 		assert.Equal(t, assets[0].Title, assetsResponse[0].Title)
// 		assert.Equal(t, assets[0].Path, assetsResponse[0].Path)

// 		// Get the next 8 assets
// 		params = url.Values{
// 			"orderBy":                    {"created_at asc"},
// 			pagination.PageQueryParam:    {"2"},
// 			pagination.PerPageQueryParam: {"10"},
// 		}
// 		resp, err = f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/?"+params.Encode(), nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ = io.ReadAll(resp.Body)

// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 18, int(respData.TotalItems))
// 		assert.Len(t, respData.Items, 8)

// 		// Unmarshal
// 		assetsResponse = []assetResponse{}
// 		for _, item := range respData.Items {
// 			var asset assetResponse
// 			require.Nil(t, json.Unmarshal(item, &asset))
// 			assetsResponse = append(assetsResponse, asset)
// 		}

// 		assert.Equal(t, assets[10].ID, assetsResponse[0].ID)
// 		assert.Equal(t, assets[10].Title, assetsResponse[0].Title)
// 		assert.Equal(t, assets[10].Path, assetsResponse[0].Path)
// 	})

// 	t.Run("500 (internal error)", func(t *testing.T) {
// 		_, db, _, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindAssetsApi(f.Group("/api"), db)

// 		// Drop the assets table
// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Asset{}))

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestAssets_GetAsset(t *testing.T) {
// 	t.Run("200 (found)", func(t *testing.T) {
// 		_, db, _, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindAssetsApi(f.Group("/api"), db)

// 		// Create 2 courses with 5 assets with 2 attachments
// 		courses := models.CreateTestCourses(t, db, 2)
// 		assets := models.CreateTestAssets(t, db, courses, 5)
// 		models.CreateTestAttachments(t, db, assets, 2)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/"+assets[6].ID, nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData assetResponse
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)

// 		assert.Equal(t, assets[6].ID, respData.ID)
// 		assert.Equal(t, assets[6].Title, respData.Title)
// 		assert.Equal(t, assets[6].Path, respData.Path)
// 		assert.Equal(t, assets[6].CourseID, respData.CourseID)
// 		assert.Len(t, respData.Attachments, 2)
// 	})

// 	t.Run("404 (not found)", func(t *testing.T) {
// 		_, db, _, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindAssetsApi(f.Group("/api"), db)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/test", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusNotFound, resp.StatusCode)
// 	})

// 	t.Run("500 (internal error)", func(t *testing.T) {
// 		_, db, _, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindAssetsApi(f.Group("/api"), db)

// 		// Drop the table
// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Asset{}))

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/test", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
// 	})
// }
