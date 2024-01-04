package models

import (
	"database/sql"
	"testing"
	"time"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CountAssets(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		count, err := CountAssets(ctx, db, nil)
		require.Nil(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		NewTestAssets(t, db, []*Course{course}, 5)

		count, err := CountAssets(ctx, db, nil)
		require.Nil(t, err)
		assert.Equal(t, 5, count)
	})

	t.Run("where", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 3)
		assets := NewTestAssets(t, db, courses, 2)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: assets[1].ID}}}
		count, err := CountAssets(ctx, db, dbParams)
		require.Nil(t, err)
		assert.Equal(t, 1, count)

		dbParams = &database.DatabaseParams{Where: []database.Where{{Query: "? = ?", Column: "id", Value: assets[0].ID}}}
		count, err = CountAssets(ctx, db, dbParams)
		require.Nil(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Asset{}).Exec(ctx)
		require.Nil(t, err)

		_, err = CountAssets(ctx, db, nil)
		require.ErrorContains(t, err, "no such table: assets")
	})

	t.Run("error where", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "", Value: ""}}}
		count, err := CountAssets(ctx, db, dbParams)
		require.ErrorContains(t, err, "syntax error")
		assert.Equal(t, 0, count)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetAssets(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		assets, err := GetAssets(ctx, db, nil)
		require.Nil(t, err)
		require.Len(t, assets, 0)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		// 5 courses with 2 assets each (10 assets total)
		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)

		result, err := GetAssets(ctx, db, nil)
		require.Nil(t, err)
		require.Len(t, result, 10)
		assert.Equal(t, assets[0].ID, result[0].ID)
		assert.Equal(t, courses[0].ID, result[0].CourseID)
		assert.NotEmpty(t, result[0].Title)
		assert.Greater(t, result[0].Prefix, 0)
		assert.NotEmpty(t, result[0].Type)
		assert.NotEmpty(t, result[0].Path)
		assert.False(t, result[0].CreatedAt.IsZero())
		assert.False(t, result[0].UpdatedAt.IsZero())

		// Relations are empty
		assert.Nil(t, result[0].Course)
		assert.Nil(t, result[0].Attachments)
	})

	t.Run("relations", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		// ----------------------------
		// Course relation
		// ----------------------------
		relation := []database.Relation{{Struct: "Course"}}

		result, err := GetAssets(ctx, db, &database.DatabaseParams{Relation: relation})
		require.Nil(t, err)
		require.Len(t, result, 10)
		assert.Equal(t, assets[0].ID, result[0].ID)
		assert.Equal(t, assets[0].CourseID, result[0].CourseID)

		// Assert the course
		require.NotNil(t, result[0].Course)
		assert.Equal(t, courses[0].Title, result[0].Course.Title)

		// Assert attachments is nil
		assert.Nil(t, result[0].Attachments)

		// ----------------------------
		// Course and attachments relation
		// ----------------------------
		relation = []database.Relation{{Struct: "Course"}, {Struct: "Attachments"}}

		result, err = GetAssets(ctx, db, &database.DatabaseParams{Relation: relation})
		require.Nil(t, err)
		require.Len(t, result, 10)
		assert.Equal(t, assets[0].ID, result[0].ID)
		assert.Equal(t, assets[0].CourseID, result[0].CourseID)

		// Assert the course
		require.NotNil(t, result[0].Course)
		assert.Equal(t, courses[0].Title, result[0].Course.Title)

		// Assert the attachments
		require.NotNil(t, result[0].Attachments)
		require.Len(t, result[0].Attachments, 2)
		assert.Equal(t, attachments[0].ID, result[0].Attachments[0].ID)
	})

	t.Run("relations orderBy", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		dbParams := &database.DatabaseParams{
			Relation: []database.Relation{{Struct: "Course"}, {Struct: "Attachments", OrderBy: []string{"created_at desc"}}},
		}

		result, err := GetAssets(ctx, db, dbParams)
		require.Nil(t, err)
		assert.Equal(t, assets[0].ID, result[0].ID)
		assert.Equal(t, assets[0].CourseID, result[0].CourseID)

		// Assert the course
		require.NotNil(t, result[0].Course)
		assert.Equal(t, courses[0].Title, result[0].Course.Title)

		// Assert the attachments. The last attachment (for this asset) should be the first in the
		// list
		require.NotNil(t, result[0].Attachments)
		require.Len(t, result[0].Attachments, 2)
		assert.Equal(t, attachments[1].ID, result[0].Attachments[0].ID)
		assert.Equal(t, attachments[0].ID, result[0].Attachments[1].ID)
	})

	t.Run("pagination", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 4)
		assets := NewTestAssets(t, db, courses, 4)

		// Pagination context
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(c)

		// Page 1 with 10 items
		c.Request().SetRequestURI("/dummy?" + pagination.PageQueryParam + "=1" + "&" + pagination.PerPageQueryParam + "=10")
		p := pagination.New(c)

		// Assert the last asset in the paginated response
		result, err := GetAssets(ctx, db, &database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 16, p.TotalItems())
		assert.Equal(t, assets[9].ID, result[9].ID)
		assert.Equal(t, assets[9].Title, result[9].Title)
		assert.Equal(t, assets[9].Path, result[9].Path)

		// Page 2 with 6 items
		c.Request().SetRequestURI("/dummy?" + pagination.PageQueryParam + "=2" + "&" + pagination.PerPageQueryParam + "=10")
		p = pagination.New(c)

		// Assert the last asset in the paginated response
		result, err = GetAssets(ctx, db, &database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 6)
		require.Equal(t, 16, p.TotalItems())
		assert.Equal(t, assets[15].ID, result[5].ID)
		assert.Equal(t, assets[15].Title, result[5].Title)
		assert.Equal(t, assets[15].Path, result[5].Path)
	})

	t.Run("orderby", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)

		dbParams := &database.DatabaseParams{OrderBy: []string{"created_at desc"}}

		result, err := GetAssets(ctx, db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 10)
		assert.Equal(t, assets[9].ID, result[0].ID)
		assert.Equal(t, assets[9].CourseID, result[0].CourseID)
	})

	t.Run("where", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)

		// Get asset 3
		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: assets[2].ID}}}
		result, err := GetAssets(ctx, db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, assets[2].ID, result[0].ID)

		// Get all assets for course 1
		dbParams = &database.DatabaseParams{Where: []database.Where{{Query: "? = ?", Column: "course_id", Value: courses[0].ID}}}
		result, err = GetAssets(ctx, db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, assets[0].ID, result[0].ID)
		assert.Equal(t, assets[1].ID, result[1].ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Asset{}).Exec(ctx)
		require.Nil(t, err)

		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(c)

		c.Request().SetRequestURI("/dummy?" + pagination.PageQueryParam + "=1" + "&" + pagination.PerPageQueryParam + "=10")
		p := pagination.New(c)

		// With pagination
		_, err = GetAssets(ctx, db, &database.DatabaseParams{Pagination: p})
		require.ErrorContains(t, err, "no such table: assets")

		// Without pagination
		_, err = GetAssets(ctx, db, nil)
		require.ErrorContains(t, err, "no such table: assets")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetAsset(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: "1"}}}

		asset, err := GetAsset(ctx, db, dbParams)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, asset)
	})

	t.Run("found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: assets[2].ID}}}

		result, err := GetAsset(ctx, db, dbParams)
		require.Nil(t, err)
		assert.Equal(t, assets[2].ID, result.ID)
		assert.Equal(t, assets[2].CourseID, result.CourseID)
		assert.NotEmpty(t, result.Title)
		assert.Greater(t, result.Prefix, 0)
		assert.NotEmpty(t, result.Type)
		assert.NotEmpty(t, result.Path)
		assert.False(t, result.CreatedAt.IsZero())
		assert.False(t, result.UpdatedAt.IsZero())

		// Relations are empty
		assert.Nil(t, result.Course)
		assert.Nil(t, result.Attachments)
	})

	t.Run("relations", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		// ----------------------------
		// Course relation
		// ----------------------------
		dbParams := &database.DatabaseParams{
			Where:    []database.Where{{Column: "id", Value: assets[2].ID}},
			Relation: []database.Relation{{Struct: "Course"}},
		}

		result, err := GetAsset(ctx, db, dbParams)
		require.Nil(t, err)
		assert.Equal(t, assets[2].ID, result.ID)
		assert.Equal(t, assets[2].CourseID, result.CourseID)

		// Assert the course
		require.NotNil(t, result.Course)
		assert.Equal(t, courses[1].Title, result.Course.Title)

		// Assert attachments is nil
		assert.Nil(t, result.Attachments)

		// ----------------------------
		// Course and attachments relation
		// ----------------------------
		dbParams = &database.DatabaseParams{
			Where:    []database.Where{{Column: "id", Value: assets[6].ID}},
			Relation: []database.Relation{{Struct: "Course"}, {Struct: "Attachments"}},
		}

		result, err = GetAsset(ctx, db, dbParams)
		require.Nil(t, err)
		assert.Equal(t, assets[6].ID, result.ID)
		assert.Equal(t, assets[6].CourseID, result.CourseID)

		// Assert the course
		require.NotNil(t, result.Course)
		assert.Equal(t, courses[3].Title, result.Course.Title)

		// Assert the attachments
		require.NotNil(t, result.Attachments)
		require.Len(t, result.Attachments, 2)
		assert.Equal(t, attachments[12].ID, result.Attachments[0].ID)
	})

	t.Run("relations orderBy", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		dbParams := &database.DatabaseParams{
			Where:    []database.Where{{Column: "id", Value: assets[6].ID}},
			Relation: []database.Relation{{Struct: "Course"}, {Struct: "Attachments", OrderBy: []string{"created_at desc"}}},
		}

		result, err := GetAsset(ctx, db, dbParams)
		require.Nil(t, err)
		assert.Equal(t, assets[6].ID, result.ID)
		assert.Equal(t, assets[6].CourseID, result.CourseID)

		// Assert the course
		require.NotNil(t, result.Course)
		assert.Equal(t, courses[3].Title, result.Course.Title)

		// Assert the attachments. The last attachment (for this asset) should be the first in the
		// list
		require.NotNil(t, result.Attachments)
		require.Len(t, result.Attachments, 2)
		assert.Equal(t, attachments[13].ID, result.Attachments[0].ID)
		assert.Equal(t, attachments[12].ID, result.Attachments[1].ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Asset{}).Exec(ctx)
		require.Nil(t, err)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: "1"}}}
		_, err = GetAsset(ctx, db, dbParams)
		require.ErrorContains(t, err, "no such table: assets")
	})

	t.Run("missing where clause", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Asset{}).Exec(ctx)
		require.Nil(t, err)

		_, err = GetAsset(ctx, db, nil)
		require.ErrorContains(t, err, "where clause required")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetAssetById(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		asset, err := GetAssetById(ctx, db, nil, "1")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, asset)
	})

	t.Run("found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)

		result, err := GetAssetById(ctx, db, nil, assets[2].ID)
		require.Nil(t, err)
		assert.Equal(t, assets[2].ID, result.ID)
		assert.Equal(t, assets[2].CourseID, result.CourseID)
		assert.NotEmpty(t, result.Title)
		assert.Greater(t, result.Prefix, 0)
		assert.NotEmpty(t, result.Type)
		assert.NotEmpty(t, result.Path)
		assert.False(t, result.CreatedAt.IsZero())
		assert.False(t, result.UpdatedAt.IsZero())

		// Relations are empty
		assert.Nil(t, result.Course)
		assert.Nil(t, result.Attachments)
	})

	t.Run("relations", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		// ----------------------------
		// Course relation
		// ----------------------------
		dbParams := &database.DatabaseParams{Relation: []database.Relation{{Struct: "Course"}}}

		result, err := GetAssetById(ctx, db, dbParams, assets[2].ID)
		require.Nil(t, err)
		assert.Equal(t, assets[2].ID, result.ID)
		assert.Equal(t, assets[2].CourseID, result.CourseID)

		// Assert the course
		require.NotNil(t, result.Course)
		assert.Equal(t, courses[1].Title, result.Course.Title)

		// Assert attachments is nil
		assert.Nil(t, result.Attachments)

		// ----------------------------
		// Course and attachments relation
		// ----------------------------
		dbParams = &database.DatabaseParams{Relation: []database.Relation{{Struct: "Course"}, {Struct: "Attachments"}}}

		result, err = GetAssetById(ctx, db, dbParams, assets[6].ID)
		require.Nil(t, err)
		assert.Equal(t, assets[6].ID, result.ID)
		assert.Equal(t, assets[6].CourseID, result.CourseID)

		// Assert the course
		require.NotNil(t, result.Course)
		assert.Equal(t, courses[3].Title, result.Course.Title)

		// Assert the attachments
		require.NotNil(t, result.Attachments)
		require.Len(t, result.Attachments, 2)
		assert.Equal(t, attachments[12].ID, result.Attachments[0].ID)
	})

	t.Run("relations orderBy", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		dbParams := &database.DatabaseParams{
			Relation: []database.Relation{{Struct: "Course"}, {Struct: "Attachments", OrderBy: []string{"created_at desc"}}},
		}

		result, err := GetAssetById(ctx, db, dbParams, assets[6].ID)
		require.Nil(t, err)
		assert.Equal(t, assets[6].ID, result.ID)
		assert.Equal(t, assets[6].CourseID, result.CourseID)

		// Assert the course
		require.NotNil(t, result.Course)
		assert.Equal(t, courses[3].Title, result.Course.Title)

		// Assert the attachments. The last attachment (for this asset) should be the first in the
		// list
		require.NotNil(t, result.Attachments)
		require.Len(t, result.Attachments, 2)
		assert.Equal(t, attachments[13].ID, result.Attachments[0].ID)
		assert.Equal(t, attachments[12].ID, result.Attachments[1].ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Asset{}).Exec(ctx)
		require.Nil(t, err)

		_, err = GetAssetById(ctx, db, nil, "1")
		require.ErrorContains(t, err, "no such table: assets")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetAssetsByCourseId(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		assets, err := GetAssetsByCourseId(ctx, db, nil, "1")
		require.Nil(t, err)
		require.Len(t, assets, 0)
	})

	t.Run("found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)

		result, err := GetAssetsByCourseId(ctx, db, nil, courses[1].ID)
		require.Nil(t, err)
		require.Len(t, result, 2)

		// Assert the first asset in the result
		assert.Equal(t, assets[2].ID, result[0].ID)
		assert.Equal(t, assets[2].CourseID, result[0].CourseID)
		assert.NotEmpty(t, result[0].Title)
		assert.Greater(t, result[0].Prefix, 0)
		assert.NotEmpty(t, result[0].Type)
		assert.NotEmpty(t, result[0].Path)
		assert.False(t, result[0].CreatedAt.IsZero())
		assert.False(t, result[0].UpdatedAt.IsZero())

		// Relations are empty
		assert.Nil(t, result[0].Course)
		assert.Nil(t, result[0].Attachments)

		// Simple check on the second asset
		assert.Equal(t, assets[3].ID, result[1].ID)
	})

	t.Run("orderby", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)

		dbParams := &database.DatabaseParams{OrderBy: []string{"created_at desc"}}

		result, err := GetAssetsByCourseId(ctx, db, dbParams, courses[1].ID)
		require.Nil(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, assets[3].ID, result[0].ID)
		assert.Equal(t, assets[2].ID, result[1].ID)
	})

	t.Run("relations", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		// ----------------------------
		// Course relation
		// ----------------------------
		dbParams := &database.DatabaseParams{Relation: []database.Relation{{Struct: "Course"}}}

		result, err := GetAssetsByCourseId(ctx, db, dbParams, courses[1].ID)
		require.Nil(t, err)
		require.Len(t, result, 2)

		// Assert the first asset in the result
		assert.Equal(t, assets[2].ID, result[0].ID)
		assert.Equal(t, assets[2].CourseID, result[0].CourseID)

		// Assert the course
		require.NotNil(t, result[0].Course)
		assert.Equal(t, courses[1].Title, result[0].Course.Title)

		// Assert attachments is nil
		assert.Nil(t, result[0].Attachments)

		// ----------------------------
		// Course and attachments relation
		// ----------------------------
		dbParams = &database.DatabaseParams{Relation: []database.Relation{{Struct: "Course"}, {Struct: "Attachments"}}}

		result, err = GetAssetsByCourseId(ctx, db, dbParams, courses[3].ID)
		require.Nil(t, err)
		require.Len(t, result, 2)

		// Assert the first asset in the result
		assert.Equal(t, assets[6].ID, result[0].ID)
		assert.Equal(t, assets[6].CourseID, result[0].CourseID)

		// Assert the course
		require.NotNil(t, result[0].Course)
		assert.Equal(t, courses[3].Title, result[0].Course.Title)

		// Assert the attachments
		require.NotNil(t, result[0].Attachments)
		require.Len(t, result[0].Attachments, 2)
		assert.Equal(t, attachments[12].ID, result[0].Attachments[0].ID)
	})

	t.Run("relations orderBy", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		dbParams := &database.DatabaseParams{
			Relation: []database.Relation{{Struct: "Course"}, {Struct: "Attachments", OrderBy: []string{"created_at desc"}}},
		}

		result, err := GetAssetsByCourseId(ctx, db, dbParams, courses[3].ID)
		require.Nil(t, err)
		require.Len(t, result, 2)

		// Assert the first asset in the result
		assert.Equal(t, assets[6].ID, result[0].ID)
		assert.Equal(t, assets[6].CourseID, result[0].CourseID)

		// Assert the course
		require.NotNil(t, result[0].Course)
		assert.Equal(t, courses[3].Title, result[0].Course.Title)

		// Assert the attachments. The last attachment (for this asset) should be the first in the
		// list
		require.NotNil(t, result[0].Attachments)
		require.Len(t, result[0].Attachments, 2)
		assert.Equal(t, attachments[13].ID, result[0].Attachments[0].ID)
		assert.Equal(t, attachments[12].ID, result[0].Attachments[1].ID)
	})

	t.Run("pagination", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 2)
		assets := NewTestAssets(t, db, courses, 17)

		// Pagination context
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(c)

		// Page 1 with 10 items
		c.Request().SetRequestURI("/dummy?" + pagination.PageQueryParam + "=1" + "&" + pagination.PerPageQueryParam + "=10")
		p := pagination.New(c)

		// Assert the last asset in the paginated response
		result, err := GetAssetsByCourseId(ctx, db, &database.DatabaseParams{Pagination: p}, courses[1].ID)
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, assets[26].ID, result[9].ID)
		assert.Equal(t, assets[26].Title, result[9].Title)
		assert.Equal(t, assets[26].Path, result[9].Path)

		// Page 2 with 7 items
		c.Request().SetRequestURI("/dummy?" + pagination.PageQueryParam + "=2" + "&" + pagination.PerPageQueryParam + "=10")
		p = pagination.New(c)

		// Assert the last asset in the paginated response
		result, err = GetAssetsByCourseId(ctx, db, &database.DatabaseParams{Pagination: p}, courses[1].ID)
		require.Nil(t, err)
		require.Len(t, result, 7)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, assets[33].ID, result[6].ID)
		assert.Equal(t, assets[33].Title, result[6].Title)
		assert.Equal(t, assets[33].Path, result[6].Path)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Asset{}).Exec(ctx)
		require.Nil(t, err)

		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(c)

		c.Request().SetRequestURI("/dummy?" + pagination.PageQueryParam + "=1" + "&" + pagination.PerPageQueryParam + "=10")
		p := pagination.New(c)

		// With pagination
		_, err = GetAssets(ctx, db, &database.DatabaseParams{Pagination: p})
		require.ErrorContains(t, err, "no such table: assets")

		// Without pagination
		_, err = GetAssetsByCourseId(ctx, db, nil, "1")
		require.ErrorContains(t, err, "no such table: assets")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateAsset(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		asset := NewTestAssets(t, nil, []*Course{course}, 1)[0]

		err := CreateAsset(ctx, db, asset)
		require.Nil(t, err)
		assert.NotEmpty(t, asset.ID)
		assert.Equal(t, course.ID, asset.CourseID)
		assert.NotEmpty(t, asset.Title)
		assert.Greater(t, asset.Prefix, 0)
		assert.NotEmpty(t, asset.Type)
		assert.NotEmpty(t, asset.Path)
		assert.False(t, asset.CreatedAt.IsZero())
		assert.False(t, asset.UpdatedAt.IsZero())

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: asset.ID}}}

		_, err = GetAsset(ctx, db, dbParams)
		require.Nil(t, err)
	})

	t.Run("duplicate path", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		asset := NewTestAssets(t, nil, []*Course{course}, 1)[0]

		err := CreateAsset(ctx, db, asset)
		require.Nil(t, err)
		assert.NotEmpty(t, asset.ID)

		err = CreateAsset(ctx, db, asset)
		assert.ErrorContains(t, err, "UNIQUE constraint failed: assets.path")
	})

	t.Run("constraint errors", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		// Title
		asset := &Asset{}
		assert.ErrorContains(t, CreateAsset(ctx, db, asset), "NOT NULL constraint failed: assets.title")
		asset = &Asset{Title: ""}
		assert.ErrorContains(t, CreateAsset(ctx, db, asset), "NOT NULL constraint failed: assets.title")

		// Prefix
		asset = &Asset{Title: "Course 1"}
		assert.ErrorContains(t, CreateAsset(ctx, db, asset), "NOT NULL constraint failed: assets.prefix")
		asset = &Asset{Title: "Course 1", Prefix: -1}
		assert.ErrorContains(t, CreateAsset(ctx, db, asset), "prefix must be greater than 0")

		// Type
		asset = &Asset{Title: "Course 1", Prefix: 1}
		assert.ErrorContains(t, CreateAsset(ctx, db, asset), "NOT NULL constraint failed: assets.type")
		asset = &Asset{Title: "Course 1", Prefix: 1, Type: types.Asset{}}
		assert.ErrorContains(t, CreateAsset(ctx, db, asset), "NOT NULL constraint failed: assets.type")

		// Path
		asset = &Asset{Title: "Course 1", Prefix: 1, Type: *types.NewAsset("mp4")}
		assert.ErrorContains(t, CreateAsset(ctx, db, asset), "NOT NULL constraint failed: assets.path")
		asset = &Asset{Title: "Course 1", Prefix: 1, Type: *types.NewAsset("mp4"), Path: ""}
		assert.ErrorContains(t, CreateAsset(ctx, db, asset), "NOT NULL constraint failed: assets.path")

		// Course ID
		asset = &Asset{Title: "Course 1", Prefix: 1, Type: *types.NewAsset("mp4"), Path: "/course 1/01 asset"}
		assert.ErrorContains(t, CreateAsset(ctx, db, asset), "FOREIGN KEY constraint failed")
		asset = &Asset{CourseID: "", Title: "Course 1", Prefix: 1, Type: *types.NewAsset("mp4"), Path: "/course 1/01 asset"}
		assert.ErrorContains(t, CreateAsset(ctx, db, asset), "FOREIGN KEY constraint failed")
		asset = &Asset{CourseID: "1234", Title: "Course 1", Prefix: 1, Type: *types.NewAsset("mp4"), Path: "/course 1/01 asset"}
		assert.ErrorContains(t, CreateAsset(ctx, db, asset), "FOREIGN KEY constraint failed")

		// Success
		course := NewTestCourses(t, db, 1)[0]
		asset = &Asset{CourseID: course.ID, Title: "Course 1", Prefix: 1, Type: *types.NewAsset("mp4"), Path: "/course 1/01 asset"}
		assert.Nil(t, CreateAsset(ctx, db, asset))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteAsset(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		asset := NewTestAssets(t, db, []*Course{course}, 2)[0]

		count, err := DeleteAsset(ctx, db, asset.ID)
		require.Nil(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("cascade", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		NewTestAssets(t, db, []*Course{course}, 1)

		count, err := DeleteCourse(ctx, db, course.ID)
		require.Nil(t, err)
		require.Equal(t, 1, count)

		count, err = CountAssets(ctx, db, nil)
		require.Nil(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		// Invalid ID
		count, err := DeleteAsset(ctx, db, "1")
		require.Nil(t, err)
		assert.Equal(t, 0, count)

		// Drop the table
		_, err = db.DB().NewDropTable().Model(&Asset{}).Exec(ctx)
		require.Nil(t, err)

		count, err = DeleteAsset(ctx, db, "1")
		assert.ErrorContains(t, err, "no such table: assets")
		assert.Equal(t, 0, count)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateAssetProgress(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		origAsset := NewTestAssets(t, db, []*Course{course}, 1)[0]
		require.Zero(t, origAsset.Progress)

		// ----------------------------
		// Set to 30
		// ----------------------------

		// Give time to allow `updated at` to be different
		time.Sleep(time.Millisecond * 1)

		a1, err := UpdateAssetProgress(ctx, db, origAsset.ID, 30)
		require.Nil(t, err)
		assert.Equal(t, 30, a1.Progress)
		assert.NotEqual(t, origAsset.UpdatedAt.String(), a1.UpdatedAt.String())

		// ----------------------------
		// Set to 0 (by setting a negative value)
		// ----------------------------

		// Give time to allow `updated at` to be different
		time.Sleep(time.Millisecond * 1)

		a2, err := UpdateAssetProgress(ctx, db, origAsset.ID, -1)
		require.Nil(t, err)
		assert.Zero(t, a2.Progress)
		assert.NotEqual(t, a1.UpdatedAt.String(), a2.UpdatedAt.String())
	})

	t.Run("no change", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		origAsset := NewTestAssets(t, db, []*Course{course}, 1)[0]
		require.Zero(t, origAsset.Progress)

		// Give time to allow `updated at` to be different
		time.Sleep(time.Millisecond * 1)

		a, err := UpdateAssetProgress(ctx, db, origAsset.ID, 0)
		require.Nil(t, err)
		assert.Zero(t, a.Progress)
		assert.Equal(t, origAsset.UpdatedAt.String(), a.UpdatedAt.String())
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		a, err := UpdateAssetProgress(ctx, db, "", 30)
		assert.ErrorContains(t, err, "asset ID cannot be empty")
		assert.Nil(t, a)
	})

	t.Run("no asset with id", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		a, err := UpdateAssetProgress(ctx, db, "1234", 30)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, a)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Asset{}).Exec(ctx)
		require.Nil(t, err)

		a, err := UpdateAssetProgress(ctx, db, "1234", 30)
		require.ErrorContains(t, err, "no such table: assets")
		assert.Nil(t, a)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateAssetCompleted(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		assets := NewTestAssets(t, db, []*Course{course}, 10)
		require.False(t, assets[0].Completed)
		require.Empty(t, assets[0].CompletedAt)

		// Give time to allow `updated at` to be different
		time.Sleep(time.Millisecond * 1)

		// ----------------------------
		// Set asset 1 to true
		// ----------------------------
		a1, err := UpdateAssetCompleted(ctx, db, assets[0].ID, true)
		require.Nil(t, err)
		require.NotNil(t, a1)
		assert.True(t, a1.Completed)
		assert.NotEmpty(t, a1.CompletedAt)
		assert.NotEqual(t, assets[0].UpdatedAt, a1.UpdatedAt)

		// Asset the course percentage is updated
		c1, err := GetCourseById(ctx, db, nil, course.ID)
		require.Nil(t, err)
		assert.Equal(t, 10, c1.Percent)

		// Give time to allow `updated at` to be different
		time.Sleep(time.Millisecond * 1)

		// ----------------------------
		// Set asset 2 to true
		// ----------------------------
		a2, err := UpdateAssetCompleted(ctx, db, assets[1].ID, true)
		require.Nil(t, err)
		require.NotNil(t, a2)
		assert.True(t, a2.Completed)
		assert.NotEmpty(t, a2.CompletedAt)
		assert.NotEqual(t, a1.UpdatedAt, a2.UpdatedAt)

		// Asset the course percentage is updated
		c2, err := GetCourseById(ctx, db, nil, course.ID)
		require.Nil(t, err)
		assert.Equal(t, 20, c2.Percent)

		// Give time to allow `updated at` to be different
		time.Sleep(time.Millisecond * 1)

		// ----------------------------
		// Set asset 2 to false
		// ----------------------------
		a3, err := UpdateAssetCompleted(ctx, db, assets[1].ID, false)
		require.Nil(t, err)
		require.NotNil(t, a3)
		assert.False(t, a3.Completed)
		assert.Empty(t, a3.CompletedAt)
		assert.NotEqual(t, a2.UpdatedAt, a3.UpdatedAt)

		// Asset the course percentage is updated
		c3, err := GetCourseById(ctx, db, nil, course.ID)
		require.Nil(t, err)
		assert.Equal(t, 10, c3.Percent)
	})

	t.Run("no change", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		origAsset := NewTestAssets(t, db, []*Course{course}, 1)[0]
		require.False(t, origAsset.Completed)
		require.Empty(t, origAsset.CompletedAt)

		// Give time to allow `updated at` to be different
		time.Sleep(time.Millisecond * 1)

		a, err := UpdateAssetCompleted(ctx, db, origAsset.ID, false)
		require.Nil(t, err)
		assert.False(t, a.Completed)
		assert.Empty(t, a.CompletedAt)
		assert.Equal(t, origAsset.UpdatedAt.String(), a.UpdatedAt.String())
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		a, err := UpdateAssetCompleted(ctx, db, "", true)
		assert.ErrorContains(t, err, "asset ID cannot be empty")
		assert.Nil(t, a)
	})

	t.Run("no asset with id", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		a, err := UpdateAssetCompleted(ctx, db, "1234", true)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, a)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Asset{}).Exec(ctx)
		require.Nil(t, err)

		a, err := UpdateAssetCompleted(ctx, db, "1234", true)
		require.ErrorContains(t, err, "no such table: assets")
		assert.Nil(t, a)
	})
}
