package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Slice of 20 tags for testing (programming languages)
var test_tags = []string{
	"JavaScript", "Python", "Java", "Ruby", "PHP",
	"TypeScript", "C#", "C++", "C", "Swift",
	"Kotlin", "Rust", "Go", "Perl", "Scala",
	"R", "Objective-C", "Shell", "PowerShell", "Haskell",
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestTags_GetTags(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		status, body, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := tagsUnmarshalHelper(t, body)
		assert.Zero(t, int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 0)
	})

	t.Run("200 (found)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		for _, tag := range test_tags {
			require.Nil(t, daos.NewTagDao(db).Create(&models.Tag{Tag: tag}, nil))
		}

		status, body, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp := tagsUnmarshalHelper(t, body)
		assert.Equal(t, len(test_tags), int(paginationResp.TotalItems))
		require.Len(t, tagsResp, len(test_tags))
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		for _, tag := range test_tags {
			require.Nil(t, daos.NewTagDao(db).Create(&models.Tag{Tag: tag}, nil))
			time.Sleep(time.Millisecond * 1)
		}

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		status, body, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?orderBy=created_at%20asc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp := tagsUnmarshalHelper(t, body)
		require.Equal(t, len(test_tags), int(paginationResp.TotalItems))
		require.Len(t, tagsResp, len(test_tags))
		assert.Equal(t, test_tags[0], tagsResp[0].Tag)

		// ----------------------------
		// CREATED_AT DESC
		// ----------------------------
		status, body, err = tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?orderBy=created_at%20desc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp = tagsUnmarshalHelper(t, body)
		require.Equal(t, len(test_tags), int(paginationResp.TotalItems))
		require.Len(t, tagsResp, len(test_tags))
		assert.Equal(t, test_tags[len(test_tags)-1], tagsResp[0].Tag)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		for _, tag := range test_tags {
			require.Nil(t, daos.NewTagDao(db).Create(&models.Tag{Tag: tag}, nil))
		}

		// ----------------------------
		// Get the first page (11 tags)
		// ----------------------------
		params := url.Values{
			"orderBy":                    {"tag asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"11"},
		}

		status, body, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp := tagsUnmarshalHelper(t, body)
		assert.Equal(t, len(test_tags), int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 11)

		// Check the last tag in the paginated response
		assert.Equal(t, "Perl", tagsResp[len(paginationResp.Items)-1].Tag)

		// ----------------------------
		// Get the second page (9 tags)
		// ----------------------------
		params = url.Values{
			"orderBy":                    {"tag asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"11"},
		}
		status, body, err = tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp = tagsUnmarshalHelper(t, body)
		assert.Equal(t, len(test_tags), int(paginationResp.TotalItems))
		assert.Len(t, paginationResp.Items, 9)

		// Check the last tag in the paginated response
		assert.Equal(t, "TypeScript", tagsResp[len(paginationResp.Items)-1].Tag)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		// Drop the courses table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewTagDao(db).Table)
		require.Nil(t, err)

		status, _, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestTags_GetCourse(t *testing.T) {
// 	t.Run("200 (found)", func(t *testing.T) {
// 		_, db, _, _ := setup(t)

// 		testData := daos.NewTestBuilder(t).Db(db).Courses(3).Build()

// 		status, body, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/"+testData[2].ID, nil))
// 		require.NoError(t, err)
// 		require.Equal(t, http.StatusOK, status)

// 		var courseResp courseResponse
// 		err = json.Unmarshal(body, &courseResp)
// 		require.Nil(t, err)
// 		assert.Equal(t, testData[2].ID, courseResp.ID)
// 	})

// 	t.Run("200 (availability)", func(t *testing.T) {
// 		_, db, _, _ := setup(t)

// 		testData := daos.NewTestBuilder(t).Db(db).Courses(3).Build()

// 		// ----------------------------
// 		// Unavailable
// 		// ----------------------------
// 		status, body, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/"+testData[2].ID, nil))
// 		require.NoError(t, err)
// 		require.Equal(t, http.StatusOK, status)

// 		var courseResp courseResponse
// 		err = json.Unmarshal(body, &courseResp)
// 		require.Nil(t, err)
// 		assert.Equal(t, testData[2].ID, courseResp.ID)
// 		assert.False(t, courseResp.Available)

// 		// ----------------------------
// 		// Available
// 		// ----------------------------
// 		courseDao := daos.NewCourseDao(db)
// 		testData[2].Available = true
// 		require.Nil(t, courseDao.Update(testData[2].Course))

// 		status, body, err = tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/"+testData[2].ID, nil))
// 		require.NoError(t, err)
// 		require.Equal(t, http.StatusOK, status)

// 		err = json.Unmarshal(body, &courseResp)
// 		require.Nil(t, err)
// 		assert.Equal(t, testData[2].ID, courseResp.ID)
// 		assert.True(t, courseResp.Available)
// 	})

// 	t.Run("404 (not found)", func(t *testing.T) {
// 		_, db, _, _ := setup(t)

// 		status, _, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/test", nil))
// 		require.NoError(t, err)
// 		require.Equal(t, http.StatusNotFound, status)
// 	})

// 	t.Run("500 (internal error)", func(t *testing.T) {
// 		_, db, _, _ := setup(t)

// 		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewTagDao(db).Table)
// 		require.Nil(t, err)

// 		status, _, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/test", nil))
// 		require.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, status)
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestTags_Create(t *testing.T) {
// 	t.Run("201 (created)", func(t *testing.T) {
// 		_, db, _, _ := setup(t)

// 		testData := daos.NewTestBuilder(t).Courses(1).Build()
// 		appFs.Fs.MkdirAll(testData[0].Path, os.ModePerm)

// 		postData := fmt.Sprintf(`{"title": "%s", "path": "%s" }`, testData[0].Title, testData[0].Path)
// 		req := httptest.NewRequest(http.MethodPost, "/api/tags/", strings.NewReader(postData))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		status, body, err := tagsRequestHelper(db, req)
// 		require.NoError(t, err)
// 		require.Equal(t, http.StatusCreated, status)

// 		var courseResp courseResponse
// 		err = json.Unmarshal(body, &courseResp)
// 		require.Nil(t, err)
// 		assert.NotNil(t, courseResp.ID)
// 		assert.Equal(t, testData[0].Title, courseResp.Title)
// 		assert.Equal(t, testData[0].Path, courseResp.Path)
// 		assert.True(t, courseResp.Available)
// 	})

// 	t.Run("400 (bind error)", func(t *testing.T) {
// 		_, db, _, _ := setup(t)

// 		req := httptest.NewRequest(http.MethodPost, "/api/tags/", strings.NewReader(`{`))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		status, body, err := tagsRequestHelper(db, req)
// 		require.NoError(t, err)
// 		require.Equal(t, http.StatusBadRequest, status)
// 		assert.Contains(t, string(body), "error parsing data")
// 	})

// 	t.Run("400 (invalid data)", func(t *testing.T) {
// 		_, db, _, _ := setup(t)

// 		// ----------------------------
// 		// Missing title
// 		// ----------------------------
// 		req := httptest.NewRequest(http.MethodPost, "/api/tags/", strings.NewReader(`{"title": ""}`))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		status, body, err := tagsRequestHelper(db, req)
// 		require.NoError(t, err)
// 		require.Equal(t, http.StatusBadRequest, status)
// 		assert.Contains(t, string(body), "a title and path are required")

// 		// ----------------------------
// 		// Missing path
// 		// ----------------------------
// 		req = httptest.NewRequest(http.MethodPost, "/api/tags/", strings.NewReader(`{"title": "course 1", "path": ""}`))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		status, body, err = tagsRequestHelper(db, req)
// 		require.NoError(t, err)
// 		require.Equal(t, http.StatusBadRequest, status)
// 		assert.Contains(t, string(body), "a title and path are required")

// 		// ----------------------------
// 		// Invalid path
// 		// ----------------------------
// 		req = httptest.NewRequest(http.MethodPost, "/api/tags/", strings.NewReader(`{"title": "course 1", "path": "/test"}`))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		status, body, err = tagsRequestHelper(db, req)
// 		require.NoError(t, err)
// 		require.Equal(t, http.StatusBadRequest, status)
// 		assert.Contains(t, string(body), "invalid course path")
// 	})

// 	t.Run("400 (existing course)", func(t *testing.T) {
// 		_, db, _, _ := setup(t)

// 		coursePath := "/course 1/"
// 		appFs.Fs.MkdirAll(coursePath, os.ModePerm)

// 		postData := fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
// 		req := httptest.NewRequest(http.MethodPost, "/api/tags/", strings.NewReader(postData))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		status, _, err := tagsRequestHelper(db, req)
// 		require.NoError(t, err)
// 		require.Equal(t, http.StatusCreated, status)

// 		// Create the course again
// 		status, body, err := tagsRequestHelper(db, req)
// 		require.NoError(t, err)
// 		require.Equal(t, http.StatusBadRequest, status)
// 		assert.Contains(t, string(body), "a course with this path already exists ")
// 	})

// 	t.Run("500 (internal error)", func(t *testing.T) {
// 		_, db, _, _ := setup(t)

// 		// Drop the courses table
// 		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewTagDao(db).Table)
// 		require.Nil(t, err)

// 		coursePath := "/course 1/"
// 		appFs.Fs.MkdirAll(coursePath, os.ModePerm)

// 		postData := fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
// 		req := httptest.NewRequest(http.MethodPost, "/api/tags/", strings.NewReader(postData))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		status, body, err := tagsRequestHelper(db, req)
// 		require.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, status)
// 		assert.Contains(t, string(body), "error creating course")
// 	})

// 	t.Run("500 (scan error)", func(t *testing.T) {
// 		_, db, _, _ := setup(t)

// 		// Drop scan table
// 		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.TableScans())
// 		require.Nil(t, err)

// 		coursePath := "/course 1/"
// 		appFs.Fs.MkdirAll(coursePath, os.ModePerm)

// 		postData := fmt.Sprintf(`{"title": "course 1", "path": "%s" }`, coursePath)
// 		req := httptest.NewRequest(http.MethodPost, "/api/tags/", strings.NewReader(postData))
// 		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

// 		status, body, err := tagsRequestHelper(db, req)
// 		require.NoError(t, err)
// 		require.Equal(t, http.StatusInternalServerError, status)
// 		assert.Contains(t, string(body), "error creating scan job")
// 	})
// }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestTags_DeleteTag(t *testing.T) {
	t.Run("204 (deleted)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		tags := []*models.Tag{}
		for _, taggy := range test_tags {
			tag := &models.Tag{Tag: taggy}
			require.Nil(t, daos.NewTagDao(db).Create(tag, nil))
			tags = append(tags, tag)
		}

		tagsDao := daos.NewTagDao(db)

		// ----------------------------
		// Delete tag 3
		// ----------------------------
		status, _, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodDelete, "/api/tags/"+tags[2].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)

		result, err := tagsDao.List(&database.DatabaseParams{Where: squirrel.Eq{"id": tags[2].ID}}, nil)
		require.Nil(t, err)
		assert.Zero(t, len(result))

		// // ----------------------------
		// // Cascades
		// // ----------------------------

		// // Scan
		// _, err = scanDao.Get(testData[2].ID)
		// assert.ErrorIs(t, err, sql.ErrNoRows)

		// // Assets
		// count, err := assetsDao.Count(&database.DatabaseParams{Where: squirrel.Eq{daos.TableAssets() + ".course_id": testData[2].ID}})
		// require.Nil(t, err)
		// assert.Zero(t, count)

		// // Attachments
		// count, err = attachmentsDao.Count(&database.DatabaseParams{Where: squirrel.Eq{daos.TableAttachments() + ".course_id": testData[2].ID}})
		// require.Nil(t, err)
		// assert.Zero(t, count)
	})

	t.Run("204 (not found)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		status, _, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodDelete, "/api/tags/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		// Drop the table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewTagDao(db).Table)
		require.Nil(t, err)

		status, _, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodDelete, "/api/tags/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPERS
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func tagsRequestHelper(db database.Database, req *http.Request) (int, []byte, error) {
	f := fiber.New()
	bindTagsApi(f.Group("/api"), db)

	resp, err := f.Test(req)
	if err != nil {
		return -1, nil, err
	}

	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, b, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func tagsUnmarshalHelper(t *testing.T, body []byte) (pagination.PaginationResult, []tagResponse) {
	var respData pagination.PaginationResult
	err := json.Unmarshal(body, &respData)
	require.Nil(t, err)

	var tagsResponse []tagResponse
	for _, item := range respData.Items {
		var tag tagResponse
		require.Nil(t, json.Unmarshal(item, &tag))
		tagsResponse = append(tagsResponse, tag)
	}

	return respData, tagsResponse
}
