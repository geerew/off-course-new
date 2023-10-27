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

		count, err := CountCourses(ctx, db, nil)
		require.Nil(t, err)
		assert.Equal(t, count, 0)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		NewTestCourses(t, db, 5)

		count, err := CountCourses(ctx, db, nil)
		require.Nil(t, err)
		assert.Equal(t, count, 5)
	})

	t.Run("where", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: courses[1].ID}}}

		count, err := CountCourses(ctx, db, dbParams)
		require.Nil(t, err)
		assert.Equal(t, 1, count)

		dbParams = &database.DatabaseParams{Where: []database.Where{{Query: "? = ?", Column: "id", Value: courses[0].ID}}}

		count, err = CountCourses(ctx, db, dbParams)
		require.Nil(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model((*Course)(nil)).Exec(ctx)
		require.Nil(t, err)

		_, err = CountCourses(ctx, db, nil)
		require.ErrorContains(t, err, "no such table: courses")
	})

	t.Run("error where", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "", Value: ""}}}
		count, err := CountCourses(ctx, db, dbParams)
		require.ErrorContains(t, err, "syntax error")
		assert.Equal(t, 0, count)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetCourses(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses, err := GetCourses(ctx, db, nil)
		require.Nil(t, err)
		require.Len(t, courses, 0)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		NewTestScans(t, db, courses)

		result, err := GetCourses(ctx, db, nil)
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

		result, err := GetCourses(ctx, db, &database.DatabaseParams{Relation: relation})
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

		result, err = GetCourses(ctx, db, &database.DatabaseParams{Relation: relation})
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

	t.Run("relations orderBy", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		// ----------------------------
		// Assets relation
		// ----------------------------
		dbParams := &database.DatabaseParams{
			Relation: []database.Relation{{Struct: "Assets", OrderBy: []string{"created_at desc"}}},
		}

		result, err := GetCourses(ctx, db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 5)
		assert.Equal(t, courses[0].ID, result[0].ID)

		// Assert the assets. The last asset created (for this course) should be the first item in
		// the assets slice
		require.NotNil(t, result[0].Assets)
		require.Len(t, result[0].Assets, 2)
		assert.Equal(t, assets[1].ID, result[0].Assets[0].ID)
		assert.Equal(t, assets[0].ID, result[0].Assets[1].ID)

		// Asset the attachments for the first asset is nil
		assert.Nil(t, result[0].Assets[0].Attachments)

		// ----------------------------
		// Assets and attachments relation
		// ----------------------------
		dbParams = &database.DatabaseParams{
			Relation: []database.Relation{{Struct: "Assets"}, {Struct: "Assets.Attachments", OrderBy: []string{"created_at desc"}}},
		}

		result, err = GetCourses(ctx, db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 5)
		assert.Equal(t, courses[0].ID, result[0].ID)

		// Assert the assets
		require.NotNil(t, result[0].Assets)
		require.Len(t, result[0].Assets, 2)
		assert.Equal(t, assets[0].ID, result[0].Assets[0].ID)

		// Asset the attachments. The last attachment created (for this course) should be the
		// first item in the result slice
		assert.NotNil(t, result[0].Assets[0].Attachments)
		assert.Len(t, result[0].Assets[0].Attachments, 2)
		assert.Equal(t, attachments[1].ID, result[0].Assets[0].Attachments[0].ID)
		assert.Equal(t, attachments[0].ID, result[0].Assets[0].Attachments[1].ID)
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
		result, err := GetCourses(ctx, db, &database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, courses[9].ID, result[9].ID)
		assert.Equal(t, courses[9].Title, result[9].Title)
		assert.Equal(t, courses[9].Path, result[9].Path)

		// Page 2 with 7 items
		c.Request().SetRequestURI("/dummy?" + pagination.PageQueryParam + "=2" + "&" + pagination.PerPageQueryParam + "=10")
		p = pagination.New(c)

		// Assert the last course in the pagination response
		result, err = GetCourses(ctx, db, &database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 7)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, courses[16].ID, result[6].ID)
		assert.Equal(t, courses[16].Title, result[6].Title)
		assert.Equal(t, courses[16].Path, result[6].Path)
	})

	t.Run("orderby", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)

		dbParams := &database.DatabaseParams{OrderBy: []string{"created_at desc"}}

		result, err := GetCourses(ctx, db, dbParams)
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

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: courses[2].ID}}}

		result, err := GetCourses(ctx, db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, courses[2].ID, result[0].ID)

		dbParams = &database.DatabaseParams{Where: []database.Where{{Query: "? = ?", Column: "id", Value: courses[3].ID}}}

		result, err = GetCourses(ctx, db, dbParams)
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
		_, err = GetCourses(ctx, db, &database.DatabaseParams{Pagination: p})
		require.ErrorContains(t, err, "no such table: courses")

		// Without pagination
		_, err = GetCourses(ctx, db, nil)
		require.ErrorContains(t, err, "no such table: courses")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetCourse(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: "1"}}}

		course, err := GetCourse(ctx, db, dbParams)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, course)
	})

	t.Run("found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		NewTestScans(t, db, courses)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: courses[2].ID}}}

		result, err := GetCourse(ctx, db, dbParams)
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
		dbParams := &database.DatabaseParams{
			Where:    []database.Where{{Column: "id", Value: courses[2].ID}},
			Relation: []database.Relation{{Struct: "Assets"}},
		}

		result, err := GetCourse(ctx, db, dbParams)
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
		dbParams = &database.DatabaseParams{
			Where: []database.Where{{Column: "id", Value: courses[3].ID}},
			Relation: []database.Relation{
				{Struct: "Assets", OrderBy: []string{"chapter asc", "prefix asc"}},
				{Struct: "Assets.Attachments"},
			},
		}

		result, err = GetCourse(ctx, db, dbParams)
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

	t.Run("relations orderBy", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		// ----------------------------
		// Assets relation
		// ----------------------------
		dbParams := &database.DatabaseParams{
			Where:    []database.Where{{Column: "id", Value: courses[2].ID}},
			Relation: []database.Relation{{Struct: "Assets", OrderBy: []string{"created_at desc"}}},
		}

		result, err := GetCourse(ctx, db, dbParams)
		require.Nil(t, err)
		assert.Equal(t, courses[2].ID, result.ID)

		// Assert the assets. The last asset created (for this course) should be the first item in
		// the assets slice
		require.NotNil(t, result.Assets)
		require.Len(t, result.Assets, 2)
		assert.Equal(t, assets[5].ID, result.Assets[0].ID)
		assert.Equal(t, assets[4].ID, result.Assets[1].ID)

		// Asset the attachments for the first asset is nil
		assert.Nil(t, result.Assets[0].Attachments)

		// ----------------------------
		// Assets and attachments relation
		// ----------------------------
		dbParams = &database.DatabaseParams{
			Where:    []database.Where{{Column: "id", Value: courses[3].ID}},
			Relation: []database.Relation{{Struct: "Assets"}, {Struct: "Assets.Attachments", OrderBy: []string{"created_at desc"}}},
		}

		result, err = GetCourse(ctx, db, dbParams)
		require.Nil(t, err)
		assert.Equal(t, courses[3].ID, result.ID)

		// Assert the assets
		require.NotNil(t, result.Assets)
		require.Len(t, result.Assets, 2)
		assert.Equal(t, assets[6].ID, result.Assets[0].ID)

		// Asset the attachments. The last attachment created (for this course) should be the
		// first item in the result slice
		assert.NotNil(t, result.Assets[0].Attachments)
		assert.Len(t, result.Assets[0].Attachments, 2)
		assert.Equal(t, attachments[13].ID, result.Assets[0].Attachments[0].ID)
		assert.Equal(t, attachments[12].ID, result.Assets[0].Attachments[1].ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model((*Course)(nil)).Exec(ctx)
		require.Nil(t, err)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: "1"}}}
		_, err = GetCourse(ctx, db, dbParams)
		require.ErrorContains(t, err, "no such table: courses")
	})

	t.Run("missing where clause", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Course{}).Exec(ctx)
		require.Nil(t, err)

		_, err = GetCourse(ctx, db, nil)
		require.ErrorContains(t, err, "where clause required")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetCourseById(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course, err := GetCourseById(ctx, db, nil, "1")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, course)
	})

	t.Run("found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		NewTestScans(t, db, courses)

		result, err := GetCourseById(ctx, db, nil, courses[2].ID)
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
		dbParams := &database.DatabaseParams{Relation: []database.Relation{{Struct: "Assets"}}}

		result, err := GetCourseById(ctx, db, dbParams, courses[2].ID)
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
		dbParams = &database.DatabaseParams{Relation: []database.Relation{{Struct: "Assets"}, {Struct: "Assets.Attachments"}}}

		result, err = GetCourseById(ctx, db, dbParams, courses[3].ID)
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

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model((*Course)(nil)).Exec(ctx)
		require.Nil(t, err)

		_, err = GetCourseById(ctx, db, nil, "1")
		require.ErrorContains(t, err, "no such table: courses")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateCourse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := &Course{Title: "Course 1", Path: "/course1"}

		err := CreateCourse(ctx, db, course)
		require.Nil(t, err)
		assert.NotEmpty(t, course.ID)
		assert.False(t, course.Started)
		assert.False(t, course.Finished)
		assert.Empty(t, course.CardPath)
		assert.Empty(t, course.ScanStatus)
		assert.False(t, course.CreatedAt.IsZero())
		assert.False(t, course.UpdatedAt.IsZero())

		count, err := CountCourses(ctx, db, nil)
		require.Nil(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("duplicate paths", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, nil, 1)[0]

		err := CreateCourse(ctx, db, course)
		require.Nil(t, err)
		assert.NotEmpty(t, course.ID)

		err = CreateCourse(ctx, db, course)
		assert.ErrorContains(t, err, "UNIQUE constraint failed: courses.path")
	})

	t.Run("constraint errors", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		// Path
		course := &Course{Title: "Course 1"}
		err := CreateCourse(ctx, db, course)
		assert.ErrorContains(t, err, "NOT NULL constraint failed: courses.path")

		// Title
		course = &Course{Path: "/course 1"}
		err = CreateCourse(ctx, db, course)
		assert.ErrorContains(t, err, "NOT NULL constraint failed: courses.title")

		// Success
		course = &Course{Title: "Course 1", Path: "/course 1"}
		err = CreateCourse(ctx, db, course)
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

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: course.ID}}}
		origCourse, err := GetCourse(ctx, db, dbParams)
		require.Nil(t, err)

		// Give time to allow `updated at` to be different
		time.Sleep(time.Millisecond * 1)

		// Update the card path
		err = UpdateCourseCardPath(ctx, db, course, "/path/to/card.jpg")
		require.Nil(t, err)

		// Get the updated course
		dbParams = &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: course.ID}}}
		updatedCourse, err := GetCourse(ctx, db, dbParams)
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

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: course.ID}}}

		origCourse, err := GetCourse(ctx, db, dbParams)
		require.Nil(t, err)

		// Give time to allow `updated at` to be different
		time.Sleep(time.Millisecond * 1)

		err = UpdateCourseCardPath(ctx, db, course, "")
		require.Nil(t, err)

		// Assert there were no changes to the DB

		updatedCourse, err := GetCourse(ctx, db, dbParams)
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

		err := UpdateCourseCardPath(ctx, db, course, "123")
		assert.ErrorContains(t, err, "course ID cannot be empty")
	})

	t.Run("invalid id", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: course.ID}}}

		origCourse, err := GetCourse(ctx, db, dbParams)
		require.Nil(t, err)

		// Change the ID
		course.ID = "invalid"

		// Update the scan status to waiting
		err = UpdateCourseCardPath(ctx, db, course, "1234")
		require.Nil(t, err)

		// Assert there were no changes to the DB
		dbParams = &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: origCourse.ID}}}

		updatedCourse, err := GetCourse(ctx, db, dbParams)
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

		err = UpdateCourseCardPath(ctx, db, course, "123")
		require.ErrorContains(t, err, "no such table: courses")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateCourse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: course.ID}}}

		// Store the original course
		origCourse, err := GetCourse(ctx, db, dbParams)
		require.Nil(t, err)
		require.False(t, origCourse.Started)
		require.False(t, origCourse.Finished)

		// Give time to allow `updated at` to be different
		time.Sleep(time.Millisecond * 1)

		// Update started, finished and card path. Path and title should not change
		course.Started = true
		course.Finished = true
		course.CardPath = "/new/card/path"
		course.Path = "/new/path"
		course.Title = "New title"

		err = UpdateCourse(ctx, db, course)
		require.Nil(t, err)

		// Assert there were changes to the DB
		updatedCourse, err := GetCourse(ctx, db, dbParams)

		require.Nil(t, err)

		// Assert started, finished, card path and updated at changed
		assert.True(t, updatedCourse.Started)
		assert.True(t, updatedCourse.Finished)
		assert.NotEqual(t, origCourse.CardPath, updatedCourse.CardPath)
		assert.NotEqual(t, origCourse.UpdatedAt, updatedCourse.UpdatedAt)

		// Assert the title and path did not change
		assert.Equal(t, origCourse.Title, updatedCourse.Title)
		assert.Equal(t, origCourse.Path, updatedCourse.Path)
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		course.ID = ""

		err := UpdateCourse(ctx, db, course)
		assert.ErrorContains(t, err, "course ID cannot be empty")
	})

	t.Run("invalid id", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]

		// Store the original course
		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: course.ID}}}
		origCourse, err := GetCourse(ctx, db, dbParams)
		require.Nil(t, err)

		// Change the ID
		course.ID = "invalid"

		err = UpdateCourse(ctx, db, course)
		require.Nil(t, err)

		// Assert there were no changes
		dbParams = &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: origCourse.ID}}}
		updatedCourse, err := GetCourse(ctx, db, dbParams)
		require.Nil(t, err)
		assert.Equal(t, origCourse.UpdatedAt.String(), updatedCourse.UpdatedAt.String())
		assert.Equal(t, origCourse.UpdatedAt.String(), course.UpdatedAt.String())
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]

		_, err := db.DB().NewDropTable().Model(&Course{}).Exec(ctx)
		require.Nil(t, err)

		err = UpdateCourse(ctx, db, course)
		require.ErrorContains(t, err, "no such table: courses")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteCourse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]

		count, err := DeleteCourse(ctx, db, course.ID)
		require.Nil(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		// Invalid ID
		count, err := DeleteCourse(ctx, db, "1")
		require.Nil(t, err)
		assert.Equal(t, 0, count)

		// Drop the table
		_, err = db.DB().NewDropTable().Model(&Course{}).Exec(ctx)
		require.Nil(t, err)

		count, err = DeleteCourse(ctx, db, "1")
		assert.ErrorContains(t, err, "no such table: courses")
		assert.Equal(t, 0, count)
	})
}
