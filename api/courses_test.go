package api

import (
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

	sq "github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/jobs"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// // // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// // func TestCourses_GetCourses(t *testing.T) {
// // 	t.Run("200 (empty)", func(t *testing.T) {
// // 		appFs, db, cs, _, teardown := setup(t)
// // 		defer teardown(t)

// // 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
// // 		assert.NoError(t, err)
// // 		require.Equal(t, http.StatusOK, resp.StatusCode)

// // 		body, _ := io.ReadAll(resp.Body)

// // 		var respData pagination.PaginationResult
// // 		err = json.Unmarshal(body, &respData)
// // 		require.Nil(t, err)
// // 		assert.Equal(t, 0, int(respData.TotalItems))
// // 		assert.Len(t, respData.Items, 0)
// // 	})

// // 	t.Run("200 (found)", func(t *testing.T) {
// // 		appFs, db, cs, _, teardown := setup(t)
// // 		defer teardown(t)

// // 		// Create 5 courses with 5 assets each
// // 		courses := models.NewTestCourses(t, db, 5)
// // 		models.NewTestAssets(t, db, courses, 5)

// // 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
// // 		assert.NoError(t, err)
// // 		require.Equal(t, http.StatusOK, resp.StatusCode)

// // 		body, _ := io.ReadAll(resp.Body)

// // 		var respData pagination.PaginationResult
// // 		err = json.Unmarshal(body, &respData)
// // 		require.Nil(t, err)
// // 		assert.Equal(t, 5, int(respData.TotalItems))

// // 		// Unmarshal
// // 		var coursesResponse []courseResponse
// // 		for _, item := range respData.Items {
// // 			var course courseResponse
// // 			require.Nil(t, json.Unmarshal(item, &course))
// // 			coursesResponse = append(coursesResponse, course)
// // 		}

// // 		// Assert values. By default orderBy is desc so the last inserted course should be first
// // 		assert.Equal(t, courses[4].ID, coursesResponse[0].ID)
// // 		assert.Equal(t, courses[4].Title, coursesResponse[0].Title)
// // 		assert.Equal(t, courses[4].Path, coursesResponse[0].Path)

// // 		// Assert assets are not included
// // 		assert.Nil(t, coursesResponse[0].Assets)
// // 	})

// // 	t.Run("200 (orderBy)", func(t *testing.T) {
// // 		appFs, db, cs, _, teardown := setup(t)
// // 		defer teardown(t)

// // 		// Create 5 courses
// // 		courses := models.NewTestCourses(t, db, 5)

// // 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/?orderBy=created_at%20asc", nil))
// // 		assert.NoError(t, err)
// // 		require.Equal(t, http.StatusOK, resp.StatusCode)

// // 		body, _ := io.ReadAll(resp.Body)

// // 		var respData pagination.PaginationResult
// // 		err = json.Unmarshal(body, &respData)
// // 		require.Nil(t, err)
// // 		assert.Equal(t, 5, int(respData.TotalItems))

// // 		// Unmarshal
// // 		var coursesResponse []courseResponse
// // 		for _, item := range respData.Items {
// // 			var course courseResponse
// // 			require.Nil(t, json.Unmarshal(item, &course))
// // 			coursesResponse = append(coursesResponse, course)
// // 		}

// // 		assert.Equal(t, courses[0].ID, coursesResponse[0].ID)
// // 		assert.Equal(t, courses[0].Title, coursesResponse[0].Title)
// // 		assert.Equal(t, courses[0].Path, coursesResponse[0].Path)
// // 	})

// // 	t.Run("200 (pagination)", func(t *testing.T) {
// // 		appFs, db, cs, _, teardown := setup(t)
// // 		defer teardown(t)

// // 		// Create 17 courses
// // 		courses := models.NewTestCourses(t, db, 17)

// // 		// Get the first 10 courses
// // 		params := url.Values{
// // 			"orderBy":                    {"created_at asc"},
// // 			pagination.PageQueryParam:    {"1"},
// // 			pagination.PerPageQueryParam: {"10"},
// // 		}
// // 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/?"+params.Encode(), nil))
// // 		assert.NoError(t, err)
// // 		require.Equal(t, http.StatusOK, resp.StatusCode)

// // 		body, _ := io.ReadAll(resp.Body)

// // 		var respData pagination.PaginationResult
// // 		err = json.Unmarshal(body, &respData)
// // 		require.Nil(t, err)
// // 		assert.Equal(t, 17, int(respData.TotalItems))
// // 		assert.Len(t, respData.Items, 10)

// // 		// Unmarshal
// // 		var coursesResponse []courseResponse
// // 		for _, item := range respData.Items {
// // 			var course courseResponse
// // 			require.Nil(t, json.Unmarshal(item, &course))
// // 			coursesResponse = append(coursesResponse, course)
// // 		}

// // 		assert.Equal(t, courses[0].ID, coursesResponse[0].ID)
// // 		assert.Equal(t, courses[0].Title, coursesResponse[0].Title)
// // 		assert.Equal(t, courses[0].Path, coursesResponse[0].Path)

// // 		// Get the next 7 courses
// // 		params = url.Values{
// // 			"orderBy":                    {"created_at asc"},
// // 			pagination.PageQueryParam:    {"2"},
// // 			pagination.PerPageQueryParam: {"10"},
// // 		}
// // 		resp, err = f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/?"+params.Encode(), nil))
// // 		assert.NoError(t, err)
// // 		require.Equal(t, http.StatusOK, resp.StatusCode)

// // 		body, _ = io.ReadAll(resp.Body)

// // 		err = json.Unmarshal(body, &respData)
// // 		require.Nil(t, err)
// // 		assert.Equal(t, 17, int(respData.TotalItems))
// // 		assert.Len(t, respData.Items, 7)

// // 		// Unmarshal
// // 		coursesResponse = []courseResponse{}
// // 		for _, item := range respData.Items {
// // 			var course courseResponse
// // 			require.Nil(t, json.Unmarshal(item, &course))
// // 			coursesResponse = append(coursesResponse, course)
// // 		}

// // 		assert.Equal(t, courses[10].ID, coursesResponse[0].ID)
// // 		assert.Equal(t, courses[10].Title, coursesResponse[0].Title)
// // 		assert.Equal(t, courses[10].Path, coursesResponse[0].Path)
// // 	})

// // 	t.Run("200 (expand)", func(t *testing.T) {
// // 		appFs, db, cs, _, teardown := setup(t)
// // 		defer teardown(t)

// // 		// Create 5 courses with 5 assets each
// // 		courses := models.NewTestCourses(t, db, 5)
// // 		models.NewTestAssets(t, db, courses, 5)

// // 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/?orderBy=created_at%20asc&expand=true", nil))
// // 		assert.NoError(t, err)
// // 		require.Equal(t, http.StatusOK, resp.StatusCode)

// // 		body, _ := io.ReadAll(resp.Body)

// // 		var respData pagination.PaginationResult
// // 		err = json.Unmarshal(body, &respData)
// // 		require.Nil(t, err)
// // 		assert.Equal(t, 5, int(respData.TotalItems))

// // 		// Unmarshal
// // 		coursesResponse := []courseResponse{}
// // 		for _, item := range respData.Items {
// // 			var course courseResponse
// // 			require.Nil(t, json.Unmarshal(item, &course))
// // 			coursesResponse = append(coursesResponse, course)
// // 		}

// // 		assert.Equal(t, courses[0].ID, coursesResponse[0].ID)
// // 		assert.Equal(t, courses[0].Title, coursesResponse[0].Title)
// // 		assert.Equal(t, courses[0].Path, coursesResponse[0].Path)

// // 		// Assert the assets
// // 		require.NotNil(t, coursesResponse[0].Assets)
// // 		assert.Len(t, coursesResponse[0].Assets, 5)
// // 	})

// // 	t.Run("200 (started)", func(t *testing.T) {
// // 		appFs, db, cs, _, teardown := setup(t)
// // 		defer teardown(t)

// // 		// Create 2 courses
// // 		courses := models.NewTestCourses(t, db, 2)
// // 		assets := models.NewTestAssets(t, db, courses, 2)

// // 		// For the first asset, set the progress and mark as completed
// // 		_, err := models.UpdateAssetProgress(context.Background(), db, assets[0].ID, 50)
// // 		require.Nil(t, err)
// // 		_, err = models.UpdateAssetCompleted(context.Background(), db, assets[0].ID, true)
// // 		require.Nil(t, err)

// // 		// ------------------
// // 		// `started` not defined
// // 		// ------------------
// // 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
// // 		assert.NoError(t, err)
// // 		require.Equal(t, http.StatusOK, resp.StatusCode)

// // 		respData := paginatedBody(t, resp.Body)
// // 		assert.Equal(t, 2, int(respData.TotalItems))

// // 		for _, item := range respData.Items {
// // 			var course courseResponse
// // 			require.Nil(t, json.Unmarshal(item, &course))

// // 			fmt.Printf("course: %+v\n", course)
// // 		}

// // 		// ------------------
// // 		// `started` is true
// // 		// ------------------
// // 		resp, err = f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/?started=true", nil))
// // 		assert.NoError(t, err)
// // 		require.Equal(t, http.StatusOK, resp.StatusCode)

// // 		respData = paginatedBody(t, resp.Body)
// // 		assert.Equal(t, 1, int(respData.TotalItems))

// // 		// Unmarshal
// // 		coursesResponse := unmarshalCourses(t, respData.Items)
// // 		assert.Equal(t, courses[0].ID, coursesResponse[0].ID)
// // 		assert.Equal(t, 50, coursesResponse[0].Percent)

// // 		// ------------------
// // 		// `started` is false
// // 		// ------------------
// // 		resp, err = f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/?started=false", nil))
// // 		assert.NoError(t, err)
// // 		require.Equal(t, http.StatusOK, resp.StatusCode)

// // 		respData = paginatedBody(t, resp.Body)
// // 		require.Equal(t, 1, int(respData.TotalItems))

// // 		// Unmarshal (Note: orderBy is desc by default)
// // 		coursesResponse = unmarshalCourses(t, respData.Items)
// // 		assert.Equal(t, courses[1].ID, coursesResponse[0].ID)
// // 		assert.Equal(t, 0, coursesResponse[0].Percent)
// // 	})

// // 	t.Run("500 (internal error)", func(t *testing.T) {
// // 		appFs, db, cs, _, teardown := setup(t)
// // 		defer teardown(t)

// // 		// Drop the courses table
// // 		_, err := db.DB().NewDropTable().Model(&models.Course{}).Exec(context.Background())
// // 		require.Nil(t, err)

// // 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
// // 		assert.NoError(t, err)
// // 		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
// // 	})
// // }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetCourse(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 3, false, 0, 0)

		status, body, err := coursesRequestHelper(t, appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[2].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var courseResp courseResponse
		err = json.Unmarshal(body, &courseResp)
		require.Nil(t, err)
		assert.Equal(t, workingData[2].ID, courseResp.ID)
	})

	t.Run("404 (not found)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		status, _, err := coursesRequestHelper(t, appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableCourses())
		require.Nil(t, err)

		status, _, err := coursesRequestHelper(t, appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_Create(t *testing.T) {
	t.Run("201 (created)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, nil, 1, false, 0, 0)
		appFs.Fs.MkdirAll(workingData[0].Path, os.ModePerm)

		postData := fmt.Sprintf(`{"title": "%s", "path": "%s" }`, workingData[0].Title, workingData[0].Path)
		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(postData))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		var courseResp courseResponse
		err = json.Unmarshal(body, &courseResp)
		require.Nil(t, err)
		assert.NotNil(t, courseResp.ID)
		assert.Equal(t, workingData[0].Title, courseResp.Title)
		assert.Equal(t, workingData[0].Path, courseResp.Path)
	})

	t.Run("400 (bind error)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		assert.Contains(t, string(body), "error parsing data")
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		// ----------------------------
		// Missing title
		// ----------------------------
		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		assert.Contains(t, string(body), "a title and path are required")

		// ----------------------------
		// Missing path
		// ----------------------------
		req = httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": "course 1", "path": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		assert.Contains(t, string(body), "a title and path are required")

		// ----------------------------
		// Invalid path
		// ----------------------------
		req = httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": "course 1", "path": "/test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		assert.Contains(t, string(body), "invalid course path")
	})

	t.Run("400 (existing course)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		coursePath := "/course 1/"
		appFs.Fs.MkdirAll(coursePath, os.ModePerm)

		postData := fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(postData))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, _, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		// Create the course again
		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		assert.Contains(t, string(body), "a course with this path already exists ")
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		// Drop the courses table
		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableCourses())
		require.Nil(t, err)

		coursePath := "/course 1/"
		appFs.Fs.MkdirAll(coursePath, os.ModePerm)

		postData := fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(postData))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		assert.Contains(t, string(body), "error creating course")
	})

	t.Run("500 (scan error)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		// Drop scan table
		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableScans())
		require.Nil(t, err)

		coursePath := "/course 1/"
		appFs.Fs.MkdirAll(coursePath, os.ModePerm)

		postData := fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(postData))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		assert.Contains(t, string(body), "error creating scan job")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_DeleteCourse(t *testing.T) {
	t.Run("204 (deleted)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 3, true, 3, 3)

		// ----------------------------
		// Delete course 3
		// ----------------------------
		status, _, err := coursesRequestHelper(t, appFs, db, cs, httptest.NewRequest(http.MethodDelete, "/api/courses/"+workingData[2].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)

		_, err = models.GetCourse(db, workingData[2].ID)
		assert.ErrorIs(t, err, sql.ErrNoRows)

		// ----------------------------
		// Cascades
		// ----------------------------

		// Scan
		_, err = models.GetScan(db, workingData[2].ID)
		assert.ErrorIs(t, err, sql.ErrNoRows)

		// Assets
		count, err := models.CountAssets(db, &database.DatabaseParams{Where: sq.Eq{models.TableAssets() + ".course_id": workingData[2].ID}})
		require.Nil(t, err)
		assert.Zero(t, count)

		// Attachments
		count, err = models.CountAttachments(db, &database.DatabaseParams{Where: sq.Eq{models.TableAttachments() + ".course_id": workingData[2].ID}})
		require.Nil(t, err)
		assert.Zero(t, count)
	})

	t.Run("204 (not found)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		status, _, err := coursesRequestHelper(t, appFs, db, cs, httptest.NewRequest(http.MethodDelete, "/api/courses/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		// Drop the table
		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableCourses())
		require.Nil(t, err)

		status, _, err := coursesRequestHelper(t, appFs, db, cs, httptest.NewRequest(http.MethodDelete, "/api/courses/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_Card(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 1, false, 0, 0)

		updatedCourse, err := models.UpdateCourseCardPath(db, workingData[0].ID, "/"+workingData[0].Path+"/card.png")
		require.Nil(t, err)

		// Create the card
		appFs.Fs.MkdirAll("/"+updatedCourse.Path, os.ModePerm)
		require.Nil(t, afero.WriteFile(appFs.Fs, "/"+updatedCourse.Path+"/card.png", []byte("test"), os.ModePerm))

		status, body, err := coursesRequestHelper(t, appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/"+updatedCourse.ID+"/card", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "test", string(body))
	})

	t.Run("404 (invalid id)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		status, body, err := coursesRequestHelper(t, appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/test/card", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
		assert.Equal(t, "Course not found", string(body))
	})

	t.Run("404 (no card)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 1, false, 0, 0)

		status, body, err := coursesRequestHelper(t, appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/card", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
		assert.Equal(t, "Course has no card", string(body))
	})

	t.Run("404 (card not found)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 1, false, 0, 0)

		updatedCourse, err := models.UpdateCourseCardPath(db, workingData[0].ID, workingData[0].Path+"/card.png")
		require.Nil(t, err)

		status, body, err := coursesRequestHelper(t, appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/"+updatedCourse.ID+"/card", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
		assert.Equal(t, "Course card not found", string(body))
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		// Drop the table
		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableCourses())
		require.Nil(t, err)

		status, _, err := coursesRequestHelper(t, appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/test/card", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetAssets(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 2, false, 0, 0)

		status, body, err := coursesRequestHelper(t, appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[1].ID+"/assets", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := assetsUnmarshalHelper(t, body)
		assert.Zero(t, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 0)
	})

	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 2, false, 3, 0)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[1].ID+"/assets/?orderBy=created_at%20asc", nil)
		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := assetsUnmarshalHelper(t, body)
		assert.Equal(t, 3, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 3)

		assert.Equal(t, workingData[1].Assets[0].ID, assetsResp[0].ID)
		assert.Equal(t, workingData[1].Assets[1].ID, assetsResp[1].ID)
		assert.Equal(t, workingData[1].Assets[2].ID, assetsResp[2].ID)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 2, false, 3, 0)

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[1].ID+"/assets/?orderBy=created_at%20asc", nil)
		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := assetsUnmarshalHelper(t, body)
		assert.Equal(t, 3, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 3)

		assert.Equal(t, workingData[1].Assets[0].ID, assetsResp[0].ID)
		assert.Equal(t, workingData[1].Assets[1].ID, assetsResp[1].ID)
		assert.Equal(t, workingData[1].Assets[2].ID, assetsResp[2].ID)

		// ----------------------------
		// CREATED_AT DESC
		// ----------------------------
		req = httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[1].ID+"/assets/?orderBy=created_at%20desc", nil)
		status, body, err = coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp = assetsUnmarshalHelper(t, body)
		assert.Equal(t, 3, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 3)

		assert.Equal(t, workingData[1].Assets[2].ID, assetsResp[0].ID)
		assert.Equal(t, workingData[1].Assets[1].ID, assetsResp[1].ID)
		assert.Equal(t, workingData[1].Assets[0].ID, assetsResp[2].ID)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 1, false, 17, 0)

		// ------------------------
		// Get the first page (10 assets)
		// ------------------------
		params := url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/?"+params.Encode(), nil)
		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := assetsUnmarshalHelper(t, body)
		assert.Equal(t, 17, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 10)

		// Check the last asset in the paginated response
		assert.Equal(t, workingData[0].Assets[9].ID, assetsResp[9].ID)

		// ------------------------
		// Get the second page (7 assets)
		// ------------------------
		params = url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}

		req = httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/?"+params.Encode(), nil)
		status, body, err = coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp = assetsUnmarshalHelper(t, body)
		assert.Equal(t, 17, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 7)

		// Check the last asset in the paginated response
		assert.Equal(t, workingData[0].Assets[16].ID, assetsResp[6].ID)
	})

	t.Run("404 (course not found)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		status, _, err := coursesRequestHelper(t, appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/test/assets", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (course internal error)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableCourses())
		require.Nil(t, err)

		status, _, err := coursesRequestHelper(t, appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/test/assets", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})

	t.Run("500 (asset internal error)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 1, false, 0, 0)

		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableAssets())
		require.Nil(t, err)

		status, _, err := coursesRequestHelper(t, appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetAsset(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 2, false, 3, 0)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/"+workingData[0].Assets[1].ID, nil)
		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp assetResponse
		err = json.Unmarshal(body, &assetResp)
		require.Nil(t, err)

		assert.Equal(t, workingData[0].Assets[1].ID, assetResp.ID)
		assert.Equal(t, workingData[0].Assets[1].Title, assetResp.Title)
		assert.Equal(t, workingData[0].Assets[1].Path, assetResp.Path)
	})

	t.Run("400 (invalid asset for course)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 2, false, 3, 0)

		// Request an asset that does not belong to the course
		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/"+workingData[1].Assets[0].ID, nil)
		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, "Asset does not belong to course", string(body))
	})

	t.Run("404 (course not found)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		status, _, err := coursesRequestHelper(t, appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/test_course/assets/test_asset", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("404 (asset not found)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 2, false, 0, 0)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[1].ID+"/assets/test_asset", nil)
		status, _, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (course internal error)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		// Drop the courses table
		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableCourses())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/test_course/assets/test_asset", nil)
		status, _, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})

	t.Run("500 (asset internal error)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 1, false, 0, 0)

		// Drop the assets table
		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableAssets())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/test_asset", nil)
		status, _, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)

	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetAttachments(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 2, false, 2, 0)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[1].ID+"/assets/"+workingData[1].Assets[1].ID+"/attachments", nil)
		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := attachmentsUnmarshalHelper(t, body)
		assert.Zero(t, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 0)
	})

	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 2, false, 2, 3)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[1].ID+"/assets/"+workingData[1].Assets[0].ID+"/attachments?orderBy=created_at%20asc", nil)
		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentResp := attachmentsUnmarshalHelper(t, body)
		assert.Equal(t, 3, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 3)

		assert.Equal(t, workingData[1].Assets[0].Attachments[0].ID, attachmentResp[0].ID)
		assert.Equal(t, workingData[1].Assets[0].Attachments[1].ID, attachmentResp[1].ID)
		assert.Equal(t, workingData[1].Assets[0].Attachments[2].ID, attachmentResp[2].ID)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 2, false, 2, 3)

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[1].ID+"/assets/"+workingData[1].Assets[0].ID+"/attachments?orderBy=created_at%20asc", nil)
		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentResp := attachmentsUnmarshalHelper(t, body)
		assert.Equal(t, 3, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 3)

		assert.Equal(t, workingData[1].Assets[0].Attachments[0].ID, attachmentResp[0].ID)
		assert.Equal(t, workingData[1].Assets[0].Attachments[1].ID, attachmentResp[1].ID)
		assert.Equal(t, workingData[1].Assets[0].Attachments[2].ID, attachmentResp[2].ID)

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		req = httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[1].ID+"/assets/"+workingData[1].Assets[0].ID+"/attachments?orderBy=created_at%20desc", nil)
		status, body, err = coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentResp = attachmentsUnmarshalHelper(t, body)
		assert.Equal(t, 3, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 3)

		assert.Equal(t, workingData[1].Assets[0].Attachments[2].ID, attachmentResp[0].ID)
		assert.Equal(t, workingData[1].Assets[0].Attachments[1].ID, attachmentResp[1].ID)
		assert.Equal(t, workingData[1].Assets[0].Attachments[0].ID, attachmentResp[2].ID)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 1, false, 1, 17)

		// ----------------------------
		// Get the first page (10 attachments)
		// ----------------------------
		params := url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/"+workingData[0].Assets[0].ID+"/attachments?"+params.Encode(), nil)
		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentResp := attachmentsUnmarshalHelper(t, body)
		assert.Equal(t, 17, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 10)

		// Check the last attachment in the paginated response
		assert.Equal(t, workingData[0].Assets[0].Attachments[9].ID, attachmentResp[9].ID)

		// ----------------------------
		// Get the second page (7 attachments)
		// ----------------------------
		params = url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}
		req = httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/"+workingData[0].Assets[0].ID+"/attachments?"+params.Encode(), nil)
		status, body, err = coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentResp = attachmentsUnmarshalHelper(t, body)
		assert.Equal(t, 17, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 7)

		// Check the last attachment in the paginated response
		assert.Equal(t, workingData[0].Assets[0].Attachments[16].ID, attachmentResp[6].ID)
	})

	t.Run("404 (course not found)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/test_course/assets/test_asset/attachments", nil)
		status, _, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)

	})

	t.Run("500 (course lookup internal error)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		// Drop the courses table
		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableCourses())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/test_course/assets/test_asset/attachments", nil)
		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		assert.Contains(t, string(body), "error looking up course")
	})

	t.Run("404 (asset not found)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 1, false, 0, 0)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/1234/attachments", nil)
		status, _, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)

	})

	t.Run("500 (asset lookup internal error)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 1, false, 0, 0)

		// Drop the assets table
		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableAssets())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/1234/attachments", nil)
		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		assert.Contains(t, string(body), "error looking up asset")
	})

	t.Run("400 (invalid asset for course)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 2, false, 2, 5)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/"+workingData[1].Assets[0].ID+"/attachments", nil)
		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, "Asset does not belong to course", string(body))
	})

	t.Run("500 (attachments lookup internal error)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 1, false, 1, 0)

		// Drop the attachments table
		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableAttachments())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/"+workingData[0].Assets[0].ID+"/attachments", nil)
		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		assert.Contains(t, string(body), "error looking up asset")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetAssetAttachment(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 2, false, 5, 3)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/"+workingData[0].Assets[0].ID+"/attachments/"+workingData[0].Assets[0].Attachments[2].ID, nil)
		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var respData attachmentResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)

		assert.Equal(t, workingData[0].Assets[0].Attachments[2].ID, respData.ID)
	})

	t.Run("400 (invalid asset for course)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 2, false, 5, 0)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/"+workingData[1].Assets[0].ID+"/attachments/test_attachment", nil)
		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, "Asset does not belong to course", string(body))
	})

	t.Run("400 (invalid attachment for course)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 2, false, 5, 3)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/"+workingData[0].Assets[0].ID+"/attachments/"+workingData[0].Assets[1].Attachments[2].ID, nil)
		status, body, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, "Attachment does not belong to asset", string(body))
	})

	t.Run("404 (course not found)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/test_course/assets/test_asset/attachments/test_attachment", nil)
		status, _, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("404 (asset not found)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 2, false, 0, 0)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/test_asset/attachments/test_attachment", nil)
		status, _, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("404 (attachment not found)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 2, false, 5, 0)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/"+workingData[0].Assets[2].ID+"/attachments/test_attachment", nil)
		status, _, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (course internal error)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 1, false, 0, 0)

		// Drop the courses table
		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableCourses())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/test_asset/attachments/test_attachment", nil)
		status, _, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})

	t.Run("500 (asset internal error)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 1, false, 0, 0)

		// Drop the assets table
		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableAssets())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/test_asset/attachments/test_attachment", nil)
		status, _, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})

	t.Run("500 (attachments internal error)", func(t *testing.T) {
		appFs, db, cs, _, teardown := setup(t)
		defer teardown(t)

		workingData := models.NewTestData(t, db, 1, false, 1, 0)

		// Drop the attachments table
		_, err := db.DB().Exec("DROP TABLE IF EXISTS " + models.TableAttachments())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+workingData[0].ID+"/assets/"+workingData[0].Assets[0].ID+"/attachments/test_attachment", nil)
		status, _, err := coursesRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPERS
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func coursesRequestHelper(t *testing.T, appFs *appFs.AppFs, db database.Database, cs *jobs.CourseScanner, req *http.Request) (int, []byte, error) {
	f := fiber.New()
	bindCoursesApi(f.Group("/api"), appFs, db, cs)

	resp, err := f.Test(req)
	if err != nil {
		return -1, nil, err
	}

	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, b, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func coursesUnmarshalHelper(t *testing.T, body []byte) (pagination.PaginationResult, []courseResponse) {
	var respData pagination.PaginationResult
	err := json.Unmarshal(body, &respData)
	require.Nil(t, err)

	var coursesResponse []courseResponse
	for _, item := range respData.Items {
		var course courseResponse
		require.Nil(t, json.Unmarshal(item, &course))
		coursesResponse = append(coursesResponse, course)
	}

	return respData, coursesResponse
}
