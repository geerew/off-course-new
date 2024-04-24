package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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

		daos.NewTestBuilder(t).Db(db).Courses(2).Tags([]string{"PHP", "Go", "Java", "TypeScript", "JavaScript"}).Build()

		status, body, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp := tagsUnmarshalHelper(t, body)
		assert.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, tagsResp, 5)
		require.Nil(t, tagsResp[0].Courses)

		// ----------------------------
		// Courses
		// ----------------------------

		status, body, err = tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?expand=true", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp = tagsUnmarshalHelper(t, body)
		assert.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, tagsResp, 5)
		require.Len(t, tagsResp[0].Courses, 2)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		daos.NewTestBuilder(t).
			Db(db).
			Courses([]string{"course 1", "course 2"}).
			Tags([]string{"PHP", "Go", "Java", "TypeScript", "JavaScript"}).
			Build()

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		status, body, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?orderBy=created_at%20asc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp := tagsUnmarshalHelper(t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, tagsResp, 5)
		assert.Equal(t, "PHP", tagsResp[0].Tag)

		// ----------------------------
		// CREATED_AT DESC
		// ----------------------------
		status, body, err = tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?orderBy=created_at%20desc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp = tagsUnmarshalHelper(t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, tagsResp, 5)
		assert.Equal(t, "JavaScript", tagsResp[0].Tag)

		// ----------------------------
		// CREATED_AT ASC + COURSES.TITLE DESC
		// ----------------------------
		courseDao := daos.NewCourseDao(db)

		status, body, err = tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?expand=true&orderBy=created_at%20asc,"+courseDao.Table+".title%20desc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp = tagsUnmarshalHelper(t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, tagsResp, 5)
		assert.Equal(t, "PHP", tagsResp[0].Tag)
		require.Len(t, tagsResp[0].Courses, 2)
		assert.Equal(t, "course 2", tagsResp[0].Courses[0].Title)
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
