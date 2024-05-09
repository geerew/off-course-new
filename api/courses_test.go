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

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/jobs"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetCourses(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		status, body, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := coursesUnmarshalHelper(t, body)
		require.Zero(t, int(paginationResp.TotalItems))
		require.Zero(t, len(paginationResp.Items))
	})

	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		daos.NewTestBuilder(t).Db(db).Courses(5).Build()

		status, body, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp := coursesUnmarshalHelper(t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, coursesResp, 5)
	})

	t.Run("200 (availability)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(3).Build()
		courseDao := daos.NewCourseDao(db)

		// ----------------------------
		// All unavailable
		// ----------------------------
		status, body, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/?orderBy=created_at%20asc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp := coursesUnmarshalHelper(t, body)
		require.Equal(t, 3, int(paginationResp.TotalItems))
		require.Len(t, coursesResp, 3)

		for _, c := range coursesResp {
			require.False(t, c.Available)
		}

		// ----------------------------
		// Several available
		// ----------------------------
		testData[0].Available = true
		require.Nil(t, courseDao.Update(testData[0].Course))

		testData[2].Available = true
		require.Nil(t, courseDao.Update(testData[2].Course))

		status, body, err = coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/?orderBy=created_at%20asc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp = coursesUnmarshalHelper(t, body)
		require.Equal(t, 3, int(paginationResp.TotalItems))
		require.Len(t, coursesResp, 3)

		require.True(t, coursesResp[0].Available)
		require.False(t, coursesResp[1].Available)
		require.True(t, coursesResp[2].Available)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(5).Build()

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		status, body, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/?orderBy=created_at%20asc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp := coursesUnmarshalHelper(t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, coursesResp, 5)
		require.Equal(t, testData[0].ID, coursesResp[0].ID)

		// ----------------------------
		// CREATED_AT DESC
		// ----------------------------
		status, body, err = coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/?orderBy=created_at%20desc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp = coursesUnmarshalHelper(t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, coursesResp, 5)
		require.Equal(t, testData[4].ID, coursesResp[0].ID)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(17).Build()

		// ----------------------------
		// Get the first page (10 courses)
		// ----------------------------
		params := url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}

		status, body, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp := coursesUnmarshalHelper(t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 10)

		// Check the last asset in the paginated response
		require.Equal(t, testData[9].ID, coursesResp[9].ID)

		// ----------------------------
		// Get the second page (7 courses)
		// ----------------------------
		params = url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}
		status, body, err = coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp = coursesUnmarshalHelper(t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 7)

		// Check the last asset in the paginated response
		require.Equal(t, testData[16].ID, coursesResp[6].ID)
	})

	t.Run("200 (progress)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Assets(2).Build()

		// Set the first course as started (by marking the first asset as started)
		apDao := daos.NewAssetProgressDao(db)
		ap := &models.AssetProgress{
			AssetID:  testData[0].Assets[0].ID,
			VideoPos: 10,
		}
		require.Nil(t, apDao.Update(ap, nil))

		// ------------------
		// not defined
		// ------------------
		status, body, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := coursesUnmarshalHelper(t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 2)

		// ------------------
		// started
		// ------------------
		status, body, err = coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/?progress=started", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp := coursesUnmarshalHelper(t, body)
		require.Equal(t, 1, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 1)
		require.Equal(t, testData[0].ID, coursesResp[0].ID)

		// ------------------
		// Not started
		// ------------------
		status, body, err = coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/?progress=not%20started", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp = coursesUnmarshalHelper(t, body)
		require.Equal(t, 1, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 1)
		require.Equal(t, testData[1].ID, coursesResp[0].ID)

		// ------------------
		// not completed
		// ------------------
		status, body, err = coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/?progress=not%20completed", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ = coursesUnmarshalHelper(t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 2)

		// ------------------
		// Completed
		// ------------------

		// Set the first course as completed by marking the assets as completed
		for _, a := range testData[0].Assets {
			ap := &models.AssetProgress{
				AssetID:   a.ID,
				Completed: true,
			}
			require.Nil(t, apDao.Update(ap, nil))
		}

		status, body, err = coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/?progress=completed", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp = coursesUnmarshalHelper(t, body)
		require.Equal(t, 1, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 1)
		require.Equal(t, testData[0].ID, coursesResp[0].ID)
	})

	t.Run("200 (tags)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		course1 := daos.NewTestBuilder(t).Db(db).Courses([]string{"course 1"}).Tags([]string{"Go", "Data Structures"}).Build()[0]
		course2 := daos.NewTestBuilder(t).Db(db).Courses([]string{"course 2"}).Tags([]string{"Data Structures", "TypeScript", "PHP"}).Build()[0]
		course3 := daos.NewTestBuilder(t).Db(db).Courses([]string{"course 3"}).Tags([]string{"Go", "Data Structures", "PHP"}).Build()[0]

		// ------------------
		// Not defined
		// ------------------
		status, body, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := coursesUnmarshalHelper(t, body)
		require.Equal(t, 3, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 3)

		// ------------------
		// Go
		// ------------------
		status, body, err = coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/?tags=Go&orderBy=title%20asc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp := coursesUnmarshalHelper(t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 2)
		require.Equal(t, course1.ID, coursesResp[0].ID)
		require.Equal(t, course3.ID, coursesResp[1].ID)

		// ------------------
		// Go and Data Structures
		// ------------------
		status, body, err = coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/?tags=Go,Data%20Structures&orderBy=title%20asc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp = coursesUnmarshalHelper(t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 2)
		require.Equal(t, course1.ID, coursesResp[0].ID)
		require.Equal(t, course3.ID, coursesResp[1].ID)

		// ------------------
		// Go, Data Structures and PHP
		// ------------------
		status, body, err = coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/?tags=Go,Data%20Structures,PHP&orderBy=title%20asc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp = coursesUnmarshalHelper(t, body)
		require.Equal(t, 1, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 1)
		require.Equal(t, course3.ID, coursesResp[0].ID)

		// ------------------
		// // Go, Data Structures, PHP and TypeScript
		// ------------------

		status, body, err = coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/?tags=Go,Data%20Structures,PHP,TypeScript&orderBy=title%20asc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ = coursesUnmarshalHelper(t, body)
		require.Equal(t, 0, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 0)

		// ------------------
		// Data Structures
		// ------------------
		status, body, err = coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/?tags=Data%20Structures&orderBy=title%20asc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp = coursesUnmarshalHelper(t, body)
		require.Equal(t, 3, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 3)
		require.Equal(t, course1.ID, coursesResp[0].ID)
		require.Equal(t, course2.ID, coursesResp[1].ID)
		require.Equal(t, course3.ID, coursesResp[2].ID)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		// Drop the courses table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewCourseDao(db).Table())
		require.Nil(t, err)

		status, _, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetCourse(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(3).Build()

		status, body, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[2].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var courseResp courseResponse
		err = json.Unmarshal(body, &courseResp)
		require.Nil(t, err)
		require.Equal(t, testData[2].ID, courseResp.ID)
	})

	t.Run("200 (availability)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(3).Build()

		// ----------------------------
		// Unavailable
		// ----------------------------
		status, body, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[2].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var courseResp courseResponse
		err = json.Unmarshal(body, &courseResp)
		require.Nil(t, err)
		require.Equal(t, testData[2].ID, courseResp.ID)
		require.False(t, courseResp.Available)

		// ----------------------------
		// Available
		// ----------------------------
		courseDao := daos.NewCourseDao(db)
		testData[2].Available = true
		require.Nil(t, courseDao.Update(testData[2].Course))

		status, body, err = coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[2].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		err = json.Unmarshal(body, &courseResp)
		require.Nil(t, err)
		require.Equal(t, testData[2].ID, courseResp.ID)
		require.True(t, courseResp.Available)
	})

	t.Run("404 (not found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		status, _, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewCourseDao(db).Table())
		require.Nil(t, err)

		status, _, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_CreateCourse(t *testing.T) {
	t.Run("201 (created)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Courses(1).Build()
		appFs.Fs.MkdirAll(testData[0].Path, os.ModePerm)

		postData := fmt.Sprintf(`{"title": "%s", "path": "%s" }`, testData[0].Title, testData[0].Path)
		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(postData))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		var courseResp courseResponse
		err = json.Unmarshal(body, &courseResp)
		require.Nil(t, err)
		require.NotNil(t, courseResp.ID)
		require.Equal(t, testData[0].Title, courseResp.Title)
		require.Equal(t, testData[0].Path, courseResp.Path)
		require.True(t, courseResp.Available)
	})

	t.Run("400 (bind error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "error parsing data")
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		// ----------------------------
		// Missing title
		// ----------------------------
		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "a title and path are required")

		// ----------------------------
		// Missing path
		// ----------------------------
		req = httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": "course 1", "path": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "a title and path are required")

		// ----------------------------
		// Invalid path
		// ----------------------------
		req = httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": "course 1", "path": "/test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "invalid course path")
	})

	t.Run("400 (existing course)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		coursePath := "/course 1/"
		appFs.Fs.MkdirAll(coursePath, os.ModePerm)

		postData := fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(postData))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, _, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		// Create the course again
		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "a course with this path already exists ")
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		// Drop the courses table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewCourseDao(db).Table())
		require.Nil(t, err)

		coursePath := "/course 1/"
		appFs.Fs.MkdirAll(coursePath, os.ModePerm)

		postData := fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(postData))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		require.Contains(t, string(body), "error creating course")
	})

	t.Run("500 (scan error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		// Drop scan table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewScanDao(db).Table())
		require.Nil(t, err)

		coursePath := "/course 1/"
		appFs.Fs.MkdirAll(coursePath, os.ModePerm)

		postData := fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(postData))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		require.Contains(t, string(body), "error creating scan job")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_DeleteCourse(t *testing.T) {
	t.Run("204 (deleted)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(3).Assets(3).Attachments(3).Build()

		courseDao := daos.NewCourseDao(db)
		scanDao := daos.NewScanDao(db)
		assetsDao := daos.NewAssetDao(db)
		attachmentsDao := daos.NewAttachmentDao(db)

		// ----------------------------
		// Delete course 3
		// ----------------------------
		status, _, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodDelete, "/api/courses/"+testData[2].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)

		_, err = courseDao.Get(testData[2].ID, nil, nil)
		require.ErrorIs(t, err, sql.ErrNoRows)

		// ----------------------------
		// Cascades
		// ----------------------------

		// Scan
		_, err = scanDao.Get(testData[2].ID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		// Assets
		count, err := assetsDao.Count(&database.DatabaseParams{Where: squirrel.Eq{daos.NewAssetDao(db).Table() + ".course_id": testData[2].ID}})
		require.Nil(t, err)
		require.Zero(t, count)

		// Attachments
		count, err = attachmentsDao.Count(&database.DatabaseParams{Where: squirrel.Eq{daos.NewAttachmentDao(db).Table() + ".course_id": testData[2].ID}})
		require.Nil(t, err)
		require.Zero(t, count)
	})

	t.Run("204 (not found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		status, _, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodDelete, "/api/courses/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		// Drop the table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewCourseDao(db).Table())
		require.Nil(t, err)

		status, _, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodDelete, "/api/courses/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetCard(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Build()
		courseDao := daos.NewCourseDao(db)

		// Update card path
		testData[0].CardPath = "/" + testData[0].Path + "/card.png"
		require.Nil(t, courseDao.Update(testData[0].Course))

		// Create
		appFs.Fs.MkdirAll("/"+testData[0].Path, os.ModePerm)
		require.Nil(t, afero.WriteFile(appFs.Fs, testData[0].CardPath, []byte("test"), os.ModePerm))

		status, body, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/card", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, "test", string(body))
	})

	t.Run("404 (invalid id)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		status, body, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/test/card", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
		require.Equal(t, "Course not found", string(body))
	})

	t.Run("404 (no card)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Build()

		status, body, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/card", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
		require.Equal(t, "Course has no card", string(body))
	})

	t.Run("404 (card not found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Build()
		courseDao := daos.NewCourseDao(db)

		// Update card path
		testData[0].CardPath = "/" + testData[0].Path + "/card.png"
		require.Nil(t, courseDao.Update(testData[0].Course))

		status, body, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/card", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
		require.Equal(t, "Course card not found", string(body))
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		// Drop the table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewCourseDao(db).Table())
		require.Nil(t, err)

		status, _, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/test/card", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetAssets(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Build()

		status, body, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[1].ID+"/assets", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := assetsUnmarshalHelper(t, body)
		require.Zero(t, int(paginationResp.TotalItems))
		require.Zero(t, len(paginationResp.Items))
	})

	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Assets(3).Attachments(2).Build()

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[1].ID+"/assets/?orderBy=created_at%20asc", nil)
		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := assetsUnmarshalHelper(t, body)
		require.Equal(t, 3, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 3)
		require.Equal(t, testData[1].Assets[0].ID, assetsResp[0].ID)
		require.Equal(t, testData[1].Assets[1].ID, assetsResp[1].ID)
		require.Equal(t, testData[1].Assets[2].ID, assetsResp[2].ID)
		require.Nil(t, assetsResp[0].Attachments)

		// ----------------------------
		// Attachments
		// ----------------------------

		req = httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[1].ID+"/assets/?expand=true&orderBy=created_at%20asc", nil)
		status, body, err = coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp = assetsUnmarshalHelper(t, body)
		require.Equal(t, 3, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 3)
		require.Len(t, assetsResp[0].Attachments, 2)

	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Assets(3).Attachments(2).Build()

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[1].ID+"/assets/?orderBy=created_at%20asc", nil)
		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := assetsUnmarshalHelper(t, body)
		require.Equal(t, 3, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 3)
		require.Equal(t, testData[1].Assets[0].ID, assetsResp[0].ID)
		require.Equal(t, testData[1].Assets[1].ID, assetsResp[1].ID)
		require.Equal(t, testData[1].Assets[2].ID, assetsResp[2].ID)

		// ----------------------------
		// CREATED_AT DESC
		// ----------------------------
		req = httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[1].ID+"/assets/?orderBy=created_at%20desc", nil)
		status, body, err = coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp = assetsUnmarshalHelper(t, body)
		require.Equal(t, 3, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 3)
		require.Equal(t, testData[1].Assets[2].ID, assetsResp[0].ID)
		require.Equal(t, testData[1].Assets[1].ID, assetsResp[1].ID)
		require.Equal(t, testData[1].Assets[0].ID, assetsResp[2].ID)

		// ----------------------------
		// CREATED_AT ASC + ATTACHMENTS.TITLE DESC
		// ----------------------------
		attDao := daos.NewAttachmentDao(db)

		req = httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[1].ID+"/assets/?expand=true&orderBy=created_at%20asc,"+attDao.Table()+".created_at%20desc", nil)
		status, body, err = coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp = assetsUnmarshalHelper(t, body)
		require.Equal(t, 3, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 3)
		require.Equal(t, testData[1].Assets[0].ID, assetsResp[0].ID)
		require.Len(t, assetsResp[0].Attachments, 2)
		require.Equal(t, testData[1].Assets[0].Attachments[1].ID, assetsResp[0].Attachments[0].ID)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Assets(17).Build()

		// ------------------------
		// Get the first page (10 assets)
		// ------------------------
		params := url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/?"+params.Encode(), nil)
		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := assetsUnmarshalHelper(t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 10)

		// Check the last asset in the paginated response
		require.Equal(t, testData[0].Assets[9].ID, assetsResp[9].ID)

		// ------------------------
		// Get the second page (7 assets)
		// ------------------------
		params = url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}

		req = httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/?"+params.Encode(), nil)
		status, body, err = coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp = assetsUnmarshalHelper(t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 7)

		// Check the last asset in the paginated response
		require.Equal(t, testData[0].Assets[16].ID, assetsResp[6].ID)
	})

	t.Run("404 (course not found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		status, _, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/test/assets", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (course internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewCourseDao(db).Table())
		require.Nil(t, err)

		status, _, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/test/assets", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})

	t.Run("500 (asset internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Build()

		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewAssetDao(db).Table())
		require.Nil(t, err)

		status, _, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetAsset(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Assets(3).Attachments(2).Build()

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/"+testData[0].Assets[1].ID, nil)
		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp assetResponse
		err = json.Unmarshal(body, &assetResp)
		require.Nil(t, err)
		require.Equal(t, testData[0].Assets[1].ID, assetResp.ID)
		require.Equal(t, testData[0].Assets[1].Title, assetResp.Title)
		require.Equal(t, testData[0].Assets[1].Path, assetResp.Path)
		require.Nil(t, assetResp.Attachments)

		// ----------------------------
		// Attachments
		// ----------------------------
		req = httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/"+testData[0].Assets[1].ID+"/?expand=true", nil)
		status, body, err = coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		err = json.Unmarshal(body, &assetResp)
		require.Nil(t, err)
		require.Equal(t, testData[0].Assets[1].ID, assetResp.ID)
		require.Equal(t, testData[0].Assets[1].Title, assetResp.Title)
		require.Equal(t, testData[0].Assets[1].Path, assetResp.Path)
		require.Len(t, assetResp.Attachments, 2)
	})

	t.Run("400 (invalid asset for course)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Assets(3).Build()

		// Request an asset that does not belong to the course
		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/"+testData[1].Assets[0].ID, nil)
		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Equal(t, "Asset does not belong to course", string(body))
	})

	t.Run("404 (course not found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		status, _, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/test_course/assets/test_asset", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("404 (asset not found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Build()

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[1].ID+"/assets/test_asset", nil)
		status, _, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (course internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		// Drop the courses table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewCourseDao(db).Table())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/test_course/assets/test_asset", nil)
		status, _, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})

	t.Run("500 (asset internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Build()

		// Drop the assets table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewAssetDao(db).Table())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/test_asset", nil)
		status, _, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetAssetAttachments(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Assets(2).Build()

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[1].ID+"/assets/"+testData[1].Assets[1].ID+"/attachments", nil)
		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := attachmentsUnmarshalHelper(t, body)
		require.Zero(t, int(paginationResp.TotalItems))
		require.Zero(t, len(paginationResp.Items))
	})

	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Assets(2).Attachments(3).Build()

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[1].ID+"/assets/"+testData[1].Assets[0].ID+"/attachments?orderBy=created_at%20asc", nil)
		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentResp := attachmentsUnmarshalHelper(t, body)
		require.Equal(t, 3, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 3)

		require.Equal(t, testData[1].Assets[0].Attachments[0].ID, attachmentResp[0].ID)
		require.Equal(t, testData[1].Assets[0].Attachments[1].ID, attachmentResp[1].ID)
		require.Equal(t, testData[1].Assets[0].Attachments[2].ID, attachmentResp[2].ID)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Assets(2).Attachments(3).Build()

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[1].ID+"/assets/"+testData[1].Assets[0].ID+"/attachments?orderBy=created_at%20asc", nil)
		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentResp := attachmentsUnmarshalHelper(t, body)
		require.Equal(t, 3, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 3)

		require.Equal(t, testData[1].Assets[0].Attachments[0].ID, attachmentResp[0].ID)
		require.Equal(t, testData[1].Assets[0].Attachments[1].ID, attachmentResp[1].ID)
		require.Equal(t, testData[1].Assets[0].Attachments[2].ID, attachmentResp[2].ID)

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		req = httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[1].ID+"/assets/"+testData[1].Assets[0].ID+"/attachments?orderBy=created_at%20desc", nil)
		status, body, err = coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentResp = attachmentsUnmarshalHelper(t, body)
		require.Equal(t, 3, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 3)

		require.Equal(t, testData[1].Assets[0].Attachments[2].ID, attachmentResp[0].ID)
		require.Equal(t, testData[1].Assets[0].Attachments[1].ID, attachmentResp[1].ID)
		require.Equal(t, testData[1].Assets[0].Attachments[0].ID, attachmentResp[2].ID)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Assets(1).Attachments(17).Build()

		// ----------------------------
		// Get the first page (10 attachments)
		// ----------------------------
		params := url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/"+testData[0].Assets[0].ID+"/attachments?"+params.Encode(), nil)
		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentResp := attachmentsUnmarshalHelper(t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 10)

		// Check the last attachment in the paginated response
		require.Equal(t, testData[0].Assets[0].Attachments[9].ID, attachmentResp[9].ID)

		// ----------------------------
		// Get the second page (7 attachments)
		// ----------------------------
		params = url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}
		req = httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/"+testData[0].Assets[0].ID+"/attachments?"+params.Encode(), nil)
		status, body, err = coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentResp = attachmentsUnmarshalHelper(t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 7)

		// Check the last attachment in the paginated response
		require.Equal(t, testData[0].Assets[0].Attachments[16].ID, attachmentResp[6].ID)
	})

	t.Run("404 (course not found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/test_course/assets/test_asset/attachments", nil)
		status, _, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)

	})

	t.Run("500 (course lookup internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		// Drop the courses table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewCourseDao(db).Table())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/test_course/assets/test_asset/attachments", nil)
		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		require.Contains(t, string(body), "error looking up course")
	})

	t.Run("404 (asset not found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Build()

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/1234/attachments", nil)
		status, _, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)

	})

	t.Run("500 (asset lookup internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Build()

		// Drop the assets table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewAssetDao(db).Table())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/1234/attachments", nil)
		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		require.Contains(t, string(body), "error looking up asset")
	})

	t.Run("400 (invalid asset for course)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Assets(2).Attachments(5).Build()

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/"+testData[1].Assets[0].ID+"/attachments", nil)
		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Equal(t, "Asset does not belong to course", string(body))
	})

	t.Run("500 (attachments lookup internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Assets(1).Build()

		// Drop the attachments table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewAttachmentDao(db).Table())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/"+testData[0].Assets[0].ID+"/attachments", nil)
		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		require.Contains(t, string(body), "error looking up attachments")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetAssetAttachment(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Assets(5).Attachments(3).Build()

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/"+testData[0].Assets[0].ID+"/attachments/"+testData[0].Assets[0].Attachments[2].ID, nil)
		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var respData attachmentResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)

		require.Equal(t, testData[0].Assets[0].Attachments[2].ID, respData.ID)
	})

	t.Run("400 (invalid asset for course)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Assets(5).Build()

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/"+testData[1].Assets[0].ID+"/attachments/test_attachment", nil)
		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Equal(t, "Asset does not belong to course", string(body))
	})

	t.Run("400 (invalid attachment for course)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Assets(5).Attachments(3).Build()

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/"+testData[0].Assets[0].ID+"/attachments/"+testData[0].Assets[1].Attachments[2].ID, nil)
		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Equal(t, "Attachment does not belong to asset", string(body))
	})

	t.Run("404 (course not found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/test_course/assets/test_asset/attachments/test_attachment", nil)
		status, _, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("404 (asset not found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Build()

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/test_asset/attachments/test_attachment", nil)
		status, _, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("404 (attachment not found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Assets(5).Build()

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/"+testData[0].Assets[2].ID+"/attachments/test_attachment", nil)
		status, _, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (course internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Build()

		// Drop the courses table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewCourseDao(db).Table())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/test_asset/attachments/test_attachment", nil)
		status, _, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})

	t.Run("500 (asset internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Build()

		// Drop the assets table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewAssetDao(db).Table())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/test_asset/attachments/test_attachment", nil)
		status, _, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})

	t.Run("500 (attachments internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Assets(1).Build()

		// Drop the attachments table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewAttachmentDao(db).Table())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/assets/"+testData[0].Assets[0].ID+"/attachments/test_attachment", nil)
		status, _, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetTags(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Build()

		status, body, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[1].ID+"/tags", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var tags []courseTagResponse
		err = json.Unmarshal(body, &tags)
		require.Nil(t, err)
		require.Zero(t, len(tags))
	})

	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(2).Tags([]string{"Go", "C", "JavaScript", "TypeScript", "Java", "Python"}).Build()

		status, body, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[1].ID+"/tags", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var tags []courseTagResponse
		err = json.Unmarshal(body, &tags)
		require.Nil(t, err)

		require.Len(t, tags, 6)
		require.Equal(t, "C", tags[0].Tag)
		require.Equal(t, "TypeScript", tags[5].Tag)
	})

	t.Run("404 (course not found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		status, _, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/test/tags", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (course internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewCourseDao(db).Table())
		require.Nil(t, err)

		status, _, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/test/tags", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})

	t.Run("500 (courses_tags internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Build()

		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewCourseTagDao(db).Table())
		require.Nil(t, err)

		status, _, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[0].ID+"/tags/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_CreateTag(t *testing.T) {
	t.Run("201 (created)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Build()

		req := httptest.NewRequest(http.MethodPost, "/api/courses/"+testData[0].ID+"/tags", strings.NewReader(`{"tag": "Go" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		var courseTagResp courseTagResponse
		err = json.Unmarshal(body, &courseTagResp)
		require.Nil(t, err)
		require.NotNil(t, courseTagResp.ID)
		require.Equal(t, "Go", courseTagResp.Tag)
	})

	t.Run("400 (bind error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Build()

		req := httptest.NewRequest(http.MethodPost, "/api/courses/"+testData[0].ID+"/tags", strings.NewReader(`{`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "error parsing data")
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Build()

		req := httptest.NewRequest(http.MethodPost, "/api/courses/"+testData[0].ID+"/tags", strings.NewReader(`{"tag": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "a tag is required")
	})

	t.Run("400 (existing tag)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Build()

		req := httptest.NewRequest(http.MethodPost, "/api/courses/"+testData[0].ID+"/tags", strings.NewReader(`{"tag": "Go"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, _, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		// Create the tag again
		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "a tag for this course already exists")
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Build()

		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewCourseTagDao(db).Table())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/courses/"+testData[0].ID+"/tags", strings.NewReader(`{"tag": "Go"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := coursesRequestHelper(appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		require.Contains(t, string(body), "error creating course tag")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_DeleteTag(t *testing.T) {
	t.Run("204 (deleted)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(3).Tags([]string{"Go", "C", "JavaScript", "TypeScript", "Java", "Python"}).Build()

		courseTagDao := daos.NewCourseTagDao(db)

		// Delete the third tag from the second course
		status, _, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodDelete, "/api/courses/"+testData[1].ID+"/tags/"+testData[1].Tags[2].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)

		// Asset the tag has been deleted
		tags, err := courseTagDao.List(&database.DatabaseParams{Where: squirrel.Eq{"course_id": testData[1].ID}}, nil)
		require.Nil(t, err)
		require.Len(t, tags, 5)
	})

	t.Run("204 (not found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		status, _, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodDelete, "/api/courses/test/tags/1234", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)
	})

	t.Run("204 (invalid tag for course)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		course1 := daos.NewTestBuilder(t).Db(db).Courses(1).Tags([]string{"Go"}).Build()
		course2 := daos.NewTestBuilder(t).Db(db).Courses(1).Tags([]string{"C"}).Build()

		courseTagDao := daos.NewCourseTagDao(db)

		// Delete the course2 tag from course1
		status, _, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodDelete, "/api/courses/"+course1[0].ID+"/tags/"+course2[0].Tags[0].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)

		// Asset the tag for course 1 has not been deleted
		tags, err := courseTagDao.List(&database.DatabaseParams{Where: squirrel.Eq{"course_id": course1[0].ID}}, nil)
		require.Nil(t, err)
		require.Len(t, tags, 1)
		require.Equal(t, "Go", tags[0].Tag)

		// Asset the tag for course 2 has not been deleted
		tags, err = courseTagDao.List(&database.DatabaseParams{Where: squirrel.Eq{"course_id": course2[0].ID}}, nil)
		require.Nil(t, err)
		require.Len(t, tags, 1)
		require.Equal(t, "C", tags[0].Tag)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		// Drop the table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewCourseTagDao(db).Table())
		require.Nil(t, err)
		_, err = db.Exec("DROP TABLE IF EXISTS " + daos.NewCourseDao(db).Table())
		require.Nil(t, err)

		status, _, err := coursesRequestHelper(appFs, db, cs, httptest.NewRequest(http.MethodDelete, "/api/courses/test/tags/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPERS
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func coursesRequestHelper(appFs *appFs.AppFs, db database.Database, cs *jobs.CourseScanner, req *http.Request) (int, []byte, error) {
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
