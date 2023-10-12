package api

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"net/url"
// 	"strings"
// 	"testing"

// 	"github.com/geerew/off-course/models"
// 	"github.com/geerew/off-course/utils/pagination"
// 	"github.com/gofiber/fiber/v2"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestScans_GetScans(t *testing.T) {
// 	t.Run("200 (empty)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindScansApi(f.Group("/api"), appFs, db, courseScanner)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/scans/", nil))
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
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindScansApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 5 scans
// 		courses := models.CreateTestCourses(t, db, 5)
// 		scans := models.CreateTestScans(t, db, courses)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/scans/", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData pagination.PaginationResult
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 5, int(respData.TotalItems))

// 		// Unmarshal
// 		var scansResponse []scanResponse
// 		for _, item := range respData.Items {
// 			var scan scanResponse
// 			require.Nil(t, json.Unmarshal(item, &scan))			require.Nil(t, err)
// 			scansResponse = append(scansResponse, scan)
// 		}

// 		// Assert values. By default orderBy is desc so the last inserted scan should be first
// 		assert.Equal(t, scans[4].ID, scansResponse[0].ID)
// 		assert.Equal(t, scans[4].CourseID, scansResponse[0].CourseID)
// 		assert.Equal(t, scans[4].Status, scansResponse[0].Status)

// 		// Assert the course
// 		assert.Nil(t, scansResponse[0].Course)
// 	})

// 	t.Run("200 (orderBy)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindScansApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 5 scans
// 		courses := models.CreateTestCourses(t, db, 5)
// 		scans := models.CreateTestScans(t, db, courses)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/scans/?orderBy=created_at%20asc", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData pagination.PaginationResult
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 5, int(respData.TotalItems))

// 		// Unmarshal
// 		var scansResponse []scanResponse
// 		for _, item := range respData.Items {
// 			var scan scanResponse
// 			require.Nil(t, json.Unmarshal(item, &scan))			require.Nil(t, err)
// 			scansResponse = append(scansResponse, scan)
// 		}

// 		assert.Equal(t, scans[0].ID, scansResponse[0].ID)
// 		assert.Equal(t, scans[0].CourseID, scansResponse[0].CourseID)
// 		assert.Equal(t, scans[0].Status, scansResponse[0].Status)
// 	})

// 	t.Run("200 (pagination)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindScansApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 17 scans
// 		courses := models.CreateTestCourses(t, db, 17)
// 		scans := models.CreateTestScans(t, db, courses)

// 		// Get the first 10 scans
// 		params := url.Values{
// 			"orderBy":                    {"created_at asc"},
// 			pagination.PageQueryParam:    {"1"},
// 			pagination.PerPageQueryParam: {"10"},
// 		}
// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/scans/?"+params.Encode(), nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData pagination.PaginationResult
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 17, int(respData.TotalItems))
// 		assert.Len(t, respData.Items, 10)

// 		// Unmarshal
// 		var scansResponse []scanResponse
// 		for _, item := range respData.Items {
// 			var scan scanResponse
// 			require.Nil(t, json.Unmarshal(item, &scan))			require.Nil(t, err)
// 			scansResponse = append(scansResponse, scan)
// 		}

// 		assert.Equal(t, scans[0].ID, scansResponse[0].ID)
// 		assert.Equal(t, scans[0].CourseID, scansResponse[0].CourseID)
// 		assert.Equal(t, scans[0].Status, scansResponse[0].Status)

// 		// Get the next 7 scans
// 		params = url.Values{
// 			"orderBy":                    {"created_at asc"},
// 			pagination.PageQueryParam:    {"2"},
// 			pagination.PerPageQueryParam: {"10"},
// 		}
// 		resp, err = f.Test(httptest.NewRequest(http.MethodGet, "/api/scans/?"+params.Encode(), nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ = io.ReadAll(resp.Body)

// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 17, int(respData.TotalItems))
// 		assert.Len(t, respData.Items, 7)

// 		// Unmarshal
// 		scansResponse = []scanResponse{}
// 		for _, item := range respData.Items {
// 			var scan scanResponse
// 			require.Nil(t, json.Unmarshal(item, &scan))			require.Nil(t, err)
// 			scansResponse = append(scansResponse, scan)
// 		}

// 		assert.Equal(t, scans[10].ID, scansResponse[0].ID)
// 		assert.Equal(t, scans[10].CourseID, scansResponse[0].CourseID)
// 		assert.Equal(t, scans[10].Status, scansResponse[0].Status)
// 	})

// 	t.Run("200 (include course)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindScansApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 5 scans
// 		courses := models.CreateTestCourses(t, db, 5)
// 		scans := models.CreateTestScans(t, db, courses)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/scans/?orderBy=created_at%20asc&includeCourse=true", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData pagination.PaginationResult
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 5, int(respData.TotalItems))

// 		// Unmarshal
// 		var scansResponse []scanResponse
// 		for _, item := range respData.Items {
// 			var scan scanResponse
// 			require.Nil(t, json.Unmarshal(item, &scan))			require.Nil(t, err)
// 			scansResponse = append(scansResponse, scan)
// 		}

// 		assert.Equal(t, scans[0].ID, scansResponse[0].ID)
// 		assert.Equal(t, scans[0].CourseID, scansResponse[0].CourseID)
// 		assert.Equal(t, scans[0].Status, scansResponse[0].Status)

// 		// Assert the course
// 		require.NotNil(t, scansResponse[0].Course)
// 		assert.Equal(t, courses[0].Path, scansResponse[0].Course.Path)
// 	})

// 	t.Run("500 (internal error)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindScansApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Drop the scans table
// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Scan{}))

// 		// Get the first 10 scans
// 		params := url.Values{
// 			"orderBy":                    {"created_at asc"},
// 			pagination.PageQueryParam:    {"1"},
// 			pagination.PerPageQueryParam: {"10"},
// 		}
// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/scans/?"+params.Encode(), nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestScans_GetScan(t *testing.T) {
// 	t.Run("200 (found)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindScansApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 5 scans
// 		courses := models.CreateTestCourses(t, db, 5)
// 		scans := models.CreateTestScans(t, db, courses)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/scans/"+scans[2].ID, nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData scanResponse
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, scans[2].ID, respData.ID)
// 		assert.Equal(t, scans[2].CourseID, respData.CourseID)
// 		assert.Equal(t, scans[2].Status, respData.Status)
// 		assert.Nil(t, respData.Course)
// 	})

// 	t.Run("200 (include course)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindScansApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 5 scans
// 		courses := models.CreateTestCourses(t, db, 5)
// 		scans := models.CreateTestScans(t, db, courses)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/scans/"+scans[2].ID+"?includeCourse=true", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData scanResponse
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, scans[2].ID, respData.ID)
// 		assert.Equal(t, scans[2].CourseID, respData.CourseID)
// 		assert.Equal(t, scans[2].Status, respData.Status)

// 		// Assert the course
// 		require.NotNil(t, respData.Course)
// 		assert.Equal(t, courses[2].Path, respData.Course.Path)
// 	})

// 	t.Run("404 (not found)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindScansApi(f.Group("/api"), appFs, db, courseScanner)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/scans/test", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusNotFound, resp.StatusCode)
// 	})

// 	t.Run("500 (internal error)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindScansApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Drop the table
// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Scan{}))

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/scans/test", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestScans_GetScanByCourseId(t *testing.T) {
// 	t.Run("200 (found)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindScansApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 5 scans
// 		courses := models.CreateTestCourses(t, db, 5)
// 		scans := models.CreateTestScans(t, db, courses)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/scans/course/"+scans[2].CourseID, nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData scanResponse
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, scans[2].ID, respData.ID)
// 		assert.Equal(t, scans[2].CourseID, respData.CourseID)
// 		assert.Equal(t, scans[2].Status, respData.Status)
// 	})

// 	t.Run("404 (not found)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindScansApi(f.Group("/api"), appFs, db, courseScanner)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/scans/course/test", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusNotFound, resp.StatusCode)
// 	})

// 	t.Run("500 (internal error)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindScansApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Drop the table
// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Scan{}))

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/scans/course/test", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestScans_CreateScan(t *testing.T) {
// 	t.Run("201 (created)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindScansApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create a course
// 		course := models.CreateTestCourses(t, db, 1)[0]

// 		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(fmt.Sprintf(`{"courseID": "%s"}`, course.ID)))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		resp, err := f.Test(req)
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusCreated, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)
// 		var respData scanResponse
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, course.ID, respData.CourseID)
// 	})

// 	t.Run("400 (bind error)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindScansApi(f.Group("/api"), appFs, db, courseScanner)

// 		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(`{`))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		resp, err := f.Test(req)
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)
// 		assert.Contains(t, string(body), "error parsing data")
// 	})

// 	t.Run("400 (invalid data)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindScansApi(f.Group("/api"), appFs, db, courseScanner)

// 		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(`{"courseID": ""}`))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		resp, err := f.Test(req)
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)
// 		assert.Contains(t, string(body), "a course ID is required")
// 	})

// 	t.Run("400 (invalid course id)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindScansApi(f.Group("/api"), appFs, db, courseScanner)

// 		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(`{"courseID": "test"}`))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		resp, err := f.Test(req)
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)
// 		assert.Contains(t, string(body), "invalid course ID")
// 	})

// 	t.Run("500 (internal error)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindScansApi(f.Group("/api"), appFs, db, courseScanner)

// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Scan{}))

// 		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(`{"courseID": "test"}`))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		resp, err := f.Test(req)
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)
// 		assert.Contains(t, string(body), "error creating scan job")
// 	})
// }
