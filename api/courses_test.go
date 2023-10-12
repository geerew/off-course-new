package api

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"net/url"
// 	"os"
// 	"strings"
// 	"testing"

// 	"github.com/geerew/off-course/models"
// 	"github.com/geerew/off-course/utils/pagination"
// 	"github.com/gofiber/fiber/v2"
// 	"github.com/spf13/afero"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"gorm.io/gorm"
// )

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestCourses_GetCourses(t *testing.T) {
// 	t.Run("200 (empty)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
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
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 5 courses with 5 assets each
// 		courses := models.CreateTestCourses(t, db, 5)
// 		models.CreateTestAssets(t, db, courses, 5)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData pagination.PaginationResult
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 5, int(respData.TotalItems))

// 		// Unmarshal
// 		var coursesResponse []courseResponse
// 		for _, item := range respData.Items {
// 			var course courseResponse
// 			require.Nil(t, json.Unmarshal(item, &course))
// 			coursesResponse = append(coursesResponse, course)
// 		}

// 		// Assert values. By default orderBy is desc so the last inserted course should be first
// 		assert.Equal(t, courses[4].ID, coursesResponse[0].ID)
// 		assert.Equal(t, courses[4].Title, coursesResponse[0].Title)
// 		assert.Equal(t, courses[4].Path, coursesResponse[0].Path)

// 		// Assert assets are not included
// 		assert.Nil(t, coursesResponse[0].Assets)
// 	})

// 	t.Run("200 (orderBy)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 5 courses
// 		courses := models.CreateTestCourses(t, db, 5)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/?orderBy=created_at%20asc", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData pagination.PaginationResult
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 5, int(respData.TotalItems))

// 		// Unmarshal
// 		var coursesResponse []courseResponse
// 		for _, item := range respData.Items {
// 			var course courseResponse
// 			require.Nil(t, json.Unmarshal(item, &course))
// 			coursesResponse = append(coursesResponse, course)
// 		}

// 		assert.Equal(t, courses[0].ID, coursesResponse[0].ID)
// 		assert.Equal(t, courses[0].Title, coursesResponse[0].Title)
// 		assert.Equal(t, courses[0].Path, coursesResponse[0].Path)
// 	})

// 	t.Run("200 (pagination)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 17 courses
// 		courses := models.CreateTestCourses(t, db, 17)

// 		// Get the first 10 courses
// 		params := url.Values{
// 			"orderBy":                    {"created_at asc"},
// 			pagination.PageQueryParam:    {"1"},
// 			pagination.PerPageQueryParam: {"10"},
// 		}
// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/?"+params.Encode(), nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData pagination.PaginationResult
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 17, int(respData.TotalItems))
// 		assert.Len(t, respData.Items, 10)

// 		// Unmarshal
// 		var coursesResponse []courseResponse
// 		for _, item := range respData.Items {
// 			var course courseResponse
// 			require.Nil(t, json.Unmarshal(item, &course))
// 			coursesResponse = append(coursesResponse, course)
// 		}

// 		assert.Equal(t, courses[0].ID, coursesResponse[0].ID)
// 		assert.Equal(t, courses[0].Title, coursesResponse[0].Title)
// 		assert.Equal(t, courses[0].Path, coursesResponse[0].Path)

// 		// Get the next 7 courses
// 		params = url.Values{
// 			"orderBy":                    {"created_at asc"},
// 			pagination.PageQueryParam:    {"2"},
// 			pagination.PerPageQueryParam: {"10"},
// 		}
// 		resp, err = f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/?"+params.Encode(), nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ = io.ReadAll(resp.Body)

// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 17, int(respData.TotalItems))
// 		assert.Len(t, respData.Items, 7)

// 		// Unmarshal
// 		coursesResponse = []courseResponse{}
// 		for _, item := range respData.Items {
// 			var course courseResponse
// 			require.Nil(t, json.Unmarshal(item, &course))
// 			coursesResponse = append(coursesResponse, course)
// 		}

// 		assert.Equal(t, courses[10].ID, coursesResponse[0].ID)
// 		assert.Equal(t, courses[10].Title, coursesResponse[0].Title)
// 		assert.Equal(t, courses[10].Path, coursesResponse[0].Path)
// 	})

// 	t.Run("200 (include assets)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 5 courses with 5 assets each
// 		courses := models.CreateTestCourses(t, db, 5)
// 		models.CreateTestAssets(t, db, courses, 5)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/?orderBy=created_at%20asc&includeAssets=true", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData pagination.PaginationResult
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 5, int(respData.TotalItems))

// 		// Unmarshal
// 		coursesResponse := []courseResponse{}
// 		for _, item := range respData.Items {
// 			var course courseResponse
// 			require.Nil(t, json.Unmarshal(item, &course))
// 			coursesResponse = append(coursesResponse, course)
// 		}

// 		assert.Equal(t, courses[0].ID, coursesResponse[0].ID)
// 		assert.Equal(t, courses[0].Title, coursesResponse[0].Title)
// 		assert.Equal(t, courses[0].Path, coursesResponse[0].Path)

// 		// Assert the assets
// 		require.NotNil(t, coursesResponse[0].Assets)
// 		assert.Len(t, coursesResponse[0].Assets, 5)
// 	})

// 	t.Run("500 (internal error)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Drop the courses table
// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Course{}))

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestCourses_GetCourse(t *testing.T) {
// 	t.Run("200 (found)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 5 courses with 5 asserts each
// 		courses := models.CreateTestCourses(t, db, 5)
// 		models.CreateTestAssets(t, db, courses, 5)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[2].ID, nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData courseResponse
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, courses[2].ID, respData.ID)
// 		assert.Equal(t, courses[2].Title, respData.Title)
// 		assert.Equal(t, courses[2].Path, respData.Path)
// 		assert.Nil(t, respData.Assets)
// 	})

// 	t.Run("200 (include assets)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 5 courses with 5 assets each
// 		courses := models.CreateTestCourses(t, db, 5)
// 		assets := models.CreateTestAssets(t, db, courses, 5)
// 		models.CreateTestAttachments(t, db, assets, 3)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[2].ID+"?includeAssets=true", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData courseResponse
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, courses[2].ID, respData.ID)
// 		assert.Equal(t, courses[2].Title, respData.Title)
// 		assert.Equal(t, courses[2].Path, respData.Path)
// 		require.Len(t, respData.Assets, 5)
// 		assert.Len(t, respData.Assets[0].Attachments, 3)
// 	})

// 	t.Run("404 (not found)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/test", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusNotFound, resp.StatusCode)
// 	})

// 	t.Run("500 (internal error)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Drop the table
// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Course{}))

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/test", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestCourses_Create(t *testing.T) {
// 	t.Run("201 (created)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create the course path
// 		coursePath := "/course 1/"
// 		appFs.Fs.MkdirAll(coursePath, os.ModePerm)

// 		postData := fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
// 		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(postData))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		resp, err := f.Test(req)
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusCreated, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)
// 		var respData courseResponse
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.NotNil(t, respData.ID)
// 		assert.Equal(t, "course 1", respData.Title)
// 		assert.Equal(t, coursePath, respData.Path)
// 	})

// 	t.Run("400 (bind error)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{`))
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
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Missing title
// 		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": ""}`))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		resp, err := f.Test(req)
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
// 		body, _ := io.ReadAll(resp.Body)
// 		assert.Contains(t, string(body), "a title and path are required")

// 		// Missing path
// 		req = httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": "course 1", "path": ""}`))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		resp, err = f.Test(req)
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
// 		body, _ = io.ReadAll(resp.Body)
// 		assert.Contains(t, string(body), "a title and path are required")

// 		// Invalid path
// 		req = httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": "course 1", "path": "/test"}`))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		resp, err = f.Test(req)
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
// 		body, _ = io.ReadAll(resp.Body)
// 		assert.Contains(t, string(body), "invalid course path")
// 	})

// 	t.Run("400 (existing course)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create the course path
// 		coursePath := "/course 1/"
// 		appFs.Fs.MkdirAll(coursePath, os.ModePerm)

// 		postData := fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
// 		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(postData))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		resp, err := f.Test(req)
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusCreated, resp.StatusCode)

// 		// Do it again
// 		postData = fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
// 		req = httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(postData))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		resp, err = f.Test(req)
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
// 		body, _ := io.ReadAll(resp.Body)
// 		assert.Contains(t, string(body), "a course with this path already exists ")
// 	})

// 	t.Run("500 (internal error)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Course{}))

// 		// Create the course path
// 		coursePath := "/course 1/"
// 		appFs.Fs.MkdirAll(coursePath, os.ModePerm)

// 		postData := fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
// 		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(postData))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		resp, err := f.Test(req)
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)
// 		assert.Contains(t, string(body), "error creating course")
// 	})

// 	t.Run("500 (scan error)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Scan{}))

// 		// Create the course path
// 		coursePath := "/course 1/"
// 		appFs.Fs.MkdirAll(coursePath, os.ModePerm)

// 		postData := fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
// 		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(postData))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		resp, err := f.Test(req)
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)
// 		assert.Contains(t, string(body), "error creating scan job")
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestCourses_DeleteCourse(t *testing.T) {
// 	t.Run("204 (deleted)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 5 courses
// 		courses := models.CreateTestCourses(t, db, 5)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodDelete, "/api/courses/"+courses[2].ID, nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusNoContent, resp.StatusCode)

// 		// Ensure the course was deleted
// 		_, err = models.GetCourse(db, courses[2].ID, nil)
// 		assert.Equal(t, gorm.ErrRecordNotFound, err)
// 	})

// 	t.Run("404 (not found)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodDelete, "/api/courses/test", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusNotFound, resp.StatusCode)
// 	})

// 	t.Run("500 (internal error)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Drop the table
// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Course{}))

// 		resp, err := f.Test(httptest.NewRequest(http.MethodDelete, "/api/courses/test", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
// 	})
// }

// func TestCourses_Card(t *testing.T) {
// 	t.Run("200 (found)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create a course and set the card path
// 		course := models.CreateTestCourses(t, db, 1)[0]
// 		require.Nil(t, models.UpdateCourseCardPath(db, course, "/"+course.Path+"/card.png"))

// 		// Create the course path
// 		appFs.Fs.MkdirAll("/"+course.Path, os.ModePerm)
// 		require.Nil(t, afero.WriteFile(appFs.Fs, "/"+course.Path+"/card.png", []byte("test"), os.ModePerm))

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/card", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)
// 		assert.Equal(t, "test", string(body))
// 	})

// 	t.Run("404 (invalid id)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/test/card", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusNotFound, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)
// 		assert.Equal(t, "Course not found", string(body))
// 	})

// 	t.Run("404 (no card)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		course := models.CreateTestCourses(t, db, 1)[0]

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/card", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusNotFound, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)
// 		assert.Equal(t, "Course has no card", string(body))
// 	})

// 	t.Run("404 (card not found)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create a course and set the card path
// 		course := models.CreateTestCourses(t, db, 1)[0]
// 		require.Nil(t, models.UpdateCourseCardPath(db, course, course.Path+"/card.png"))

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/card", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusNotFound, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)
// 		assert.Equal(t, "Course card not found", string(body))
// 	})

// 	t.Run("500 (internal error)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Drop the table
// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Course{}))

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/test/card", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestCourses_GetAssets(t *testing.T) {
// 	t.Run("200 (empty)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 2 courses with 0 assets each
// 		courses := models.CreateTestCourses(t, db, 2)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets", nil))
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
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 2 courses with 5 assets each  (10 assets total)
// 		courses := models.CreateTestCourses(t, db, 2)
// 		assets := models.CreateTestAssets(t, db, courses, 5)
// 		models.CreateTestAttachments(t, db, assets, 2)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData pagination.PaginationResult
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 5, int(respData.TotalItems))

// 		// Unmarshal
// 		assetsResponse := []assetResponse{}
// 		for _, item := range respData.Items {
// 			var asset assetResponse
// 			require.Nil(t, json.Unmarshal(item, &asset))
// 			assetsResponse = append(assetsResponse, asset)
// 		}

// 		// Assert values. By default orderBy is desc so the last inserted asset should be first
// 		assert.Len(t, assetsResponse[0].Attachments, 2)
// 		assert.Equal(t, assets[9].ID, assetsResponse[0].ID)
// 		assert.Equal(t, assets[9].Title, assetsResponse[0].Title)
// 		assert.Equal(t, assets[9].Path, assetsResponse[0].Path)
// 	})

// 	t.Run("200 (orderBy)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 2 courses with 5 assets each  (10 assets total)
// 		courses := models.CreateTestCourses(t, db, 2)
// 		assets := models.CreateTestAssets(t, db, courses, 5)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets/?orderBy=created_at%20asc", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData pagination.PaginationResult
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 5, int(respData.TotalItems))

// 		// Unmarshal
// 		assetsResponse := []assetResponse{}
// 		for _, item := range respData.Items {
// 			var asset assetResponse
// 			require.Nil(t, json.Unmarshal(item, &asset))
// 			assetsResponse = append(assetsResponse, asset)
// 		}

// 		assert.Equal(t, assets[5].ID, assetsResponse[0].ID)
// 		assert.Equal(t, assets[5].Title, assetsResponse[0].Title)
// 		assert.Equal(t, assets[5].Path, assetsResponse[0].Path)
// 	})

// 	t.Run("200 (pagination)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 1 course with 17 assets
// 		courses := models.CreateTestCourses(t, db, 1)
// 		assets := models.CreateTestAssets(t, db, courses, 17)

// 		// Get the first 10 assets
// 		params := url.Values{
// 			"orderBy":                    {"created_at asc"},
// 			pagination.PageQueryParam:    {"1"},
// 			pagination.PerPageQueryParam: {"10"},
// 		}
// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/?"+params.Encode(), nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData pagination.PaginationResult
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 17, int(respData.TotalItems))
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

// 		// Get the next 7 assets
// 		params = url.Values{
// 			"orderBy":                    {"created_at asc"},
// 			pagination.PageQueryParam:    {"2"},
// 			pagination.PerPageQueryParam: {"10"},
// 		}
// 		resp, err = f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/?"+params.Encode(), nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ = io.ReadAll(resp.Body)

// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, 17, int(respData.TotalItems))
// 		assert.Len(t, respData.Items, 7)

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

// 	t.Run("404 (course not found)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/test/assets", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusNotFound, resp.StatusCode)
// 	})

// 	t.Run("500 (course lookup internal error)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Drop the courses table
// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Course{}))

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/test/assets", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
// 	})

// 	t.Run("500 (asset lookup internal error)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create a course
// 		courses := models.CreateTestCourses(t, db, 1)

// 		// Drop the courses table
// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Asset{}))

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestCourses_GetAttachments(t *testing.T) {
// 	t.Run("200 (empty)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 2 courses with 2 assets with 0 attachments
// 		courses := models.CreateTestCourses(t, db, 2)
// 		assets := models.CreateTestAssets(t, db, courses, 2)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets/"+assets[2].ID+"/attachments", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)
// 		assert.Equal(t, "[]", string(body))
// 	})

// 	t.Run("200 (found)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 2 courses with 2 assets with 5 attachments (20 attachments total)
// 		courses := models.CreateTestCourses(t, db, 2)
// 		assets := models.CreateTestAssets(t, db, courses, 2)
// 		attachments := models.CreateTestAttachments(t, db, assets, 5)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets/"+assets[2].ID+"/attachments", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		attachmentsResponse := []attachmentResponse{}
// 		require.Nil(t, json.Unmarshal(body, &attachmentsResponse))

// 		assert.Len(t, attachmentsResponse, 5)
// 		assert.Equal(t, attachments[10].ID, attachmentsResponse[0].ID)
// 		assert.Equal(t, attachments[10].Title, attachmentsResponse[0].Title)
// 		assert.Equal(t, attachments[10].Path, attachmentsResponse[0].Path)
// 	})

// 	t.Run("404 (course not found)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/test/assets/1234/attachments", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusNotFound, resp.StatusCode)
// 	})

// 	t.Run("500 (course lookup internal error)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Drop the courses table
// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Course{}))

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/test/assets/1234/attachments", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)
// 		assert.Contains(t, string(body), "error looking up course")
// 	})

// 	t.Run("404 (asset not found)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create a course
// 		course := models.CreateTestCourses(t, db, 1)[0]

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/1234/attachments", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusNotFound, resp.StatusCode)
// 	})

// 	t.Run("500 (asset lookup internal error)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create a course
// 		course := models.CreateTestCourses(t, db, 1)[0]

// 		// Drop the assets table
// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Asset{}))

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/1234/attachments", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)
// 		assert.Contains(t, string(body), "error looking up asset")
// 	})

// 	t.Run("400 (invalid asset for course)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create 2 courses with 2 assets with 5 attachments (20 attachments total)
// 		courses := models.CreateTestCourses(t, db, 2)
// 		assets := models.CreateTestAssets(t, db, courses, 2)
// 		models.CreateTestAttachments(t, db, assets, 5)

// 		// Set an asset that belongs to a different course
// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets/"+assets[0].ID+"/attachments", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)
// 		assert.Equal(t, "Asset does not belong to course", string(body))
// 	})

// 	t.Run("500 (attachments lookup internal error)", func(t *testing.T) {
// 		appFs, db, courseScanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

// 		// Create a course with an asset
// 		course := models.CreateTestCourses(t, db, 1)[0]
// 		asset := models.CreateTestAssets(t, db, []*models.Course{course}, 1)[0]

// 		// Drop the courses table
// 		require.Nil(t, db.DB().Migrator().DropTable(&models.Attachment{}))

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/attachments", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
// 		body, _ := io.ReadAll(resp.Body)
// 		assert.Contains(t, string(body), "error looking up attachments")
// 	})
// }
