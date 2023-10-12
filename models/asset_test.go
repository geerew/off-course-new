package models

import (
	"database/sql"
	"testing"

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

		count, err := CountAssets(db, nil, ctx)
		require.Nil(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		NewTestAssets(t, db, []*Course{course}, 5)

		count, err := CountAssets(db, nil, ctx)
		require.Nil(t, err)
		assert.Equal(t, 5, count)
	})

	t.Run("where", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 3)
		assets := NewTestAssets(t, db, courses, 2)

		where := []database.Where{{Column: "id", Value: assets[1].ID}}
		count, err := CountAssets(db, &database.DatabaseParams{Where: where}, ctx)
		require.Nil(t, err)
		assert.Equal(t, 1, count)

		where = []database.Where{{Query: "? = ?", Column: "id", Value: assets[0].ID}}
		count, err = CountAssets(db, &database.DatabaseParams{Where: where}, ctx)
		require.Nil(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Asset{}).Exec(ctx)
		require.Nil(t, err)

		_, err = CountAssets(db, nil, ctx)
		require.ErrorContains(t, err, "no such table: assets")
	})

	t.Run("error where", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		where := []database.Where{{Column: "", Value: ""}}
		count, err := CountAssets(db, &database.DatabaseParams{Where: where}, ctx)
		require.ErrorContains(t, err, "syntax error")
		assert.Equal(t, 0, count)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetAssets(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		assets, err := GetAssets(db, nil, ctx)
		require.Nil(t, err)
		require.Len(t, assets, 0)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		// 5 courses with 2 assets each (10 assets total)
		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)

		result, err := GetAssets(db, nil, ctx)
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

		result, err := GetAssets(db, &database.DatabaseParams{Relation: relation}, ctx)
		require.Nil(t, err)
		require.Len(t, result, 10)
		assert.Equal(t, assets[0].ID, result[0].ID)
		assert.Equal(t, assets[0].CourseID, result[0].CourseID)

		// Assert the course
		require.NotNil(t, result[0].Course)
		assert.Equal(t, courses[0].Title, result[0].Course.Title)

		// ----------------------------
		// Course and attachments relation
		// ----------------------------
		relation = []database.Relation{{Struct: "Course"}, {Struct: "Attachments"}}

		result, err = GetAssets(db, &database.DatabaseParams{Relation: relation}, ctx)
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
		result, err := GetAssets(db, &database.DatabaseParams{Pagination: p}, ctx)
		require.Nil(t, err)
		require.Len(t, result, 10)
		assert.Equal(t, assets[9].ID, result[9].ID)
		assert.Equal(t, assets[9].Title, result[9].Title)
		assert.Equal(t, assets[9].Path, result[9].Path)

		// Page 2 with 10 items
		c.Request().SetRequestURI("/dummy?" + pagination.PageQueryParam + "=2" + "&" + pagination.PerPageQueryParam + "=10")
		p = pagination.New(c)

		// Assert the last asset in the paginated response
		result, err = GetAssets(db, &database.DatabaseParams{Pagination: p}, ctx)
		require.Nil(t, err)
		require.Len(t, result, 6)
		assert.Equal(t, assets[15].ID, result[5].ID)
		assert.Equal(t, assets[15].Title, result[5].Title)
		assert.Equal(t, assets[15].Path, result[5].Path)
	})

	t.Run("orderby", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)

		result, err := GetAssets(db, &database.DatabaseParams{OrderBy: []string{"created_at desc"}}, ctx)
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
		where := []database.Where{{Column: "id", Value: assets[2].ID}}

		result, err := GetAssets(db, &database.DatabaseParams{Where: where}, ctx)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, assets[2].ID, result[0].ID)

		// Get all assets for course 1
		where = []database.Where{{Query: "? = ?", Column: "course_id", Value: courses[0].ID}}

		result, err = GetAssets(db, &database.DatabaseParams{Where: where}, ctx)
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
		_, err = GetAssets(db, &database.DatabaseParams{Pagination: p}, ctx)
		require.ErrorContains(t, err, "no such table: assets")

		// Without pagination
		_, err = GetAssets(db, nil, ctx)
		require.ErrorContains(t, err, "no such table: assets")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetAsset(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		asset, err := GetAssetById(db, "1", nil, ctx)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, asset)
	})

	t.Run("found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)

		result, err := GetAssetById(db, assets[2].ID, nil, ctx)
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

	t.Run("course relation", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)

		relation := []database.Relation{{Struct: "Course"}}

		result, err := GetAssetById(db, assets[2].ID, &database.DatabaseParams{Relation: relation}, ctx)
		require.Nil(t, err)
		assert.Equal(t, assets[2].ID, result.ID)
		assert.Equal(t, assets[2].CourseID, result.CourseID)

		// Assert the course
		require.NotNil(t, result.Course)
		assert.Equal(t, courses[1].Title, result.Course.Title)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Asset{}).Exec(ctx)
		require.Nil(t, err)

		_, err = GetAssetById(db, "1", nil, ctx)
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

		err := CreateAsset(db, asset, ctx)
		require.Nil(t, err)
		assert.NotEmpty(t, asset.ID)
		assert.Equal(t, course.ID, asset.CourseID)
		assert.NotEmpty(t, asset.Title)
		assert.Greater(t, asset.Prefix, 0)
		assert.NotEmpty(t, asset.Type)
		assert.NotEmpty(t, asset.Path)
		assert.False(t, asset.CreatedAt.IsZero())
		assert.False(t, asset.UpdatedAt.IsZero())

		_, err = GetAssetById(db, asset.ID, nil, ctx)
		require.Nil(t, err)
	})

	t.Run("duplicate path", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		asset := NewTestAssets(t, nil, []*Course{course}, 1)[0]

		err := CreateAsset(db, asset, ctx)
		require.Nil(t, err)
		assert.NotEmpty(t, asset.ID)

		err = CreateAsset(db, asset, ctx)
		assert.ErrorContains(t, err, "UNIQUE constraint failed: assets.path")
	})

	t.Run("constraint errors", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		// Title
		asset := &Asset{}
		assert.ErrorContains(t, CreateAsset(db, asset, ctx), "NOT NULL constraint failed: assets.title")
		asset = &Asset{Title: ""}
		assert.ErrorContains(t, CreateAsset(db, asset, ctx), "NOT NULL constraint failed: assets.title")

		// Prefix
		asset = &Asset{Title: "Course 1"}
		assert.ErrorContains(t, CreateAsset(db, asset, ctx), "NOT NULL constraint failed: assets.prefix")
		asset = &Asset{Title: "Course 1", Prefix: -1}
		assert.ErrorContains(t, CreateAsset(db, asset, ctx), "prefix must be greater than 0")

		// Type
		asset = &Asset{Title: "Course 1", Prefix: 1}
		assert.ErrorContains(t, CreateAsset(db, asset, ctx), "NOT NULL constraint failed: assets.type")
		asset = &Asset{Title: "Course 1", Prefix: 1, Type: types.Asset{}}
		assert.ErrorContains(t, CreateAsset(db, asset, ctx), "NOT NULL constraint failed: assets.type")

		// Path
		asset = &Asset{Title: "Course 1", Prefix: 1, Type: *types.NewAsset("mp4")}
		assert.ErrorContains(t, CreateAsset(db, asset, ctx), "NOT NULL constraint failed: assets.path")
		asset = &Asset{Title: "Course 1", Prefix: 1, Type: *types.NewAsset("mp4"), Path: ""}
		assert.ErrorContains(t, CreateAsset(db, asset, ctx), "NOT NULL constraint failed: assets.path")

		// Course ID
		asset = &Asset{Title: "Course 1", Prefix: 1, Type: *types.NewAsset("mp4"), Path: "/course 1/01 asset"}
		assert.ErrorContains(t, CreateAsset(db, asset, ctx), "FOREIGN KEY constraint failed")
		asset = &Asset{CourseID: "", Title: "Course 1", Prefix: 1, Type: *types.NewAsset("mp4"), Path: "/course 1/01 asset"}
		assert.ErrorContains(t, CreateAsset(db, asset, ctx), "FOREIGN KEY constraint failed")
		asset = &Asset{CourseID: "1234", Title: "Course 1", Prefix: 1, Type: *types.NewAsset("mp4"), Path: "/course 1/01 asset"}
		assert.ErrorContains(t, CreateAsset(db, asset, ctx), "FOREIGN KEY constraint failed")

		// Success
		course := NewTestCourses(t, db, 1)[0]
		asset = &Asset{CourseID: course.ID, Title: "Course 1", Prefix: 1, Type: *types.NewAsset("mp4"), Path: "/course 1/01 asset"}
		assert.Nil(t, CreateAsset(db, asset, ctx))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteAsset(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		asset := NewTestAssets(t, db, []*Course{course}, 2)[0]

		count, err := DeleteAsset(db, asset.ID, ctx)
		require.Nil(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("cascade", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		NewTestAssets(t, db, []*Course{course}, 1)

		count, err := DeleteCourse(db, course.ID, ctx)
		require.Nil(t, err)
		require.Equal(t, 1, count)

		count, err = CountAssets(db, nil, ctx)
		require.Nil(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		// Invalid ID
		count, err := DeleteAsset(db, "1", ctx)
		require.Nil(t, err)
		assert.Equal(t, 0, count)

		// Drop the table
		_, err = db.DB().NewDropTable().Model(&Asset{}).Exec(ctx)
		require.Nil(t, err)

		count, err = DeleteAsset(db, "1", ctx)
		assert.ErrorContains(t, err, "no such table: assets")
		assert.Equal(t, 0, count)
	})
}