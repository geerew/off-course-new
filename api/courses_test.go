package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetCourses(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
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
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create 5 courses with 5 assets each
		courses := models.NewTestCourses(t, db, 5)
		models.NewTestAssets(t, db, courses, 5)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 5, int(respData.TotalItems))

		// Unmarshal
		var coursesResponse []courseResponse
		for _, item := range respData.Items {
			var course courseResponse
			require.Nil(t, json.Unmarshal(item, &course))
			coursesResponse = append(coursesResponse, course)
		}

		// Assert values. By default orderBy is desc so the last inserted course should be first
		assert.Equal(t, courses[4].ID, coursesResponse[0].ID)
		assert.Equal(t, courses[4].Title, coursesResponse[0].Title)
		assert.Equal(t, courses[4].Path, coursesResponse[0].Path)

		// Assert assets are not included
		assert.Nil(t, coursesResponse[0].Assets)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create 5 courses
		courses := models.NewTestCourses(t, db, 5)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/?orderBy=created_at%20asc", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 5, int(respData.TotalItems))

		// Unmarshal
		var coursesResponse []courseResponse
		for _, item := range respData.Items {
			var course courseResponse
			require.Nil(t, json.Unmarshal(item, &course))
			coursesResponse = append(coursesResponse, course)
		}

		assert.Equal(t, courses[0].ID, coursesResponse[0].ID)
		assert.Equal(t, courses[0].Title, coursesResponse[0].Title)
		assert.Equal(t, courses[0].Path, coursesResponse[0].Path)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create 17 courses
		courses := models.NewTestCourses(t, db, 17)

		// Get the first 10 courses
		params := url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}
		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/?"+params.Encode(), nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 17, int(respData.TotalItems))
		assert.Len(t, respData.Items, 10)

		// Unmarshal
		var coursesResponse []courseResponse
		for _, item := range respData.Items {
			var course courseResponse
			require.Nil(t, json.Unmarshal(item, &course))
			coursesResponse = append(coursesResponse, course)
		}

		assert.Equal(t, courses[0].ID, coursesResponse[0].ID)
		assert.Equal(t, courses[0].Title, coursesResponse[0].Title)
		assert.Equal(t, courses[0].Path, coursesResponse[0].Path)

		// Get the next 7 courses
		params = url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}
		resp, err = f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/?"+params.Encode(), nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ = io.ReadAll(resp.Body)

		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 17, int(respData.TotalItems))
		assert.Len(t, respData.Items, 7)

		// Unmarshal
		coursesResponse = []courseResponse{}
		for _, item := range respData.Items {
			var course courseResponse
			require.Nil(t, json.Unmarshal(item, &course))
			coursesResponse = append(coursesResponse, course)
		}

		assert.Equal(t, courses[10].ID, coursesResponse[0].ID)
		assert.Equal(t, courses[10].Title, coursesResponse[0].Title)
		assert.Equal(t, courses[10].Path, coursesResponse[0].Path)
	})

	t.Run("200 (expand)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create 5 courses with 5 assets each
		courses := models.NewTestCourses(t, db, 5)
		models.NewTestAssets(t, db, courses, 5)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/?orderBy=created_at%20asc&expand=true", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 5, int(respData.TotalItems))

		// Unmarshal
		coursesResponse := []courseResponse{}
		for _, item := range respData.Items {
			var course courseResponse
			require.Nil(t, json.Unmarshal(item, &course))
			coursesResponse = append(coursesResponse, course)
		}

		assert.Equal(t, courses[0].ID, coursesResponse[0].ID)
		assert.Equal(t, courses[0].Title, coursesResponse[0].Title)
		assert.Equal(t, courses[0].Path, coursesResponse[0].Path)

		// Assert the assets
		require.NotNil(t, coursesResponse[0].Assets)
		assert.Len(t, coursesResponse[0].Assets, 5)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Drop the courses table
		_, err := db.DB().NewDropTable().Model(&models.Course{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetCourse(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create 5 courses with 5 asserts each
		courses := models.NewTestCourses(t, db, 5)
		models.NewTestAssets(t, db, courses, 5)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[2].ID, nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData courseResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, courses[2].ID, respData.ID)
		assert.Equal(t, courses[2].Title, respData.Title)
		assert.Equal(t, courses[2].Path, respData.Path)
		assert.Nil(t, respData.Assets)
	})

	t.Run("200 (expand)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		courses := models.NewTestCourses(t, db, 5)
		assets := models.NewTestAssets(t, db, courses, 5)
		models.NewTestAttachments(t, db, assets, 3)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[2].ID+"?expand=true", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData courseResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, courses[2].ID, respData.ID)
		assert.Equal(t, courses[2].Title, respData.Title)
		assert.Equal(t, courses[2].Path, respData.Path)
		require.Len(t, respData.Assets, 5)
		assert.Len(t, respData.Assets[0].Attachments, 3)
	})

	t.Run("404 (not found)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/test", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Drop the table
		_, err := db.DB().NewDropTable().Model(&models.Course{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/test", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_Create(t *testing.T) {
	t.Run("201 (created)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create the course path
		coursePath := "/course 1/"
		appFs.Fs.MkdirAll(coursePath, os.ModePerm)

		postData := fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(postData))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var respData courseResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.NotNil(t, respData.ID)
		assert.Equal(t, "course 1", respData.Title)
		assert.Equal(t, coursePath, respData.Path)
	})

	t.Run("400 (bind error)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "error parsing data")
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Missing title
		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "a title and path are required")

		// Missing path
		req = httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": "course 1", "path": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err = f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, _ = io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "a title and path are required")

		// Invalid path
		req = httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": "course 1", "path": "/test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err = f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, _ = io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "invalid course path")
	})

	t.Run("400 (existing course)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create the course path
		coursePath := "/course 1/"
		appFs.Fs.MkdirAll(coursePath, os.ModePerm)

		postData := fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(postData))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		// Do it again
		postData = fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
		req = httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(postData))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err = f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "a course with this path already exists ")
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		_, err := db.DB().NewDropTable().Model(&models.Course{}).Exec(context.Background())
		require.Nil(t, err)

		// Create the course path
		coursePath := "/course 1/"
		appFs.Fs.MkdirAll(coursePath, os.ModePerm)

		postData := fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(postData))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "error creating course")
	})

	t.Run("500 (scan error)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Drop scan table
		_, err := db.DB().NewDropTable().Model(&models.Scan{}).Exec(context.Background())
		require.Nil(t, err)

		// Create the course path
		coursePath := "/course 1/"
		appFs.Fs.MkdirAll(coursePath, os.ModePerm)

		postData := fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(postData))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "error creating scan job")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_DeleteCourse(t *testing.T) {
	t.Run("204 (deleted)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create 5 courses
		courses := models.NewTestCourses(t, db, 5)

		resp, err := f.Test(httptest.NewRequest(http.MethodDelete, "/api/courses/"+courses[2].ID, nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Ensure the course was deleted
		_, err = models.GetCourseById(context.Background(), db, nil, courses[2].ID)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("404 (not found)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		resp, err := f.Test(httptest.NewRequest(http.MethodDelete, "/api/courses/test", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Drop the table
		_, err := db.DB().NewDropTable().Model(&models.Course{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodDelete, "/api/courses/test", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_Card(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create a course and set the card path
		course := models.NewTestCourses(t, db, 1)[0]
		require.Nil(t, models.UpdateCourseCardPath(context.Background(), db, course, "/"+course.Path+"/card.png"))

		// Create the course path
		appFs.Fs.MkdirAll("/"+course.Path, os.ModePerm)
		require.Nil(t, afero.WriteFile(appFs.Fs, "/"+course.Path+"/card.png", []byte("test"), os.ModePerm))

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/card", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "test", string(body))
	})

	t.Run("404 (invalid id)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/test/card", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "Course not found", string(body))
	})

	t.Run("404 (no card)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		course := models.NewTestCourses(t, db, 1)[0]

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/card", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "Course has no card", string(body))
	})

	t.Run("404 (card not found)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create a course and set the card path
		course := models.NewTestCourses(t, db, 1)[0]
		require.Nil(t, models.UpdateCourseCardPath(context.Background(), db, course, course.Path+"/card.png"))

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/card", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "Course card not found", string(body))
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Drop the table
		_, err := db.DB().NewDropTable().Model(&models.Course{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/test/card", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_UpdateCourse(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		course := models.NewTestCourses(t, db, 1)[0]

		// Store the original
		origCourse, err := models.GetCourseById(context.Background(), db, nil, course.ID)
		require.Nil(t, err)

		// Update
		course.Title = "new title"
		course.Path = "/new/path"
		course.Started = true
		course.Finished = true

		data, err := json.Marshal(toCourseResponse([]*models.Course{course})[0])
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/courses/"+course.ID, strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData courseResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)

		assert.Equal(t, origCourse.ID, respData.ID)
		assert.Equal(t, origCourse.Title, respData.Title)
		assert.Equal(t, origCourse.Path, respData.Path)

		// Assert the updated values
		assert.True(t, respData.Started)
		assert.True(t, respData.Finished)
		assert.NotEqual(t, origCourse.UpdatedAt.String(), respData.UpdatedAt.String())
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		req := httptest.NewRequest(http.MethodPut, "/api/courses/test", strings.NewReader(`bob`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("404 (not found)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		req := httptest.NewRequest(http.MethodPut, "/api/courses/test", strings.NewReader(`{"id": "1234567"}`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Drop the table
		_, err := db.DB().NewDropTable().Model(&models.Course{}).Exec(context.Background())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/courses/test", strings.NewReader(`{"id": "1234567"}`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetAssets(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create 2 courses with 0 assets each
		courses := models.NewTestCourses(t, db, 2)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets", nil))
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
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create 2 courses with 5 assets each  (10 assets total)
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)
		models.NewTestAttachments(t, db, assets, 2)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets/?orderBy=created_at%20desc", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 5, int(respData.TotalItems))

		// Unmarshal
		assetsResponse := []assetResponse{}
		for _, item := range respData.Items {
			var asset assetResponse
			require.Nil(t, json.Unmarshal(item, &asset))
			assetsResponse = append(assetsResponse, asset)
		}

		// Assert values. We set the orderBy query param to `created_at desc` so the last inserted
		// asset should be first
		assert.Equal(t, assets[9].ID, assetsResponse[0].ID)
		assert.Equal(t, assets[9].Title, assetsResponse[0].Title)
		assert.Equal(t, assets[9].Path, assetsResponse[0].Path)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create 2 courses with 5 assets each  (10 assets total)
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets/?orderBy=created_at%20asc", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 5, int(respData.TotalItems))

		// Unmarshal
		assetsResponse := []assetResponse{}
		for _, item := range respData.Items {
			var asset assetResponse
			require.Nil(t, json.Unmarshal(item, &asset))
			assetsResponse = append(assetsResponse, asset)
		}

		assert.Equal(t, assets[5].ID, assetsResponse[0].ID)
		assert.Equal(t, assets[5].Title, assetsResponse[0].Title)
		assert.Equal(t, assets[5].Path, assetsResponse[0].Path)
	})

	t.Run("200 (expand)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create 5 courses with 5 assets each
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)
		models.NewTestAttachments(t, db, assets, 3)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets/?orderBy=created_at%20asc&expand=true", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 5, int(respData.TotalItems))

		// Unmarshal
		assetsResponse := []assetResponse{}
		for _, item := range respData.Items {
			var asset assetResponse
			require.Nil(t, json.Unmarshal(item, &asset))
			assetsResponse = append(assetsResponse, asset)
		}

		assert.Equal(t, assets[5].ID, assetsResponse[0].ID)
		assert.Equal(t, assets[5].Title, assetsResponse[0].Title)
		assert.Equal(t, assets[5].Path, assetsResponse[0].Path)

		// Assert the attachments
		require.NotNil(t, assetsResponse[0].Attachments)
		assert.Len(t, assetsResponse[0].Attachments, 3)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create 1 course with 17 assets
		courses := models.NewTestCourses(t, db, 1)
		assets := models.NewTestAssets(t, db, courses, 17)

		// Get the first 10 assets
		params := url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}
		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/?"+params.Encode(), nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 17, int(respData.TotalItems))
		assert.Len(t, respData.Items, 10)

		// Unmarshal
		assetsResponse := []assetResponse{}
		for _, item := range respData.Items {
			var asset assetResponse
			require.Nil(t, json.Unmarshal(item, &asset))
			assetsResponse = append(assetsResponse, asset)
		}

		// Assert the last item in the paginated response
		assert.Equal(t, assets[9].ID, assetsResponse[9].ID)
		assert.Equal(t, assets[9].Title, assetsResponse[9].Title)
		assert.Equal(t, assets[9].Path, assetsResponse[9].Path)

		// Get the next 7 assets
		params = url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}
		resp, err = f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/?"+params.Encode(), nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ = io.ReadAll(resp.Body)

		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 17, int(respData.TotalItems))
		assert.Len(t, respData.Items, 7)

		// Unmarshal
		assetsResponse = []assetResponse{}
		for _, item := range respData.Items {
			var asset assetResponse
			require.Nil(t, json.Unmarshal(item, &asset))
			assetsResponse = append(assetsResponse, asset)
		}

		// Assert the last item in the paginated response
		assert.Equal(t, assets[16].ID, assetsResponse[6].ID)
		assert.Equal(t, assets[16].Title, assetsResponse[6].Title)
		assert.Equal(t, assets[16].Path, assetsResponse[6].Path)
	})

	t.Run("404 (course not found)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/test/assets", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("500 (course internal error)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Drop the courses table
		_, err := db.DB().NewDropTable().Model(&models.Course{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/test/assets", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("500 (asset internal error)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create a course
		courses := models.NewTestCourses(t, db, 1)

		// Drop the assets table
		_, err := db.DB().NewDropTable().Model(&models.Asset{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetAsset(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/"+assets[1].ID, nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var respData assetResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)

		assert.Equal(t, assets[1].ID, respData.ID)
		assert.Equal(t, assets[1].Title, respData.Title)
		assert.Equal(t, assets[1].Path, respData.Path)
	})

	t.Run("200 (expand)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)
		models.NewTestAttachments(t, db, assets, 3)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/"+assets[1].ID+"?expand=true", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData assetResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)

		assert.Equal(t, assets[1].ID, respData.ID)
		assert.Equal(t, assets[1].Title, respData.Title)
		assert.Equal(t, assets[1].Path, respData.Path)

		// Assert the attachments
		require.NotNil(t, respData.Attachments)
		assert.Len(t, respData.Attachments, 3)
	})

	t.Run("400 (invalid asset for course)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)

		// Request an asset that does not belong to the course
		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/"+assets[8].ID, nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "Asset does not belong to course", string(body))
	})

	t.Run("404 (course not found)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/test/assets/1234", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("404 (asset not found)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		courses := models.NewTestCourses(t, db, 2)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets/1234", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("500 (course internal error)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		courses := models.NewTestCourses(t, db, 1)

		// Drop the assets table
		_, err := db.DB().NewDropTable().Model(&models.Course{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/1234", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("500 (asset internal error)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		courses := models.NewTestCourses(t, db, 1)

		// Drop the assets table
		_, err := db.DB().NewDropTable().Model(&models.Asset{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/1234", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetAttachments(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create 2 courses with 2 assets with 0 attachments
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 2)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets/"+assets[2].ID+"/attachments", nil))
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
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create 2 courses with 2 assets with 5 attachments
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 2)
		attachments := models.NewTestAttachments(t, db, assets, 5)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets/"+assets[2].ID+"/attachments?orderBy=created_at%20desc", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 5, int(respData.TotalItems))
		assert.Len(t, respData.Items, 5)

		// Unmarshal
		attachmentsResponse := []attachmentResponse{}
		for _, item := range respData.Items {
			var attachment attachmentResponse
			require.Nil(t, json.Unmarshal(item, &attachment))
			attachmentsResponse = append(attachmentsResponse, attachment)
		}

		// Assert values
		assert.Equal(t, attachments[14].ID, attachmentsResponse[0].ID)
		assert.Equal(t, attachments[14].Title, attachmentsResponse[0].Title)
		assert.Equal(t, attachments[14].Path, attachmentsResponse[0].Path)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create 2 courses with 2 assets with 5 attachments
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 2)
		attachments := models.NewTestAttachments(t, db, assets, 5)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets/"+assets[2].ID+"/attachments?orderBy=created_at%20asc", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 5, int(respData.TotalItems))
		assert.Len(t, respData.Items, 5)

		// Unmarshal
		attachmentsResponse := []attachmentResponse{}
		for _, item := range respData.Items {
			var attachment attachmentResponse
			require.Nil(t, json.Unmarshal(item, &attachment))
			attachmentsResponse = append(attachmentsResponse, attachment)
		}

		// Assert values
		assert.Equal(t, attachments[10].ID, attachmentsResponse[0].ID)
		assert.Equal(t, attachments[10].Title, attachmentsResponse[0].Title)
		assert.Equal(t, attachments[10].Path, attachmentsResponse[0].Path)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create 2 courses with 2 assets with 17 attachments
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 2)
		attachments := models.NewTestAttachments(t, db, assets, 17)

		// Get the first 10 attachments
		params := url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}
		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets/"+assets[2].ID+"/attachments?"+params.Encode(), nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 17, int(respData.TotalItems))
		assert.Len(t, respData.Items, 10)

		// Unmarshal
		attachmentsResponse := []attachmentResponse{}
		for _, item := range respData.Items {
			var attachment attachmentResponse
			require.Nil(t, json.Unmarshal(item, &attachment))
			attachmentsResponse = append(attachmentsResponse, attachment)
		}

		// Assert the last item in the paginated response
		assert.Equal(t, attachments[43].ID, attachmentsResponse[9].ID)
		assert.Equal(t, attachments[43].Title, attachmentsResponse[9].Title)
		assert.Equal(t, attachments[43].Path, attachmentsResponse[9].Path)

		// Get the next 7 attachments
		params = url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}
		resp, err = f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets/"+assets[2].ID+"/attachments?"+params.Encode(), nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ = io.ReadAll(resp.Body)

		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 17, int(respData.TotalItems))
		assert.Len(t, respData.Items, 7)

		// Unmarshal
		attachmentsResponse = []attachmentResponse{}
		for _, item := range respData.Items {
			var attachment attachmentResponse
			require.Nil(t, json.Unmarshal(item, &attachment))
			attachmentsResponse = append(attachmentsResponse, attachment)
		}

		assert.Equal(t, attachments[50].ID, attachmentsResponse[6].ID)
		assert.Equal(t, attachments[50].Title, attachmentsResponse[6].Title)
		assert.Equal(t, attachments[50].Path, attachmentsResponse[6].Path)
	})

	t.Run("404 (course not found)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/test/assets/1234/attachments", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("500 (course lookup internal error)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Drop the courses table
		_, err := db.DB().NewDropTable().Model(&models.Course{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/test/assets/1234/attachments", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "error looking up course")
	})

	t.Run("404 (asset not found)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create a course
		course := models.NewTestCourses(t, db, 1)[0]

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/1234/attachments", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("500 (asset lookup internal error)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create a course
		course := models.NewTestCourses(t, db, 1)[0]

		// Drop the assets table
		_, err := db.DB().NewDropTable().Model(&models.Asset{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/1234/attachments", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "error looking up asset")
	})

	t.Run("400 (invalid asset for course)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create 2 courses with 2 assets with 5 attachments
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 2)
		models.NewTestAttachments(t, db, assets, 5)

		// Set an asset that belongs to a different course
		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets/"+assets[0].ID+"/attachments", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "Asset does not belong to course", string(body))
	})

	t.Run("500 (attachments lookup internal error)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		// Create a course with an asset
		course := models.NewTestCourses(t, db, 1)[0]
		asset := models.NewTestAssets(t, db, []*models.Course{course}, 1)[0]

		// Drop the courses table
		_, err := db.DB().NewDropTable().Model(&models.Attachment{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/attachments", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "error looking up attachments")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetAssetAttachment(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)
		attachments := models.NewTestAttachments(t, db, assets, 3)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/"+assets[0].ID+"/attachments/"+attachments[0].ID, nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var respData attachmentResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)

		assert.Equal(t, attachments[0].ID, respData.ID)
		assert.Equal(t, attachments[0].Title, respData.Title)
		assert.Equal(t, attachments[0].Path, respData.Path)
	})

	t.Run("400 (invalid asset for course)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)

		// Request an asset that does not belong to the course
		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/"+assets[8].ID+"/attachments/4321", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "Asset does not belong to course", string(body))
	})

	t.Run("400 (invalid asset for course)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)
		attachments := models.NewTestAttachments(t, db, assets, 3)

		// Request an asset that does not belong to the course
		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/"+assets[0].ID+"/attachments/"+attachments[5].ID, nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "Attachment does not belong to asset", string(body))
	})

	t.Run("404 (course not found)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/test/assets/1234/attachments/4321", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("404 (asset not found)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		courses := models.NewTestCourses(t, db, 2)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/1234/attachments/4321", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("404 (attachment not found)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/"+assets[0].ID+"/attachments/4321", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("500 (course internal error)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		courses := models.NewTestCourses(t, db, 1)

		// Drop the assets table
		_, err := db.DB().NewDropTable().Model(&models.Course{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/1234/attachments/4321", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("500 (asset internal error)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		courses := models.NewTestCourses(t, db, 1)

		// Drop the assets table
		_, err := db.DB().NewDropTable().Model(&models.Asset{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/1234/attachments/4321", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("500 (attachments internal error)", func(t *testing.T) {
		appFs, db, courseScanner, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindCoursesApi(f.Group("/api"), appFs, db, courseScanner)

		courses := models.NewTestCourses(t, db, 1)
		assets := models.NewTestAssets(t, db, courses, 1)

		// Drop the assets table
		_, err := db.DB().NewDropTable().Model(&models.Attachment{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[0].ID+"/assets/"+assets[0].ID+"/attachments/4321", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
