package models

import (
	"database/sql"
	"testing"
	"time"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CountCourses(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		count, err := CountCourses(db, nil, ctx)
		require.Nil(t, err)
		assert.Equal(t, count, 0)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		NewTestCourses(t, db, 5)

		count, err := CountCourses(db, nil, ctx)
		require.Nil(t, err)
		assert.Equal(t, count, 5)
	})

	t.Run("where", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)

		where := []database.Where{{Column: "id", Value: courses[1].ID}}
		count, err := CountCourses(db, &database.DatabaseParams{Where: where}, ctx)
		require.Nil(t, err)
		assert.Equal(t, 1, count)

		where = []database.Where{{Query: "? = ?", Column: "id", Value: courses[0].ID}}
		count, err = CountCourses(db, &database.DatabaseParams{Where: where}, ctx)
		require.Nil(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model((*Course)(nil)).Exec(ctx)
		require.Nil(t, err)

		_, err = CountCourses(db, nil, ctx)
		require.ErrorContains(t, err, "no such table: courses")
	})

	t.Run("error where", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		where := []database.Where{{Column: "", Value: ""}}
		count, err := CountCourses(db, &database.DatabaseParams{Where: where}, ctx)
		require.ErrorContains(t, err, "syntax error")
		assert.Equal(t, 0, count)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetCourses(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses, err := GetCourses(db, nil, ctx)
		require.Nil(t, err)
		require.Len(t, courses, 0)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		NewTestScans(t, db, courses)

		result, err := GetCourses(db, nil, ctx)
		require.Nil(t, err)
		require.Len(t, result, 5)
		assert.Equal(t, courses[0].ID, result[0].ID)
		assert.Equal(t, courses[0].Title, result[0].Title)
		assert.Equal(t, courses[0].Path, result[0].Path)
		assert.Empty(t, courses[0].CardPath)
		assert.False(t, courses[0].Started)
		assert.False(t, courses[0].Finished)
		assert.False(t, courses[0].CreatedAt.IsZero())
		assert.False(t, courses[0].UpdatedAt.IsZero())

		// Scan status
		require.NotEmpty(t, result[0].ScanStatus)
		assert.Equal(t, "waiting", result[0].ScanStatus)

		// Relations are empty
		assert.Nil(t, result[0].Assets)
	})

	t.Run("relations", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		// ----------------------------
		// Assets relation
		// ----------------------------
		relation := []database.Relation{{Struct: "Assets"}}

		result, err := GetCourses(db, &database.DatabaseParams{Relation: relation}, ctx)
		require.Nil(t, err)
		require.Len(t, result, 5)
		assert.Equal(t, courses[0].ID, result[0].ID)

		// Assert the assets
		require.NotNil(t, result[0].Assets)
		require.Len(t, result[0].Assets, 2)
		assert.Equal(t, assets[0].ID, result[0].Assets[0].ID)

		// Asset the attachments for the first asset is nil
		assert.Nil(t, result[0].Assets[0].Attachments)

		// ----------------------------
		// Assets and attachments relation
		// ----------------------------
		relation = []database.Relation{{Struct: "Assets"}, {Struct: "Assets.Attachments"}}

		result, err = GetCourses(db, &database.DatabaseParams{Relation: relation}, ctx)
		require.Nil(t, err)
		require.Len(t, result, 5)
		assert.Equal(t, courses[0].ID, result[0].ID)

		// Assert the assets
		require.NotNil(t, result[0].Assets)
		require.Len(t, result[0].Assets, 2)
		assert.Equal(t, assets[0].ID, result[0].Assets[0].ID)

		// Asset the attachments
		assert.NotNil(t, result[0].Assets[0].Attachments)
		assert.Len(t, result[0].Assets[0].Attachments, 2)
		assert.Equal(t, attachments[0].ID, result[0].Assets[0].Attachments[0].ID)
	})

	t.Run("pagination", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 17)

		// Pagination context
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(c)

		// Page 1 with 10 items
		c.Request().SetRequestURI("/dummy?" + pagination.PageQueryParam + "=1" + "&" + pagination.PerPageQueryParam + "=10")
		p := pagination.New(c)

		// Assert the last course in the pagination response
		result, err := GetCourses(db, &database.DatabaseParams{Pagination: p}, ctx)
		require.Nil(t, err)
		require.Len(t, result, 10)
		assert.Equal(t, courses[9].ID, result[9].ID)
		assert.Equal(t, courses[9].Title, result[9].Title)
		assert.Equal(t, courses[9].Path, result[9].Path)

		// Page 2 with 10 items
		c.Request().SetRequestURI("/dummy?" + pagination.PageQueryParam + "=2" + "&" + pagination.PerPageQueryParam + "=10")
		p = pagination.New(c)

		// Assert the last course in the pagination response
		result, err = GetCourses(db, &database.DatabaseParams{Pagination: p}, ctx)
		require.Nil(t, err)
		require.Len(t, result, 7)
		assert.Equal(t, courses[16].ID, result[6].ID)
		assert.Equal(t, courses[16].Title, result[6].Title)
		assert.Equal(t, courses[16].Path, result[6].Path)
	})

	t.Run("orderby", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)

		result, err := GetCourses(db, &database.DatabaseParams{OrderBy: []string{"created_at desc"}}, ctx)
		require.Nil(t, err)
		require.Len(t, result, 5)
		assert.Equal(t, courses[4].ID, result[0].ID)
		assert.Equal(t, courses[4].Title, result[0].Title)
		assert.Equal(t, courses[4].Path, result[0].Path)
	})

	t.Run("where", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)

		where := []database.Where{{Column: "course.id", Value: courses[2].ID}}

		result, err := GetCourses(db, &database.DatabaseParams{Where: where}, ctx)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, courses[2].ID, result[0].ID)

		where = []database.Where{{Query: "? = ?", Column: "course.id", Value: courses[3].ID}}

		result, err = GetCourses(db, &database.DatabaseParams{Where: where}, ctx)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, courses[3].ID, result[0].ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Course{}).Exec(ctx)
		require.Nil(t, err)

		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(c)

		c.Request().SetRequestURI("/dummy?" + pagination.PageQueryParam + "=1" + "&" + pagination.PerPageQueryParam + "=10")
		p := pagination.New(c)

		// With pagination
		_, err = GetCourses(db, &database.DatabaseParams{Pagination: p}, ctx)
		require.ErrorContains(t, err, "no such table: courses")

		// Without pagination
		_, err = GetCourses(db, nil, ctx)
		require.ErrorContains(t, err, "no such table: courses")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetCourseById(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course, err := GetCourseById(db, "1", nil, ctx)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, course)
	})

	t.Run("found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		NewTestScans(t, db, courses)

		result, err := GetCourseById(db, courses[2].ID, nil, ctx)
		require.Nil(t, err)
		assert.Equal(t, courses[2].ID, result.ID)
		assert.Equal(t, courses[2].Title, result.Title)
		assert.Equal(t, courses[2].Path, result.Path)
		assert.Empty(t, courses[2].CardPath)
		assert.False(t, courses[2].Started)
		assert.False(t, courses[2].Finished)
		assert.False(t, courses[2].CreatedAt.IsZero())
		assert.False(t, courses[2].UpdatedAt.IsZero())

		// Scan status
		require.NotEmpty(t, result.ScanStatus)
		assert.Equal(t, "waiting", result.ScanStatus)

		// Relations are empty
		assert.Nil(t, courses[2].Assets)
	})

	t.Run("relations", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		// ----------------------------
		// Assets relation
		// ----------------------------
		relation := []database.Relation{{Struct: "Assets"}}

		result, err := GetCourseById(db, courses[2].ID, &database.DatabaseParams{Relation: relation}, ctx)
		require.Nil(t, err)
		assert.Equal(t, courses[2].ID, result.ID)

		// Assert the assets
		require.NotNil(t, result.Assets)
		require.Len(t, result.Assets, 2)
		assert.Equal(t, assets[4].ID, result.Assets[0].ID)

		// Asset the attachments for the first asset is nil
		assert.Nil(t, result.Assets[0].Attachments)

		// ----------------------------
		// Assets and attachments relation
		// ----------------------------
		relation = []database.Relation{{Struct: "Assets"}, {Struct: "Assets.Attachments"}}

		result, err = GetCourseById(db, courses[3].ID, &database.DatabaseParams{Relation: relation}, ctx)
		require.Nil(t, err)
		assert.Equal(t, courses[3].ID, result.ID)

		// Assert the assets
		require.NotNil(t, result.Assets)
		require.Len(t, result.Assets, 2)
		assert.Equal(t, assets[6].ID, result.Assets[0].ID)

		// Asset the attachments
		assert.NotNil(t, result.Assets[0].Attachments)
		assert.Len(t, result.Assets[0].Attachments, 2)
		assert.Equal(t, attachments[12].ID, result.Assets[0].Attachments[0].ID)
	})

	// 	t.Run("preload", func(t *testing.T) {
	// 		_, db, ctx, teardown := setup(t)
	// 		defer teardown(t)

	// 		// Create 1 course with 5 assets with 2 attachments
	// 		course := NewTestCourses(t, db, 1)[0]
	// 		assets := CreateTestAssets(t, db, []*Course{course}, 5)
	// 		CreateTestAttachments(t, db, assets, 2)

	// 		preload := []database.Preload{
	// 			{Table: "Assets"},
	// 			{Table: "Assets.Attachments"},
	// 		}

	// 		result, err := GetCourseById(db, course.ID, &database.DatabaseParams{Preload: preload})
	// 		require.Nil(t, err)
	// 		assert.Equal(t, course.ID, result.ID)

	// 		// Assert the assets
	// 		require.Len(t, result.Assets, 5)
	// 		assert.Equal(t, assets[0].ID, result.Assets[0].ID)

	// 		// Assert the attachments for the first asset
	// 		require.Len(t, result.Assets[0].Attachments, 2)
	// 	})

	// 	t.Run("preload orderby", func(t *testing.T) {
	// 		_, db, ctx, teardown := setup(t)
	// 		defer teardown(t)

	// 		// Create 1 course with 5 assets
	// 		course := NewTestCourses(t, db, 1)[0]
	// 		assets := CreateTestAssets(t, db, []*Course{course}, 5)

	// 		preload := []database.Preload{
	// 			{Table: "Assets", OrderBy: "created_at desc"},
	// 			// {Table: "Assets.Attachments"},
	// 		}

	// 		result, err := GetCourseById(db, course.ID, &database.DatabaseParams{Preload: preload})
	// 		require.Nil(t, err)
	// 		assert.Equal(t, course.ID, result.ID)

	// 		// Assert the assets. The last created asset should now be the first result in the assets
	// 		// slice
	// 		require.Len(t, result.Assets, 5)
	// 		assert.Equal(t, assets[4].ID, result.Assets[0].ID)
	// 	})

	// 	t.Run("error preload orderby", func(t *testing.T) {
	// 		_, db, ctx, teardown := setup(t)
	// 		defer teardown(t)

	// 		// Create 1 course
	// 		course := NewTestCourses(t, db, 1)[0]

	// 		// No column
	// 		preload := []database.Preload{{Table: "Assets", OrderBy: "error_test desc"}}
	// 		result, err := GetCourseById(db, course.ID, &database.DatabaseParams{Preload: preload})
	// 		require.ErrorContains(t, err, "no such column: error_test")
	// 		assert.Nil(t, result)

	// 		// Invalid syntax
	// 		preload = []database.Preload{{Table: "Assets", OrderBy: "error_test invalid"}}
	// 		result, err = GetCourseById(db, course.ID, &database.DatabaseParams{Preload: preload})
	// 		require.ErrorContains(t, err, "near \"invalid\": syntax error")
	// 		assert.Nil(t, result)
	// 	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model((*Course)(nil)).Exec(ctx)
		require.Nil(t, err)

		_, err = GetCourseById(db, "1", nil, ctx)
		require.ErrorContains(t, err, "no such table: courses")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateCourse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := &Course{Title: "Course 1", Path: "/course1"}

		err := CreateCourse(db, course, ctx)
		require.Nil(t, err)
		assert.NotEmpty(t, course.ID)
		assert.False(t, course.Started)
		assert.False(t, course.Finished)
		assert.Empty(t, course.CardPath)
		assert.Empty(t, course.ScanStatus)
		assert.False(t, course.CreatedAt.IsZero())
		assert.False(t, course.UpdatedAt.IsZero())

		_, err = GetCourseById(db, course.ID, nil, ctx)
		assert.Nil(t, err)
	})

	t.Run("duplicate paths", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, nil, 1)[0]

		err := CreateCourse(db, course, ctx)
		require.Nil(t, err)
		assert.NotEmpty(t, course.ID)

		err = CreateCourse(db, course, ctx)
		assert.ErrorContains(t, err, "UNIQUE constraint failed: courses.path")
	})

	t.Run("constraint errors", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		// Path
		course := &Course{Title: "Course 1"}
		err := CreateCourse(db, course, ctx)
		assert.ErrorContains(t, err, "NOT NULL constraint failed: courses.path")

		// Title
		course = &Course{Path: "/course 1"}
		err = CreateCourse(db, course, ctx)
		assert.ErrorContains(t, err, "NOT NULL constraint failed: courses.title")

		// Success
		course = &Course{Title: "Course 1", Path: "/course 1"}
		err = CreateCourse(db, course, ctx)
		assert.Nil(t, err)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
func Test_UpdateCourseCardPath(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		require.Empty(t, course.CardPath)

		origCourse, err := GetCourseById(db, course.ID, nil, ctx)
		require.Nil(t, err)

		// Give time to allow `updated at` to be different
		time.Sleep(time.Millisecond * 1)

		// Update the card path
		err = UpdateCourseCardPath(db, course, "/path/to/card.jpg", ctx)
		require.Nil(t, err)

		// Get the updated course
		updatedCourse, err := GetCourseById(db, course.ID, nil, ctx)
		require.Nil(t, err)
		assert.Equal(t, "/path/to/card.jpg", updatedCourse.CardPath)
		assert.NotEqual(t, origCourse.UpdatedAt, updatedCourse.UpdatedAt)

		// Ensure the original course struct was updated
		assert.Equal(t, "/path/to/card.jpg", course.CardPath)
		assert.NotEqual(t, origCourse.UpdatedAt, course.UpdatedAt)
	})

	t.Run("same card path", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		require.Empty(t, course.CardPath)

		origCourse, err := GetCourseById(db, course.ID, nil, ctx)
		require.Nil(t, err)

		// Give time to allow `updated at` to be different
		time.Sleep(time.Millisecond * 1)

		err = UpdateCourseCardPath(db, course, "", ctx)
		require.Nil(t, err)

		// Assert there were no changes to the DB
		updatedCourse, err := GetCourseById(db, course.ID, nil, ctx)
		require.Nil(t, err)
		assert.Empty(t, updatedCourse.CardPath)
		assert.Equal(t, origCourse.UpdatedAt.String(), updatedCourse.UpdatedAt.String())

		// Assert there were no changes to the original struct
		assert.Empty(t, course.CardPath)
		assert.Equal(t, origCourse.UpdatedAt.String(), course.UpdatedAt.String())
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]

		course.ID = ""

		err := UpdateCourseCardPath(db, course, "123", ctx)
		assert.ErrorContains(t, err, "course ID cannot be empty")
	})

	t.Run("invalid id", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]

		origCourse, err := GetCourseById(db, course.ID, nil, ctx)
		require.Nil(t, err)

		// Change the ID
		course.ID = "invalid"

		// Update the scan status to waiting
		err = UpdateCourseCardPath(db, course, "1234", ctx)
		require.Nil(t, err)

		// Assert there were no changes to the DB
		updatedCourse, err := GetCourseById(db, origCourse.ID, nil, ctx)
		require.Nil(t, err)
		assert.Empty(t, updatedCourse.CardPath)
		assert.Equal(t, origCourse.UpdatedAt.String(), updatedCourse.UpdatedAt.String())

		// Assert there were no changes to the original struct
		assert.Empty(t, course.CardPath)
		assert.Equal(t, origCourse.UpdatedAt.String(), course.UpdatedAt.String())
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]

		_, err := db.DB().NewDropTable().Model(&Course{}).Exec(ctx)
		require.Nil(t, err)

		err = UpdateCourseCardPath(db, course, "123", ctx)
		require.ErrorContains(t, err, "no such table: courses")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteCourse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]

		count, err := DeleteCourse(db, course.ID, ctx)
		require.Nil(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		// Invalid ID
		count, err := DeleteCourse(db, "1", ctx)
		require.Nil(t, err)
		assert.Equal(t, 0, count)

		// Drop the table
		_, err = db.DB().NewDropTable().Model(&Course{}).Exec(ctx)
		require.Nil(t, err)

		count, err = DeleteCourse(db, "1", ctx)
		assert.ErrorContains(t, err, "no such table: courses")
		assert.Equal(t, 0, count)
	})
}
