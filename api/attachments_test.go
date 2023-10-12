package api

// import (
// 	"encoding/json"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/geerew/off-course/models"
// 	"github.com/gofiber/fiber/v2"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestAttachments_GetAttachments(t *testing.T) {
// 	t.Run("200 (empty)", func(t *testing.T) {
// 		_, db, _, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindAttachmentsApi(f.Group("/api"), db)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData []attachmentResponse
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Len(t, respData, 0)
// 	})

// 	t.Run("200 (found)", func(t *testing.T) {
// 		_, db, _, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindAttachmentsApi(f.Group("/api"), db)

// 		// Create 2 courses with 5 assets each with 2 attachments each (20 attachments total)
// 		courses := models.CreateTestCourses(t, db, 2)
// 		assets := models.CreateTestAssets(t, db, courses, 5)
// 		attachments := models.CreateTestAttachments(t, db, assets, 2)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData []attachmentResponse
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Len(t, respData, 20)
// 		assert.Equal(t, attachments[19].ID, respData[0].ID)
// 		assert.Equal(t, attachments[19].Title, respData[0].Title)
// 		assert.Equal(t, attachments[19].Path, respData[0].Path)
// 	})

// 	t.Run("200 (orderBy)", func(t *testing.T) {
// 		_, db, _, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindAttachmentsApi(f.Group("/api"), db)

// 		// Create 2 courses with 5 assets each with 2 attachments each (20 attachments total)
// 		courses := models.CreateTestCourses(t, db, 2)
// 		assets := models.CreateTestAssets(t, db, courses, 5)
// 		attachments := models.CreateTestAttachments(t, db, assets, 2)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/?orderBy=created_at%20asc", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData []attachmentResponse
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Len(t, respData, 20)

// 		assert.Equal(t, attachments[0].ID, respData[0].ID)
// 		assert.Equal(t, attachments[0].Title, respData[0].Title)
// 		assert.Equal(t, attachments[0].Path, respData[0].Path)
// 	})

// 	t.Run("500 (internal error)", func(t *testing.T) {
// 		_, db, _, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindAttachmentsApi(f.Group("/api"), db)

// 		// Drop the attachments table
// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Attachment{}))

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestAttachments_GetAsset(t *testing.T) {
// 	t.Run("200 (found)", func(t *testing.T) {
// 		_, db, _, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindAttachmentsApi(f.Group("/api"), db)

// 		// Create 2 courses with 5 assets with 2 attachments
// 		courses := models.CreateTestCourses(t, db, 2)
// 		assets := models.CreateTestAssets(t, db, courses, 5)
// 		attachments := models.CreateTestAttachments(t, db, assets, 2)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/"+attachments[6].ID, nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData attachmentResponse
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, attachments[6].ID, respData.ID)
// 		assert.Equal(t, attachments[6].Title, respData.Title)
// 		assert.Equal(t, attachments[6].Path, respData.Path)
// 	})

// 	t.Run("404 (not found)", func(t *testing.T) {
// 		_, db, _, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindAttachmentsApi(f.Group("/api"), db)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/test", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusNotFound, resp.StatusCode)
// 	})

// 	t.Run("500 (internal error)", func(t *testing.T) {
// 		_, db, _, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindAttachmentsApi(f.Group("/api"), db)

// 		// Drop the table
// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Attachment{}))

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/attachments/test", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
// 	})
// }
