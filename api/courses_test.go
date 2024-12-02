package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetCourses(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		router, _ := setup(t)

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := unmarshalHelper[courseResponse](t, body)
		require.Zero(t, int(paginationResp.TotalItems))
		require.Zero(t, len(paginationResp.Items))
	})

	t.Run("200 (found)", func(t *testing.T) {
		router, ctx := setup(t)

		for i := range 5 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course %d", i)}
			require.NoError(t, router.dao.CreateCourse(ctx, course))
		}

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp := unmarshalHelper[courseResponse](t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, coursesResp, 5)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		router, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 5 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course %d", i)}
			require.NoError(t, router.dao.CreateCourse(ctx, course))
			courses = append(courses, course)
			time.Sleep(1 * time.Millisecond)
		}

		// CREATED_AT ASC
		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/?orderBy="+models.COURSE_TABLE+".created_at%20asc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp := unmarshalHelper[courseResponse](t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, coursesResp, 5)
		require.Equal(t, courses[0].ID, coursesResp[0].ID)

		// CREATED_AT DESC
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/?orderBy="+models.COURSE_TABLE+".created_at%20desc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp = unmarshalHelper[courseResponse](t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, coursesResp, 5)
		require.Equal(t, courses[4].ID, coursesResp[0].ID)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		router, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 17 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course %d", i)}
			require.NoError(t, router.dao.CreateCourse(ctx, course))
			courses = append(courses, course)
		}

		// Get the first page (10 courses)
		params := url.Values{
			"orderBy":                    {models.COURSE_TABLE + ".created_at asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp := unmarshalHelper[courseResponse](t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 10)
		require.Equal(t, courses[0].ID, coursesResp[0].ID)
		require.Equal(t, courses[9].ID, coursesResp[9].ID)

		// Get the second page (7 courses)
		params = url.Values{
			"orderBy":                    {models.COURSE_TABLE + ".created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp = unmarshalHelper[courseResponse](t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 7)
		require.Equal(t, courses[10].ID, coursesResp[0].ID)
		require.Equal(t, courses[16].ID, coursesResp[6].ID)
	})

	t.Run("200 (progress)", func(t *testing.T) {
		router, ctx := setup(t)

		// No courses
		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/?progress=started", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := unmarshalHelper[courseResponse](t, body)
		require.Zero(t, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 0)

		// Create 3 courses, each with 1 asset
		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course %d", i)}
			require.NoError(t, router.dao.CreateCourse(ctx, course))
			courses = append(courses, course)
		}

		assets := []*models.Asset{}
		for i, c := range courses {
			asset := &models.Asset{
				CourseID: c.ID,
				Title:    "asset 1",
				Prefix:   sql.NullInt16{Int16: 1, Valid: true},
				Chapter:  "Chapter 1",
				Type:     *types.NewAsset("mp4"),
				Path:     fmt.Sprintf("/course %d/chapter 1/01 asset 1.mp4", i+1),
				Hash:     security.RandomString(64),
			}

			require.NoError(t, router.dao.CreateAsset(ctx, asset))
			assets = append(assets, asset)

		}

		// Mark course 1 as started
		require.NoError(t, router.dao.CreateOrUpdateAssetProgress(ctx, &models.AssetProgress{AssetID: assets[0].ID, VideoPos: 10}))

		// Mark course 2 as completed
		require.NoError(t, router.dao.CreateOrUpdateAssetProgress(ctx, &models.AssetProgress{AssetID: assets[1].ID, VideoPos: 10, Completed: true}))

		// `progress` not defined
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ = unmarshalHelper[courseResponse](t, body)
		require.Equal(t, 3, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 3)

		// `progress` is started
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/?progress=started", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp := unmarshalHelper[courseResponse](t, body)
		require.Equal(t, 1, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 1)
		require.Equal(t, courses[0].ID, coursesResp[0].ID)

		// `progress` is completed
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/?progress=completed", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp = unmarshalHelper[courseResponse](t, body)
		require.Equal(t, 1, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 1)
		require.Equal(t, courses[1].ID, coursesResp[0].ID)

		// `progress` is not started
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/?progress=not%20started", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp = unmarshalHelper[courseResponse](t, body)
		require.Equal(t, 1, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 1)
		require.Equal(t, courses[2].ID, coursesResp[0].ID)
	})

	t.Run("200 (tags)", func(t *testing.T) {
		router, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 2 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course %d", i)}
			require.NoError(t, router.dao.CreateCourse(ctx, course))
			courses = append(courses, course)
			time.Sleep(1 * time.Millisecond)
		}

		// Add Go and C to course 1
		require.NoError(t, router.dao.CreateCourseTag(ctx, &models.CourseTag{CourseID: courses[0].ID, Tag: "Go"}))
		require.NoError(t, router.dao.CreateCourseTag(ctx, &models.CourseTag{CourseID: courses[0].ID, Tag: "C"}))

		// Add Go and Python to course 2
		require.NoError(t, router.dao.CreateCourseTag(ctx, &models.CourseTag{CourseID: courses[1].ID, Tag: "Go"}))
		require.NoError(t, router.dao.CreateCourseTag(ctx, &models.CourseTag{CourseID: courses[1].ID, Tag: "Python"}))

		// `tags` not defined
		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := unmarshalHelper[courseResponse](t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 2)

		// `tags` is Go
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/?tags=Go&orderBy=title%20asc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp := unmarshalHelper[courseResponse](t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 2)
		require.Equal(t, courses[0].ID, coursesResp[0].ID)
		require.Equal(t, courses[1].ID, coursesResp[1].ID)

		// `tags` is Go and C
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/?tags=Go,C", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp = unmarshalHelper[courseResponse](t, body)
		require.Equal(t, 1, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 1)
		require.Equal(t, courses[0].ID, coursesResp[0].ID)

		// `tags` is Go and Data Structures
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/?tags=Go,Data%20Structures", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ = unmarshalHelper[courseResponse](t, body)
		require.Equal(t, 0, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 0)
	})

	t.Run("200 (titles)", func(t *testing.T) {
		router, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i+1), Path: fmt.Sprintf("/course %d", i+1)}
			require.NoError(t, router.dao.CreateCourse(ctx, course))
			courses = append(courses, course)
			time.Sleep(1 * time.Millisecond)
		}

		// `titles` not defined
		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := unmarshalHelper[courseResponse](t, body)
		require.Equal(t, 3, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 3)

		// `titles` is course
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/?titles=course", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ = unmarshalHelper[courseResponse](t, body)
		require.Equal(t, 3, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 3)

		// `titles` is course 1 and course 2
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/?titles=course%201,course%202&orderBy=title%20asc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp := unmarshalHelper[courseResponse](t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 2)
		require.Equal(t, courses[0].ID, coursesResp[0].ID)
		require.Equal(t, courses[1].ID, coursesResp[1].ID)

		// `titles` is nothing
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/?titles=nothing", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ = unmarshalHelper[courseResponse](t, body)
		require.Equal(t, 0, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 0)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router, _ := setup(t)

		// Drop the courses table
		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.COURSE_TABLE)
		require.NoError(t, err)

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetCourse(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		router, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course %d", i)}
			require.NoError(t, router.dao.CreateCourse(ctx, course))
			courses = append(courses, course)
		}

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var courseResp courseResponse
		err = json.Unmarshal(body, &courseResp)
		require.NoError(t, err)
		require.Equal(t, courses[1].ID, courseResp.ID)
	})

	t.Run("404 (not found)", func(t *testing.T) {
		router, _ := setup(t)

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/invalid", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router, _ := setup(t)

		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.COURSE_TABLE)
		require.NoError(t, err)

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/invalid", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_CreateCourse(t *testing.T) {
	t.Run("201 (created)", func(t *testing.T) {
		router, _ := setup(t)

		router.config.AppFs.Fs.MkdirAll("/course 1", os.ModePerm)

		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": "course 1", "path": "/course 1" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		var courseResp courseResponse
		err = json.Unmarshal(body, &courseResp)
		require.NoError(t, err)
		require.NotNil(t, courseResp.ID)
		require.Equal(t, "course 1", courseResp.Title)
		require.Equal(t, "/course 1", courseResp.Path)
		require.True(t, courseResp.Available)
	})

	t.Run("400 (bind error)", func(t *testing.T) {
		router, _ := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Error parsing data")
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		router, _ := setup(t)

		// Missing title
		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "A title and path are required")

		// Missing path
		req = httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": "course 1", "path": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "A title and path are required")

		// Invalid path
		req = httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": "course 1", "path": "/test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Invalid course path")
	})

	t.Run("400 (existing course)", func(t *testing.T) {
		router, _ := setup(t)

		router.config.AppFs.Fs.MkdirAll("/course 1", os.ModePerm)

		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": "course 1", "path": "/course 1" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "A course with this path already exists")
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router, _ := setup(t)

		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.COURSE_TABLE)
		require.NoError(t, err)

		router.config.AppFs.Fs.MkdirAll("/course 1", os.ModePerm)

		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": "course 1", "path": "/course 1" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		require.Contains(t, string(body), "Error creating course")
	})

	t.Run("500 (scan error)", func(t *testing.T) {
		router, _ := setup(t)

		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.SCAN_TABLE)
		require.NoError(t, err)

		router.config.AppFs.Fs.MkdirAll("/course 1", os.ModePerm)

		req := httptest.NewRequest(http.MethodPost, "/api/courses/", strings.NewReader(`{"title": "course 1", "path": "/course 1" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		require.Contains(t, string(body), "Error creating scan job")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_DeleteCourse(t *testing.T) {
	t.Run("204 (deleted)", func(t *testing.T) {
		router, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course %d", i)}
			require.NoError(t, router.dao.CreateCourse(ctx, course))
			courses = append(courses, course)
		}

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodDelete, "/api/courses/"+courses[1].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)

		course := &models.Course{Base: models.Base{ID: courses[1].ID}}
		err = router.dao.GetById(ctx, course)
		require.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("204 (not found)", func(t *testing.T) {
		router, _ := setup(t)

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodDelete, "/api/courses/invalid", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router, _ := setup(t)

		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.COURSE_TABLE)
		require.NoError(t, err)

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodDelete, "/api/courses/invalid", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetCard(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{
			Title:    "course 1",
			Path:     "/course 1",
			CardPath: "/course 1/card.png",
		}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		router.config.AppFs.Fs.MkdirAll("/"+course.Path, os.ModePerm)
		require.Nil(t, afero.WriteFile(router.config.AppFs.Fs, course.CardPath, []byte("test"), os.ModePerm))

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/card", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, "test", string(body))
	})

	t.Run("404 (invalid id)", func(t *testing.T) {
		router, _ := setup(t)

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/invalid/card", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
		require.Contains(t, string(body), "Course not found")
	})

	t.Run("404 (no card)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{
			Title: "course 1",
			Path:  "/course 1",
		}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/card", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
		require.Contains(t, string(body), "Course has no card")
	})

	t.Run("404 (card not found)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{
			Title:    "course 1",
			Path:     "/course 1",
			CardPath: "/course 1/card.png",
		}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/card", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
		require.Contains(t, string(body), "Course card not found")
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router, _ := setup(t)

		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.COURSE_TABLE)
		require.NoError(t, err)

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/invalid/card", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetAssets(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		router, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 2 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course %d", i)}
			require.NoError(t, router.dao.CreateCourse(ctx, course))
			courses = append(courses, course)
		}

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := unmarshalHelper[assetResponse](t, body)
		require.Zero(t, int(paginationResp.TotalItems))
		require.Zero(t, len(paginationResp.Items))
	})

	t.Run("200 (found)", func(t *testing.T) {
		router, ctx := setup(t)

		courses := []*models.Course{}
		assets := []*models.Asset{}
		attachments := []*models.Attachment{}

		for i := range 2 {
			course := &models.Course{Title: fmt.Sprintf("Course %d", i+1), Path: fmt.Sprintf("/course/%d", i+1)}
			require.NoError(t, router.dao.CreateCourse(ctx, course))
			courses = append(courses, course)
		}

		for _, c := range courses {
			for j := range 2 {
				asset := &models.Asset{
					CourseID: c.ID,
					Title:    fmt.Sprintf("asset %d", j+1),
					Prefix:   sql.NullInt16{Int16: int16(j + 1), Valid: true},
					Chapter:  fmt.Sprintf("Chapter %d", j+1),
					Type:     *types.NewAsset("mp4"),
					Path:     fmt.Sprintf("/%s/asset %d", security.RandomString(4), j+1),
					Hash:     security.RandomString(64),
				}
				require.NoError(t, router.dao.CreateAsset(ctx, asset))
				assets = append(assets, asset)
				time.Sleep(1 * time.Millisecond)
			}
		}

		for _, asset := range assets {
			for j := range 2 {
				attachment := &models.Attachment{
					AssetID: asset.ID,
					Title:   fmt.Sprintf("attachment %d", j+1),
					Path:    fmt.Sprintf("/%s/attachment %d", security.RandomString(4), j+1),
				}
				require.NoError(t, router.dao.CreateAttachment(ctx, attachment))
				attachments = append(attachments, attachment)
			}
		}

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := unmarshalHelper[assetResponse](t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 2)
		require.Equal(t, assets[2].ID, assetsResp[0].ID)
		require.Equal(t, assets[3].ID, assetsResp[1].ID)
		require.Len(t, assetsResp[0].Attachments, 2)
		require.Equal(t, attachments[4].ID, assetsResp[0].Attachments[0].ID)
		require.Equal(t, attachments[5].ID, assetsResp[0].Attachments[1].ID)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		router, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 2 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course %d", i)}
			require.NoError(t, router.dao.CreateCourse(ctx, course))
			courses = append(courses, course)
		}

		assets := []*models.Asset{}
		for i, c := range courses {
			for j := range 2 {
				asset := &models.Asset{
					CourseID: c.ID,
					Title:    fmt.Sprintf("asset %d", j+1),
					Prefix:   sql.NullInt16{Int16: int16(j + 1), Valid: true},
					Chapter:  fmt.Sprintf("Chapter %d", j+1),
					Type:     *types.NewAsset("mp4"),
					Path:     fmt.Sprintf("/course %d/chapter %d/01 asset %d.mp4", i+1, j+1, j+1),
					Hash:     security.RandomString(64),
				}
				require.NoError(t, router.dao.CreateAsset(ctx, asset))
				assets = append(assets, asset)
				time.Sleep(1 * time.Millisecond)
			}
		}

		// CREATED_AT ASC
		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets/?orderBy=created_at%20asc", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := unmarshalHelper[assetResponse](t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 2)
		require.Equal(t, assets[2].ID, assetsResp[0].ID)
		require.Equal(t, assets[3].ID, assetsResp[1].ID)

		// CREATED_AT DESC
		req = httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/assets/?orderBy=created_at%20desc", nil)
		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp = unmarshalHelper[assetResponse](t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 2)
		require.Equal(t, assets[3].ID, assetsResp[0].ID)
		require.Equal(t, assets[2].ID, assetsResp[1].ID)

		// 		// ----------------------------
		// 		// CREATED_AT ASC + ATTACHMENTS.TITLE DESC
		// 		// ----------------------------
		// 		attDao := daos.NewAttachmentDao(router.config.DbManager.DataDb)

		// 		req = httptest.NewRequest(http.MethodGet, "/api/courses/"+testData[1].ID+"/assets/?expand=true&orderBy=created_at%20asc,"+attDao.Table()+".created_at%20desc", nil)
		// 		status, body, err = requestHelper(t, router, req)
		// 		require.NoError(t, err)
		// 		require.Equal(t, http.StatusOK, status)

		// 		paginationResp, assetsResp = unmarshalHelper[assetResponse](t, body)
		// 		require.Equal(t, 3, int(paginationResp.TotalItems))
		// 		require.Len(t, paginationResp.Items, 3)
		// 		require.Equal(t, testData[1].Assets[0].ID, assetsResp[0].ID)
		// 		require.Len(t, assetsResp[0].Attachments, 2)
		// 		require.Equal(t, testData[1].Assets[0].Attachments[1].ID, assetsResp[0].Attachments[0].ID)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "course 1", Path: "/course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		assets := []*models.Asset{}
		for i := range 17 {
			asset := &models.Asset{
				CourseID: course.ID,
				Title:    fmt.Sprintf("asset %d", i+1),
				Prefix:   sql.NullInt16{Int16: int16(i + 1), Valid: true},
				Chapter:  fmt.Sprintf("Chapter %d", i+1),
				Type:     *types.NewAsset("mp4"),
				Path:     fmt.Sprintf("/%s/asset %d", security.RandomString(4), i+1),
				Hash:     security.RandomString(64),
			}
			require.NoError(t, router.dao.CreateAsset(ctx, asset))
			assets = append(assets, asset)
			time.Sleep(1 * time.Millisecond)
		}

		// Get the first page (10 assets)
		params := url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/?"+params.Encode(), nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp := unmarshalHelper[assetResponse](t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 10)
		require.Equal(t, assets[0].ID, assetsResp[0].ID)
		require.Equal(t, assets[9].ID, assetsResp[9].ID)

		// Get the second page (7 assets)
		params = url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}

		req = httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/?"+params.Encode(), nil)
		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, assetsResp = unmarshalHelper[assetResponse](t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 7)
		require.Equal(t, assets[10].ID, assetsResp[0].ID)
		require.Equal(t, assets[16].ID, assetsResp[6].ID)
	})

	t.Run("500 (asset internal error)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "course 1", Path: "/course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.ASSET_TABLE)
		require.NoError(t, err)

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetAsset(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "course 1", Path: "/course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		assets := []*models.Asset{}
		attachments := []*models.Attachment{}

		for j := range 2 {
			asset := &models.Asset{
				CourseID: course.ID,
				Title:    fmt.Sprintf("asset %d", j+1),
				Prefix:   sql.NullInt16{Int16: int16(j + 1), Valid: true},
				Chapter:  fmt.Sprintf("Chapter %d", j+1),
				Type:     *types.NewAsset("mp4"),
				Path:     fmt.Sprintf("/%s/asset %d", security.RandomString(4), j+1),
				Hash:     security.RandomString(64),
			}
			require.NoError(t, router.dao.CreateAsset(ctx, asset))
			assets = append(assets, asset)
			time.Sleep(1 * time.Millisecond)
		}

		for _, asset := range assets {
			for j := range 2 {
				attachment := &models.Attachment{
					AssetID: asset.ID,
					Title:   fmt.Sprintf("attachment %d", j+1),
					Path:    fmt.Sprintf("/%s/attachment %d", security.RandomString(4), j+1),
				}
				require.NoError(t, router.dao.CreateAttachment(ctx, attachment))
				attachments = append(attachments, attachment)
			}
		}

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+assets[1].ID, nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var assetResp assetResponse
		err = json.Unmarshal(body, &assetResp)
		require.NoError(t, err)
		require.Equal(t, assets[1].ID, assetResp.ID)
		require.Equal(t, assets[1].Title, assetResp.Title)
		require.Equal(t, assets[1].Path, assetResp.Path)
		require.Len(t, assetResp.Attachments, 2)
		require.Equal(t, attachments[2].ID, assetResp.Attachments[0].ID)
		require.Equal(t, attachments[3].ID, assetResp.Attachments[1].ID)
	})

	t.Run("400 (invalid asset for course)", func(t *testing.T) {
		router, ctx := setup(t)

		course1 := &models.Course{Title: "course 1", Path: "/course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course1))

		course2 := &models.Course{Title: "course 2", Path: "/course 2"}
		require.NoError(t, router.dao.CreateCourse(ctx, course2))

		asset := &models.Asset{
			CourseID: course2.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course 1/Chapter 1/01 Asset 1.mp4",
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		// Request an asset that does not belong to the course
		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course1.ID+"/assets/"+asset.ID, nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Asset does not belong to course")
	})

	t.Run("404 (asset not found)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "course 1", Path: "/course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/invalid", nil)
		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (asset internal error)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "course 1", Path: "/course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.ASSET_TABLE)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/invalid", nil)
		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_ServeAsset(t *testing.T) {
	t.Run("200 (full video)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		require.Nil(t, router.config.AppFs.Fs.MkdirAll(filepath.Dir(asset.Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(router.config.AppFs.Fs, asset.Path, []byte("video"), os.ModePerm))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/serve", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, "video", string(body))
	})

	t.Run("200 (stream video)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		require.Nil(t, router.config.AppFs.Fs.MkdirAll(filepath.Dir(asset.Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(router.config.AppFs.Fs, asset.Path, []byte("video"), os.ModePerm))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/serve", nil)
		req.Header.Add("Range", "bytes=0-")

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusPartialContent, status)
		require.Equal(t, "video", string(body))
	})

	t.Run("200 (html)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("html"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		require.Nil(t, router.config.AppFs.Fs.MkdirAll(filepath.Dir(asset.Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(router.config.AppFs.Fs, asset.Path, []byte("html data"), os.ModePerm))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/serve", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, "html data", string(body))
	})

	t.Run("400 (invalid asset for course)", func(t *testing.T) {
		router, ctx := setup(t)

		course1 := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course1))

		course2 := &models.Course{Title: "Course 2", Path: "/course/2"}
		require.NoError(t, router.dao.CreateCourse(ctx, course2))

		asset := &models.Asset{
			CourseID: course2.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course1.ID+"/assets/"+asset.ID+"/serve", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Asset does not belong to course")
	})

	t.Run("400 (invalid path)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/serve", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Asset does not exist")
	})

	t.Run("400 (invalid video range)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		require.Nil(t, router.config.AppFs.Fs.MkdirAll(filepath.Dir(asset.Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(router.config.AppFs.Fs, asset.Path, []byte("video"), os.ModePerm))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/serve", nil)
		req.Header.Add("Range", "bytes=10-1")

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Range start cannot be greater than end")
	})

	t.Run("404 (not found)", func(t *testing.T) {
		router, _ := setup(t)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/invalid/assets/invalid/serve", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
		require.Contains(t, string(body), "Asset not found")
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router, _ := setup(t)

		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.ASSET_TABLE)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/invalid/assets/invalid/serve", nil)
		req.Header.Set("Content-Type", "application/json")

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		require.Contains(t, string(body), "Error looking up asset")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_updateAssetProgress(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		// Update video position
		assetProgress := &assetProgressRequest{
			VideoPos: 45,
		}

		data, err := json.Marshal(assetProgress)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/progress", strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)

		assetResult := &models.Asset{Base: models.Base{ID: asset.ID}}
		require.NoError(t, router.dao.GetById(ctx, assetResult))
		require.Equal(t, 45, assetResult.Progress.VideoPos)
		require.False(t, assetResult.Progress.Completed)
		require.True(t, assetResult.Progress.CompletedAt.IsZero())

		// Set completed to true
		assetProgress.Completed = true

		data, err = json.Marshal(assetProgress)
		require.NoError(t, err)

		req = httptest.NewRequest(http.MethodPut, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/progress", strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		status, _, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)

		require.NoError(t, router.dao.GetById(ctx, assetResult))
		require.Equal(t, 45, assetResult.Progress.VideoPos)
		require.True(t, assetResult.Progress.Completed)
		require.False(t, assetResult.Progress.CompletedAt.IsZero())

		// Set video position to 10 and completed to false
		assetProgress.VideoPos = 10
		assetProgress.Completed = false

		data, err = json.Marshal(assetProgress)
		require.NoError(t, err)

		req = httptest.NewRequest(http.MethodPut, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/progress", strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		status, _, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)

		require.NoError(t, router.dao.GetById(ctx, assetResult))
		require.Equal(t, 10, assetResult.Progress.VideoPos)
		require.False(t, assetResult.Progress.Completed)
		require.True(t, assetResult.Progress.CompletedAt.IsZero())
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		router, _ := setup(t)

		req := httptest.NewRequest(http.MethodPut, "/api/courses/invalid/assets/invalid/progress", strings.NewReader(`bob`))
		req.Header.Set("Content-Type", "application/json")

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Error parsing data")
	})

	t.Run("400 (asset not found)", func(t *testing.T) {
		router, _ := setup(t)

		req := httptest.NewRequest(http.MethodPut, "/api/courses/invalid/assets/invalid/progress", strings.NewReader(`{"videoPos": 10}`))
		req.Header.Set("Content-Type", "application/json")

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
		require.Contains(t, string(body), "Asset not found")
	})

	t.Run("400 (invalid asset for course)", func(t *testing.T) {
		router, ctx := setup(t)

		course1 := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course1))

		course2 := &models.Course{Title: "Course 2", Path: "/course/2"}
		require.NoError(t, router.dao.CreateCourse(ctx, course2))

		asset := &models.Asset{
			CourseID: course2.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		req := httptest.NewRequest(http.MethodPut, "/api/courses/"+course1.ID+"/assets/"+asset.ID+"/progress", strings.NewReader(`{"videoPos": 10}`))
		req.Header.Set("Content-Type", "application/json")

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Asset does not belong to course")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetAttachments(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/attachments", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := unmarshalHelper[attachmentResponse](t, body)
		require.Zero(t, int(paginationResp.TotalItems))
		require.Zero(t, len(paginationResp.Items))
	})

	t.Run("200 (found)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		attachments := []*models.Attachment{}
		for i := range 2 {
			attachment := &models.Attachment{
				AssetID: asset.ID,
				Title:   fmt.Sprintf("attachment %d", i+1),
				Path:    fmt.Sprintf("/%s/attachment %d", security.RandomString(4), i+1),
			}
			require.NoError(t, router.dao.CreateAttachment(ctx, attachment))
			attachments = append(attachments, attachment)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/attachments", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentResp := unmarshalHelper[attachmentResponse](t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 2)
		require.Equal(t, attachments[0].ID, attachmentResp[0].ID)
		require.Equal(t, attachments[1].ID, attachmentResp[1].ID)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		attachments := []*models.Attachment{}
		for i := range 2 {
			attachment := &models.Attachment{
				AssetID: asset.ID,
				Title:   fmt.Sprintf("attachment %d", i+1),
				Path:    fmt.Sprintf("/%s/attachment %d", security.RandomString(4), i+1),
			}
			require.NoError(t, router.dao.CreateAttachment(ctx, attachment))
			attachments = append(attachments, attachment)
			time.Sleep(1 * time.Millisecond)
		}

		// CREATED_AT ASC
		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/attachments?orderBy=created_at%20asc", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentResp := unmarshalHelper[attachmentResponse](t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 2)
		require.Equal(t, attachments[0].ID, attachmentResp[0].ID)
		require.Equal(t, attachments[1].ID, attachmentResp[1].ID)

		// CREATED_AT DESC
		req = httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/attachments?orderBy=created_at%20desc", nil)
		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentResp = unmarshalHelper[attachmentResponse](t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 2)

		require.Equal(t, attachments[1].ID, attachmentResp[0].ID)
		require.Equal(t, attachments[0].ID, attachmentResp[1].ID)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		attachments := []*models.Attachment{}
		for i := range 17 {
			attachment := &models.Attachment{
				AssetID: asset.ID,
				Title:   fmt.Sprintf("attachment %d", i+1),
				Path:    fmt.Sprintf("/%s/attachment %d", security.RandomString(4), i+1),
			}
			require.NoError(t, router.dao.CreateAttachment(ctx, attachment))
			attachments = append(attachments, attachment)
			time.Sleep(1 * time.Millisecond)
		}

		// Get the first page (10 attachments)
		params := url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/attachments?"+params.Encode(), nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentResp := unmarshalHelper[attachmentResponse](t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 10)
		require.Equal(t, attachments[0].ID, attachmentResp[0].ID)
		require.Equal(t, attachments[9].ID, attachmentResp[9].ID)

		// Get the second page (7 attachments)
		params = url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}

		req = httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/attachments?"+params.Encode(), nil)
		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, attachmentResp = unmarshalHelper[attachmentResponse](t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 7)
		require.Equal(t, attachments[10].ID, attachmentResp[0].ID)
		require.Equal(t, attachments[16].ID, attachmentResp[6].ID)
	})

	t.Run("404 (asset not found)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/invalid/attachments", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
		require.Contains(t, string(body), "Asset not found")
	})

	t.Run("400 (invalid asset for course)", func(t *testing.T) {
		router, ctx := setup(t)

		course1 := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course1))

		course2 := &models.Course{Title: "Course 2", Path: "/course/2"}
		require.NoError(t, router.dao.CreateCourse(ctx, course2))

		asset := &models.Asset{
			CourseID: course2.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course1.ID+"/assets/"+asset.ID+"/attachments", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Asset does not belong to course")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetAttachment(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		attachment := &models.Attachment{
			AssetID: asset.ID,
			Title:   "attachment 1",
			Path:    fmt.Sprintf("/%s/attachment 1", security.RandomString(4)),
		}
		require.NoError(t, router.dao.CreateAttachment(ctx, attachment))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/attachments/"+attachment.ID, nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var respData attachmentResponse
		err = json.Unmarshal(body, &respData)
		require.NoError(t, err)
		require.Equal(t, attachment.ID, respData.ID)
	})

	t.Run("400 (invalid asset for course)", func(t *testing.T) {
		router, ctx := setup(t)

		course1 := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course1))

		course2 := &models.Course{Title: "Course 2", Path: "/course/2"}
		require.NoError(t, router.dao.CreateCourse(ctx, course2))

		asset := &models.Asset{
			CourseID: course2.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course1.ID+"/assets/"+asset.ID+"/attachments/invalid", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Asset does not belong to course")
	})

	t.Run("400 (invalid attachment for asset)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		asset1 := &models.Asset{
			CourseID: course.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset1))

		asset2 := &models.Asset{
			CourseID: course.ID,
			Title:    "asset 2",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 2",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 2", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset2))

		attachment := &models.Attachment{
			AssetID: asset1.ID,
			Title:   "attachment 1",
			Path:    fmt.Sprintf("/%s/attachment 1", security.RandomString(4)),
		}
		require.NoError(t, router.dao.CreateAttachment(ctx, attachment))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset2.ID+"/attachments/"+attachment.ID, nil)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Attachment does not belong to asset")
	})

	t.Run("404 (course not found)", func(t *testing.T) {
		router, _ := setup(t)

		req := httptest.NewRequest(http.MethodGet, "/api/courses/invalid/assets/invalid/attachments/invalid", nil)
		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("404 (asset not found)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/invalid/attachments/invalid", nil)
		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("404 (attachment not found)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/attachments/invalid", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
		require.Contains(t, string(body), "Attachment not found")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_ServeAttachment(t *testing.T) {
	t.Run("200 (ok)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		attachment := &models.Attachment{
			AssetID: asset.ID,
			Title:   "attachment 1",
			Path:    fmt.Sprintf("/%s/attachment 1", security.RandomString(4)),
		}
		require.NoError(t, router.dao.CreateAttachment(ctx, attachment))

		require.Nil(t, router.config.AppFs.Fs.MkdirAll(filepath.Dir(attachment.Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(router.config.AppFs.Fs, attachment.Path, []byte("hello"), os.ModePerm))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/attachments/"+attachment.ID+"/serve", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, "hello", string(body))
	})

	t.Run("400 (invalid path)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		attachment := &models.Attachment{
			AssetID: asset.ID,
			Title:   "attachment 1",
			Path:    fmt.Sprintf("/%s/attachment 1", security.RandomString(4)),
		}
		require.NoError(t, router.dao.CreateAttachment(ctx, attachment))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/attachments/"+attachment.ID+"/serve", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Attachment does not exist")
	})

	t.Run("400 (invalid asset for course)", func(t *testing.T) {
		router, ctx := setup(t)

		course1 := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course1))

		course2 := &models.Course{Title: "Course 2", Path: "/course/2"}
		require.NoError(t, router.dao.CreateCourse(ctx, course2))

		asset := &models.Asset{
			CourseID: course2.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course1.ID+"/assets/"+asset.ID+"/attachments/invalid/serve", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Asset does not belong to course")
	})

	t.Run("400 (invalid attachment for asset)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		assets := []*models.Asset{}
		for j := range 2 {
			asset := &models.Asset{
				CourseID: course.ID,
				Title:    fmt.Sprintf("asset %d", j+1),
				Prefix:   sql.NullInt16{Int16: int16(j + 1), Valid: true},
				Chapter:  fmt.Sprintf("Chapter %d", j+1),
				Type:     *types.NewAsset("mp4"),
				Path:     fmt.Sprintf("/%s/asset %d", security.RandomString(4), j+1),
				Hash:     security.RandomString(64),
			}
			require.NoError(t, router.dao.CreateAsset(ctx, asset))
			assets = append(assets, asset)
		}

		attachment := &models.Attachment{
			AssetID: assets[0].ID,
			Title:   "attachment 1",
			Path:    fmt.Sprintf("/%s/attachment 1", security.RandomString(4)),
		}
		require.NoError(t, router.dao.CreateAttachment(ctx, attachment))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+assets[1].ID+"/attachments/"+attachment.ID+"/serve", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Attachment does not belong to asset")
	})

	t.Run("404 (asset not found)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/invalid/attachments/invalid/serve", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
		require.Contains(t, string(body), "Asset not found")
	})

	t.Run("404 (attachment not found)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     fmt.Sprintf("/%s/asset 1", security.RandomString(4)),
			Hash:     security.RandomString(64),
		}
		require.NoError(t, router.dao.CreateAsset(ctx, asset))

		req := httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/assets/"+asset.ID+"/attachments/invalid/serve", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
		require.Contains(t, string(body), "Attachment not found")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_GetTags(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		router, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course %d", i)}
			require.NoError(t, router.dao.CreateCourse(ctx, course))
			courses = append(courses, course)
		}

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/tags", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var tags []courseTagResponse
		err = json.Unmarshal(body, &tags)
		require.NoError(t, err)
		require.Zero(t, len(tags))
	})

	t.Run("200 (found)", func(t *testing.T) {
		router, ctx := setup(t)

		tagOptions := []string{"Go", "C", "JavaScript", "TypeScript", "Java", "Python"}

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course %d", i)}
			require.NoError(t, router.dao.CreateCourse(ctx, course))
			courses = append(courses, course)

			for _, tag := range tagOptions {
				tag := &models.CourseTag{CourseID: course.ID, Tag: tag}
				require.NoError(t, router.dao.CreateCourseTag(ctx, tag))
			}
		}

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/"+courses[1].ID+"/tags", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var tags []courseTagResponse
		err = json.Unmarshal(body, &tags)
		require.NoError(t, err)
		require.Len(t, tags, 6)
		require.Equal(t, "C", tags[0].Tag)
		require.Equal(t, "TypeScript", tags[5].Tag)
	})

	t.Run("404 (course not found)", func(t *testing.T) {
		router, _ := setup(t)

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/invalid/tags", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (course internal error)", func(t *testing.T) {
		router, _ := setup(t)

		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.COURSE_TABLE)
		require.NoError(t, err)

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/invalid/tags", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})

	t.Run("500 (courses_tags internal error)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "course 1", Path: "/course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.COURSE_TAG_TABLE)
		require.NoError(t, err)

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/courses/"+course.ID+"/tags/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_CreateTag(t *testing.T) {
	t.Run("201 (created)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "course 1", Path: "/course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		req := httptest.NewRequest(http.MethodPost, "/api/courses/"+course.ID+"/tags", strings.NewReader(`{"tag": "Go" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		var courseTagResp courseTagResponse
		err = json.Unmarshal(body, &courseTagResp)
		require.NoError(t, err)
		require.NotNil(t, courseTagResp.ID)
		require.Equal(t, "Go", courseTagResp.Tag)
	})

	t.Run("400 (bind error)", func(t *testing.T) {
		router, _ := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/courses/invalid/tags", strings.NewReader(`{`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Error parsing data")
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "course 1", Path: "/course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		req := httptest.NewRequest(http.MethodPost, "/api/courses/"+course.ID+"/tags", strings.NewReader(`{"tag": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "A tag is required")
	})

	t.Run("400 (existing tag)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "course 1", Path: "/course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		req := httptest.NewRequest(http.MethodPost, "/api/courses/"+course.ID+"/tags", strings.NewReader(`{"tag": "Go"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		// Create the tag again
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "A tag for this course already exists")
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router, ctx := setup(t)

		course := &models.Course{Title: "course 1", Path: "/course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course))

		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.COURSE_TAG_TABLE)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/courses/"+course.ID+"/tags", strings.NewReader(`{"tag": "Go"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		require.Contains(t, string(body), "Error creating course tag")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourses_DeleteTag(t *testing.T) {
	t.Run("204 (deleted)", func(t *testing.T) {
		router, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course %d", i)}
			require.NoError(t, router.dao.CreateCourse(ctx, course))
			courses = append(courses, course)

			for j := range 3 {
				tag := &models.CourseTag{CourseID: course.ID, Tag: fmt.Sprintf("Tag %d", j)}
				require.NoError(t, router.dao.CreateCourseTag(ctx, tag))
			}
		}

		tags := []*models.CourseTag{}
		require.NoError(t, router.dao.List(ctx, &tags, &database.Options{Where: squirrel.Eq{"course_id": courses[1].ID}}))
		require.Len(t, tags, 3)

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodDelete, "/api/courses/"+courses[1].ID+"/tags/"+tags[1].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)

		require.NoError(t, router.dao.List(ctx, &tags, &database.Options{Where: squirrel.Eq{"course_id": courses[1].ID}}))
		require.Len(t, tags, 2)
	})

	t.Run("204 (not found)", func(t *testing.T) {
		router, _ := setup(t)

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodDelete, "/api/courses/invalid/tags/invalid", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)
	})

	t.Run("204 (invalid tag for course)", func(t *testing.T) {
		router, ctx := setup(t)

		course1 := &models.Course{Title: "course 1", Path: "/course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course1))

		tag1 := &models.CourseTag{CourseID: course1.ID, Tag: "Go"}
		require.NoError(t, router.dao.CreateCourseTag(ctx, tag1))

		course2 := &models.Course{Title: "course 2", Path: "/course 2"}
		require.NoError(t, router.dao.CreateCourse(ctx, course2))

		tag2 := &models.CourseTag{CourseID: course2.ID, Tag: "C"}
		require.NoError(t, router.dao.CreateCourseTag(ctx, tag2))

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodDelete, "/api/courses/"+course1.ID+"/tags/"+tag2.ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)

		require.NoError(t, router.dao.GetById(ctx, tag1))
		require.NoError(t, router.dao.GetById(ctx, tag2))
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router, _ := setup(t)

		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.COURSE_TAG_TABLE)
		require.NoError(t, err)

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodDelete, "/api/courses/invalid/tags/invalid", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}
