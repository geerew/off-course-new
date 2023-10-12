package models

import (
	"database/sql"
	"testing"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CountAttachments(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		count, err := CountAttachments(db, nil, ctx)
		require.Nil(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		assets := NewTestAssets(t, db, []*Course{course}, 2)
		NewTestAttachments(t, db, assets, 2)

		count, err := CountAttachments(db, nil, ctx)
		require.Nil(t, err)
		assert.Equal(t, 4, count)
	})

	t.Run("where", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 3)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: attachments[1].ID}}}
		count, err := CountAttachments(db, dbParams, ctx)
		require.Nil(t, err)
		assert.Equal(t, 1, count)

		dbParams = &database.DatabaseParams{Where: []database.Where{{Query: "? = ?", Column: "id", Value: attachments[0].ID}}}
		count, err = CountAttachments(db, dbParams, ctx)
		require.Nil(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Attachment{}).Exec(ctx)
		require.Nil(t, err)

		_, err = CountAttachments(db, nil, ctx)
		require.ErrorContains(t, err, "no such table: attachments")
	})

	t.Run("error where", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "", Value: ""}}}
		count, err := CountAssets(db, dbParams, ctx)
		require.ErrorContains(t, err, "syntax error")
		assert.Equal(t, 0, count)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetAttachments(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		attachments, err := GetAttachments(db, nil, ctx)
		require.Nil(t, err)
		require.Len(t, attachments, 0)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		// 5 courses with 2 assets and 2 attachments (20 attachments total)
		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		result, err := GetAttachments(db, nil, ctx)
		require.Nil(t, err)
		require.Len(t, result, 20)
		assert.Equal(t, attachments[0].ID, result[0].ID)
		assert.Equal(t, courses[0].ID, result[0].CourseID)
		assert.Equal(t, assets[0].ID, result[0].AssetID)
		assert.NotEmpty(t, result[0].Title)
		assert.NotEmpty(t, result[0].Path)
		assert.False(t, result[0].CreatedAt.IsZero())
		assert.False(t, result[0].UpdatedAt.IsZero())
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

		result, err := GetAttachments(db, &database.DatabaseParams{Relation: relation}, ctx)
		require.Nil(t, err)
		require.Len(t, result, 20)
		assert.Equal(t, attachments[0].ID, result[0].ID)
		assert.Equal(t, attachments[0].CourseID, result[0].CourseID)
		assert.Equal(t, attachments[0].AssetID, result[0].AssetID)

		// Assert the course
		require.NotNil(t, result[0].Course)
		assert.Equal(t, courses[0].ID, result[0].Course.ID)

		// Assert asset is nil
		assert.Nil(t, result[0].Asset)

		// ----------------------------
		// Course and asset relation
		// ----------------------------
		relation = []database.Relation{{Struct: "Course"}, {Struct: "Asset"}}

		result, err = GetAttachments(db, &database.DatabaseParams{Relation: relation}, ctx)
		require.Len(t, result, 20)
		require.Nil(t, err)
		assert.Equal(t, attachments[0].ID, result[0].ID)
		assert.Equal(t, attachments[0].CourseID, result[0].CourseID)
		assert.Equal(t, attachments[0].AssetID, result[0].AssetID)

		// Assert the course
		require.NotNil(t, result[0].Course)
		assert.Equal(t, courses[0].ID, result[0].Course.ID)

		// Assert the asset
		require.NotNil(t, result[0].Asset)
		assert.Equal(t, assets[0].ID, result[0].Asset.ID)
	})

	t.Run("pagination", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 2)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 4)

		// Pagination context
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(c)

		// Page 1 with 10 items
		c.Request().SetRequestURI("/dummy?" + pagination.PageQueryParam + "=1" + "&" + pagination.PerPageQueryParam + "=10")
		p := pagination.New(c)

		// Assert the last attachment in the paginated response
		result, err := GetAttachments(db, &database.DatabaseParams{Pagination: p}, ctx)
		require.Nil(t, err)
		require.Len(t, result, 10)
		assert.Equal(t, attachments[9].ID, result[9].ID)
		assert.Equal(t, attachments[9].CourseID, result[9].CourseID)
		assert.Equal(t, attachments[9].AssetID, result[9].AssetID)

		// Page 2 with 10 items
		c.Request().SetRequestURI("/dummy?" + pagination.PageQueryParam + "=2" + "&" + pagination.PerPageQueryParam + "=10")
		p = pagination.New(c)

		// Assert the last attachment in the paginated response
		result, err = GetAttachments(db, &database.DatabaseParams{Pagination: p}, ctx)
		require.Nil(t, err)
		require.Len(t, result, 6)
		assert.Equal(t, attachments[15].ID, result[5].ID)
		assert.Equal(t, attachments[15].CourseID, result[5].CourseID)
		assert.Equal(t, attachments[15].AssetID, result[5].AssetID)
	})

	t.Run("orderby", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		result, err := GetAttachments(db, &database.DatabaseParams{OrderBy: []string{"created_at desc"}}, ctx)
		require.Nil(t, err)
		require.Len(t, result, 20)
		assert.Equal(t, attachments[19].ID, result[0].ID)
		assert.Equal(t, attachments[19].CourseID, result[0].CourseID)
	})

	t.Run("where", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Query: "? = ?", Column: "id", Value: attachments[2].ID}}}
		result, err := GetAttachments(db, dbParams, ctx)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, attachments[2].ID, result[0].ID)

		dbParams = &database.DatabaseParams{Where: []database.Where{{Query: "? = ?", Column: "id", Value: attachments[3].ID}}}
		result, err = GetAttachments(db, dbParams, ctx)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, attachments[3].ID, result[0].ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Attachment{}).Exec(ctx)
		require.Nil(t, err)

		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(c)

		c.Request().SetRequestURI("/dummy?" + pagination.PageQueryParam + "=1" + "&" + pagination.PerPageQueryParam + "=10")
		p := pagination.New(c)

		// With pagination
		_, err = GetAttachments(db, &database.DatabaseParams{Pagination: p}, ctx)
		require.ErrorContains(t, err, "no such table: attachments")

		// Without pagination
		_, err = GetAttachments(db, nil, ctx)
		require.ErrorContains(t, err, "no such table: attachments")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetAttachment(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: "1"}}}

		attachment, err := GetAttachment(db, dbParams, ctx)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, attachment)
	})

	t.Run("found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: attachments[5].ID}}}

		result, err := GetAttachment(db, dbParams, ctx)
		require.Nil(t, err)
		assert.Equal(t, attachments[5].ID, result.ID)
		assert.Equal(t, attachments[5].CourseID, result.CourseID)
		assert.NotEmpty(t, result.Title)
		assert.NotEmpty(t, result.Path)
		assert.False(t, result.CreatedAt.IsZero())
		assert.False(t, result.UpdatedAt.IsZero())

		// Relations are empty
		assert.Nil(t, result.Course)
		assert.Nil(t, result.Asset)
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
			Where:    []database.Where{{Column: "attachment.id", Value: attachments[5].ID}},
			Relation: []database.Relation{{Struct: "Course"}},
		}

		result, err := GetAttachment(db, dbParams, ctx)
		require.Nil(t, err)
		assert.Equal(t, attachments[5].ID, result.ID)
		assert.Equal(t, attachments[5].CourseID, result.CourseID)
		assert.Equal(t, attachments[5].AssetID, result.AssetID)

		// Assert the course
		require.NotNil(t, result.Course)
		assert.Equal(t, courses[1].ID, result.Course.ID)

		// Assert asset is nil
		assert.Nil(t, result.Asset)

		// ----------------------------
		// Course and asset relation
		// ----------------------------
		dbParams = &database.DatabaseParams{
			Where:    []database.Where{{Column: "attachment.id", Value: attachments[12].ID}},
			Relation: []database.Relation{{Struct: "Course"}, {Struct: "Asset"}},
		}

		result, err = GetAttachment(db, dbParams, ctx)
		require.Nil(t, err)
		assert.Equal(t, attachments[12].ID, result.ID)
		assert.Equal(t, attachments[12].CourseID, result.CourseID)
		assert.Equal(t, attachments[12].AssetID, result.AssetID)

		// Assert the course
		require.NotNil(t, result.Course)
		assert.Equal(t, courses[3].ID, result.Course.ID)

		// Assert the asset
		require.NotNil(t, result.Asset)
		assert.Equal(t, assets[6].ID, result.Asset.ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Attachment{}).Exec(ctx)
		require.Nil(t, err)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: "1"}}}

		_, err = GetAttachment(db, dbParams, ctx)
		require.ErrorContains(t, err, "no such table: attachments")
	})

	t.Run("missing where clause", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Course{}).Exec(ctx)
		require.Nil(t, err)

		_, err = GetAttachment(db, nil, ctx)
		require.ErrorContains(t, err, "where clause required")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateAttachment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		asset := NewTestAssets(t, db, []*Course{course}, 1)[0]
		attachment := NewTestAttachments(t, nil, []*Asset{asset}, 1)[0]

		err := CreateAttachment(db, attachment, ctx)
		require.Nil(t, err)
		assert.NotEmpty(t, attachment.ID)
		assert.Equal(t, course.ID, attachment.CourseID)
		assert.Equal(t, asset.ID, attachment.AssetID)
		assert.NotEmpty(t, attachment.Title)
		assert.NotEmpty(t, attachment.Path)
		assert.False(t, attachment.CreatedAt.IsZero())
		assert.False(t, attachment.UpdatedAt.IsZero())

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: attachment.ID}}}
		_, err = GetAttachment(db, dbParams, ctx)
		require.Nil(t, err)
	})

	t.Run("duplicate course", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		asset := NewTestAssets(t, db, []*Course{course}, 1)[0]
		attachment := NewTestAttachments(t, nil, []*Asset{asset}, 1)[0]

		err := CreateAttachment(db, attachment, ctx)
		require.Nil(t, err)
		assert.NotEmpty(t, asset.ID)

		err = CreateAttachment(db, attachment, ctx)
		assert.ErrorContains(t, err, "UNIQUE constraint failed: attachments.path")
	})

	t.Run("constraint errors", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		// Title
		attachment := &Attachment{}
		assert.ErrorContains(t, CreateAttachment(db, attachment, ctx), "NOT NULL constraint failed: attachments.title")
		attachment = &Attachment{Title: ""}
		assert.ErrorContains(t, CreateAttachment(db, attachment, ctx), "NOT NULL constraint failed: attachments.title")

		// Path
		attachment = &Attachment{Title: "Course 1"}
		assert.ErrorContains(t, CreateAttachment(db, attachment, ctx), "NOT NULL constraint failed: attachments.path")
		attachment = &Attachment{Title: "Course 1", Path: ""}
		assert.ErrorContains(t, CreateAttachment(db, attachment, ctx), "NOT NULL constraint failed: attachments.path")

		// Course ID
		attachment = &Attachment{Title: "Course 1", Path: "/course 1/01 attachment"}
		assert.ErrorContains(t, CreateAttachment(db, attachment, ctx), "FOREIGN KEY constraint failed")
		attachment = &Attachment{CourseID: "", Title: "Course 1", Path: "/course 1/01 attachment"}
		assert.ErrorContains(t, CreateAttachment(db, attachment, ctx), "FOREIGN KEY constraint failed")
		attachment = &Attachment{CourseID: "1234", Title: "Course 1", Path: "/course 1/01 attachment"}
		assert.ErrorContains(t, CreateAttachment(db, attachment, ctx), "FOREIGN KEY constraint failed")

		course := NewTestCourses(t, db, 1)[0]

		// Asset ID
		attachment = &Attachment{CourseID: course.ID, Title: "Course 1", Path: "/course 1/01 attachment"}
		assert.ErrorContains(t, CreateAttachment(db, attachment, ctx), "FOREIGN KEY constraint failed")
		attachment = &Attachment{CourseID: course.ID, AssetID: "", Title: "Course 1", Path: "/course 1/01 attachment"}
		assert.ErrorContains(t, CreateAttachment(db, attachment, ctx), "FOREIGN KEY constraint failed")
		attachment = &Attachment{CourseID: course.ID, AssetID: "1234", Title: "Course 1", Path: "/course 1/01 attachment"}
		assert.ErrorContains(t, CreateAttachment(db, attachment, ctx), "FOREIGN KEY constraint failed")

		// Success
		asset := NewTestAssets(t, db, []*Course{course}, 1)[0]
		attachment = &Attachment{CourseID: course.ID, AssetID: asset.ID, Title: "Course 1", Path: "/course 1/01 attachment"}
		assert.Nil(t, CreateAttachment(db, attachment, ctx))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteAttachment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		assets := NewTestAssets(t, db, []*Course{course}, 2)
		attachment := NewTestAttachments(t, db, assets, 1)[0]

		count, err := DeleteAttachment(db, attachment.ID, ctx)
		require.Nil(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("cascade", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		// ----------------------------
		// Course deleted
		// ----------------------------
		course := NewTestCourses(t, db, 1)[0]
		asset := NewTestAssets(t, db, []*Course{course}, 1)[0]
		NewTestAttachments(t, db, []*Asset{asset}, 5)

		count, err := DeleteCourse(db, course.ID, ctx)
		require.Nil(t, err)
		require.Equal(t, 1, count)

		count, err = CountAttachments(db, nil, ctx)
		require.Nil(t, err)
		assert.Equal(t, 0, count)

		// ----------------------------
		// Asset deleted
		// ----------------------------
		course = NewTestCourses(t, db, 1)[0]
		asset = NewTestAssets(t, db, []*Course{course}, 1)[0]
		NewTestAttachments(t, db, []*Asset{asset}, 5)

		count, err = DeleteAsset(db, asset.ID, ctx)
		require.Nil(t, err)
		require.Equal(t, 1, count)

		count, err = CountAttachments(db, nil, ctx)
		require.Nil(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		// Invalid ID
		count, err := DeleteAttachment(db, "1", ctx)
		require.Nil(t, err)
		assert.Equal(t, 0, count)

		// Drop the table
		_, err = db.DB().NewDropTable().Model(&Attachment{}).Exec(ctx)
		require.Nil(t, err)

		count, err = DeleteAttachment(db, "1", ctx)
		assert.ErrorContains(t, err, "no such table: attachments")
		assert.Equal(t, 0, count)
	})
}
