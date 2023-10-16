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

func Test_CountScans(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		count, err := CountScans(ctx, db, nil)
		require.Nil(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		NewTestScans(t, db, []*Course{course})

		count, err := CountScans(ctx, db, nil)
		require.Nil(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("where", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 3)
		scans := NewTestScans(t, db, courses)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: scans[1].ID}}}

		count, err := CountScans(ctx, db, dbParams)
		require.Nil(t, err)
		assert.Equal(t, 1, count)

		dbParams = &database.DatabaseParams{Where: []database.Where{{Query: "? = ?", Column: "id", Value: scans[0].ID}}}
		count, err = CountScans(ctx, db, dbParams)
		require.Nil(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Scan{}).Exec(ctx)
		require.Nil(t, err)

		_, err = CountScans(ctx, db, nil)
		require.ErrorContains(t, err, "no such table: scans")
	})

	t.Run("error where", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "", Value: ""}}}
		count, err := CountScans(ctx, db, dbParams)
		require.ErrorContains(t, err, "syntax error")
		assert.Equal(t, 0, count)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetScans(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		scans, err := GetScans(ctx, db, nil)
		require.Nil(t, err)
		require.Len(t, scans, 0)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		scans := NewTestScans(t, db, courses)

		result, err := GetScans(ctx, db, nil)
		require.Nil(t, err)
		require.Len(t, result, 5)
		assert.Equal(t, scans[0].ID, result[0].ID)
		assert.Equal(t, scans[0].CourseID, result[0].CourseID)
		assert.True(t, result[0].Status.IsWaiting())
		assert.False(t, result[0].CreatedAt.IsZero())
		assert.False(t, result[0].UpdatedAt.IsZero())

		// Relations
		assert.Nil(t, result[0].Course)
	})

	t.Run("course relation", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		scans := NewTestScans(t, db, courses)

		relation := []database.Relation{{Struct: "Course"}}

		result, err := GetScans(ctx, db, &database.DatabaseParams{Relation: relation})
		require.Nil(t, err)
		require.Len(t, result, 5)
		assert.Equal(t, scans[0].ID, result[0].ID)
		assert.Equal(t, scans[0].CourseID, result[0].CourseID)

		// Assert the course
		require.NotNil(t, result[0].Course)
		assert.Equal(t, courses[0].Title, result[0].Course.Title)
	})

	t.Run("pagination", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 17)
		scans := NewTestScans(t, db, courses)

		// Pagination context
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(c)

		// Page 1 with 10 items
		c.Request().SetRequestURI("/dummy?" + pagination.PageQueryParam + "=1" + "&" + pagination.PerPageQueryParam + "=10")
		p := pagination.New(c)

		// Assert the last scan in the paginated response
		result, err := GetScans(ctx, db, &database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 10)
		assert.Equal(t, scans[9].ID, result[9].ID)
		assert.Equal(t, scans[9].CourseID, result[9].CourseID)
		assert.Equal(t, scans[9].Status, result[9].Status)

		// Page 2 with 10 items
		c.Request().SetRequestURI("/dummy?" + pagination.PageQueryParam + "=2" + "&" + pagination.PerPageQueryParam + "=10")
		p = pagination.New(c)

		// Assert the last scan in the paginated response
		result, err = GetScans(ctx, db, &database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 7)
		assert.Equal(t, scans[16].ID, result[6].ID)
		assert.Equal(t, scans[16].CourseID, result[6].CourseID)
		assert.Equal(t, scans[16].Status, result[6].Status)
	})

	t.Run("orderby", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		scans := NewTestScans(t, db, courses)

		result, err := GetScans(ctx, db, &database.DatabaseParams{OrderBy: []string{"created_at desc"}})
		require.Nil(t, err)
		require.Len(t, result, 5)
		assert.Equal(t, scans[4].ID, result[0].ID)
		assert.Equal(t, scans[4].CourseID, result[0].CourseID)
	})

	t.Run("where", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		scans := NewTestScans(t, db, courses)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: scans[2].ID}}}

		result, err := GetScans(ctx, db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, scans[2].ID, result[0].ID)

		dbParams = &database.DatabaseParams{Where: []database.Where{{Query: "? = ?", Column: "id", Value: scans[3].ID}}}

		result, err = GetScans(ctx, db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, scans[3].ID, result[0].ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Scan{}).Exec(ctx)
		require.Nil(t, err)

		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(c)

		c.Request().SetRequestURI("/dummy?" + pagination.PageQueryParam + "=1" + "&" + pagination.PerPageQueryParam + "=10")
		p := pagination.New(c)

		// With pagination
		_, err = GetScans(ctx, db, &database.DatabaseParams{Pagination: p})
		require.ErrorContains(t, err, "no such table: scans")

		// Without pagination
		_, err = GetScans(ctx, db, nil)
		require.ErrorContains(t, err, "no such table: scans")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetScan(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: "1"}}}

		scan, err := GetScan(ctx, db, dbParams)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, scan)
	})

	t.Run("found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		scans := NewTestScans(t, db, courses)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: scans[2].ID}}}

		result, err := GetScan(ctx, db, dbParams)
		require.Nil(t, err)
		assert.Equal(t, scans[2].ID, result.ID)
		assert.Equal(t, scans[2].CourseID, result.CourseID)
		assert.True(t, result.Status.IsWaiting())
		assert.False(t, result.CreatedAt.IsZero())
		assert.False(t, result.UpdatedAt.IsZero())

		// Relations
		assert.Nil(t, result.Course)
	})

	t.Run("course relation", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		scans := NewTestScans(t, db, courses)

		dbParams := &database.DatabaseParams{
			Where:    []database.Where{{Column: "id", Value: scans[2].ID}},
			Relation: []database.Relation{{Struct: "Course"}},
		}

		result, err := GetScan(ctx, db, dbParams)
		require.Nil(t, err)
		assert.Equal(t, scans[2].ID, result.ID)
		assert.Equal(t, scans[2].CourseID, result.CourseID)

		// Assert the course
		require.NotNil(t, result.Course)
		assert.Equal(t, courses[2].Title, result.Course.Title)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Scan{}).Exec(ctx)
		require.Nil(t, err)

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: "1"}}}

		_, err = GetScan(ctx, db, dbParams)
		require.ErrorContains(t, err, "no such table: scans")
	})

	t.Run("missing where clause", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Scan{}).Exec(ctx)
		require.Nil(t, err)

		_, err = GetScan(ctx, db, nil)
		require.ErrorContains(t, err, "where clause required")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetScanById(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		scan, err := GetScanById(ctx, db, nil, "1")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, scan)
	})

	t.Run("found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		scans := NewTestScans(t, db, courses)

		result, err := GetScanById(ctx, db, nil, scans[2].ID)
		require.Nil(t, err)
		assert.Equal(t, scans[2].ID, result.ID)
		assert.Equal(t, scans[2].CourseID, result.CourseID)
		assert.True(t, result.Status.IsWaiting())
		assert.False(t, result.CreatedAt.IsZero())
		assert.False(t, result.UpdatedAt.IsZero())

		// Relations
		assert.Nil(t, result.Course)
	})

	t.Run("course relation", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		scans := NewTestScans(t, db, courses)

		dbParams := &database.DatabaseParams{Relation: []database.Relation{{Struct: "Course"}}}

		result, err := GetScanById(ctx, db, dbParams, scans[2].ID)
		require.Nil(t, err)
		assert.Equal(t, scans[2].ID, result.ID)
		assert.Equal(t, scans[2].CourseID, result.CourseID)

		// Assert the course
		require.NotNil(t, result.Course)
		assert.Equal(t, courses[2].Title, result.Course.Title)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Scan{}).Exec(ctx)
		require.Nil(t, err)

		_, err = GetScanById(ctx, db, nil, "1")
		require.ErrorContains(t, err, "no such table: scans")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetScanByCourseId(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		scan, err := GetScanByCourseId(ctx, db, nil, "1")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, scan)
	})

	t.Run("found", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		scans := NewTestScans(t, db, courses)

		result, err := GetScanByCourseId(ctx, db, nil, courses[2].ID)
		require.Nil(t, err)
		assert.Equal(t, scans[2].ID, result.ID)
		assert.Equal(t, scans[2].CourseID, result.CourseID)
		assert.True(t, result.Status.IsWaiting())
		assert.False(t, result.CreatedAt.IsZero())
		assert.False(t, result.UpdatedAt.IsZero())

		// Relations
		assert.Nil(t, result.Course)
	})

	t.Run("course relation", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 5)
		scans := NewTestScans(t, db, courses)

		dbParams := &database.DatabaseParams{Relation: []database.Relation{{Struct: "Course"}}}

		result, err := GetScanByCourseId(ctx, db, dbParams, courses[2].ID)
		require.Nil(t, err)
		assert.Equal(t, scans[2].ID, result.ID)
		assert.Equal(t, scans[2].CourseID, result.CourseID)

		// Assert the course
		require.NotNil(t, result.Course)
		assert.Equal(t, courses[2].Title, result.Course.Title)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Scan{}).Exec(ctx)
		require.Nil(t, err)

		_, err = GetScanByCourseId(ctx, db, nil, "1")
		require.ErrorContains(t, err, "no such table: scans")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateScan(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		scan := NewTestScans(t, nil, []*Course{course})[0]

		err := CreateScan(ctx, db, scan)
		require.Nil(t, err)
		assert.NotEmpty(t, scan.ID)
		assert.Equal(t, course.ID, scan.CourseID)
		assert.True(t, scan.Status.IsWaiting())
		assert.False(t, scan.CreatedAt.IsZero())
		assert.False(t, scan.UpdatedAt.IsZero())

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: scan.ID}}}

		_, err = GetScan(ctx, db, dbParams)
		require.Nil(t, err)
	})

	t.Run("duplicate course", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		scan := NewTestScans(t, nil, []*Course{course})[0]

		err := CreateScan(ctx, db, scan)
		require.Nil(t, err)
		assert.NotEmpty(t, scan.ID)

		err = CreateScan(ctx, db, scan)
		assert.ErrorContains(t, err, "UNIQUE constraint failed: scans.course_id")
	})

	t.Run("constraint errors", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		// Missing course
		scan := &Scan{}
		err := CreateScan(ctx, db, scan)
		assert.ErrorContains(t, err, "FOREIGN KEY constraint failed")

		// An invalid courseID provided
		scan = &Scan{CourseID: "invalid"}
		err = CreateScan(ctx, db, scan)
		assert.ErrorContains(t, err, "FOREIGN KEY constraint failed")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateScanStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		scan := NewTestScans(t, db, []*Course{course})[0]

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: scan.ID}}}

		origScan, err := GetScan(ctx, db, dbParams)
		require.Nil(t, err)

		// Give time to allow `updated at` to be different
		time.Sleep(time.Millisecond * 1)

		// Update the scan status to processing
		err = UpdateScanStatus(ctx, db, scan, types.ScanStatusProcessing)
		require.Nil(t, err)

		// Get the updated scan and ensure the status was updated
		updatedScan, err := GetScan(ctx, db, dbParams)
		require.Nil(t, err)
		assert.False(t, updatedScan.Status.IsWaiting())
		assert.NotEqual(t, origScan.UpdatedAt.String(), updatedScan.UpdatedAt.String())

		// Ensure the original scan struct was updated
		assert.False(t, scan.Status.IsWaiting())
		assert.NotEqual(t, origScan.UpdatedAt.String(), scan.UpdatedAt.String())
	})

	t.Run("same status", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		scan := NewTestScans(t, db, []*Course{course})[0]

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: scan.ID}}}

		origScan, err := GetScan(ctx, db, dbParams)
		require.Nil(t, err)

		// Give time to allow `updated at` to be different
		time.Sleep(time.Millisecond * 1)

		err = UpdateScanStatus(ctx, db, scan, types.ScanStatusWaiting)
		require.Nil(t, err)

		// Assert there were no changes to the DB
		updatedScan, err := GetScan(ctx, db, dbParams)
		require.Nil(t, err)
		assert.True(t, updatedScan.Status.IsWaiting())
		assert.Equal(t, origScan.UpdatedAt.String(), updatedScan.UpdatedAt.String())

		// Assert there were no changes to the original struct
		assert.True(t, scan.Status.IsWaiting())
		assert.Equal(t, origScan.UpdatedAt.String(), scan.UpdatedAt.String())
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		scan := NewTestScans(t, db, []*Course{course})[0]

		scan.ID = ""

		err := UpdateScanStatus(ctx, db, scan, types.ScanStatusProcessing)
		assert.ErrorContains(t, err, "scan ID cannot be empty")
	})

	t.Run("invalid id", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		scan := NewTestScans(t, db, []*Course{course})[0]

		dbParams := &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: scan.ID}}}

		origScan, err := GetScan(ctx, db, dbParams)
		require.Nil(t, err)

		// Change the ID
		scan.ID = "invalid"

		err = UpdateScanStatus(ctx, db, scan, types.ScanStatusProcessing)
		require.Nil(t, err)

		// Assert there were no changes to the DB
		dbParams = &database.DatabaseParams{Where: []database.Where{{Column: "id", Value: origScan.ID}}}
		updatedScan, err := GetScan(ctx, db, dbParams)
		require.Nil(t, err)
		assert.True(t, updatedScan.Status.IsWaiting())
		assert.Equal(t, origScan.UpdatedAt.String(), updatedScan.UpdatedAt.String())

		// Assert there were no changes to the original struct
		assert.True(t, scan.Status.IsWaiting())
		assert.Equal(t, origScan.UpdatedAt.String(), scan.UpdatedAt.String())
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		scan := NewTestScans(t, db, []*Course{course})[0]

		_, err := db.DB().NewDropTable().Model(&Scan{}).Exec(ctx)
		require.Nil(t, err)

		err = UpdateScanStatus(ctx, db, scan, types.ScanStatusProcessing)
		require.ErrorContains(t, err, "no such table: scans")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteScan(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		scan := NewTestScans(t, db, []*Course{course})[0]

		count, err := DeleteScan(ctx, db, scan.ID)
		require.Nil(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("cascade", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		NewTestScans(t, db, []*Course{course})

		count, err := DeleteCourse(ctx, db, course.ID)
		require.Nil(t, err)
		require.Equal(t, 1, count)

		count, err = CountScans(ctx, db, nil)
		require.Nil(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		// Invalid ID
		count, err := DeleteScan(ctx, db, "1")
		require.Nil(t, err)
		assert.Equal(t, 0, count)

		// Drop the table
		_, err = db.DB().NewDropTable().Model(&Scan{}).Exec(ctx)
		require.Nil(t, err)

		count, err = DeleteScan(ctx, db, "1")
		assert.ErrorContains(t, err, "no such table: scans")
		assert.Equal(t, 0, count)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_NextScan(t *testing.T) {
	t.Run("first", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		course := NewTestCourses(t, db, 1)[0]
		scans := NewTestScans(t, db, []*Course{course})

		scan, err := NextScan(ctx, db)
		require.Nil(t, err)
		assert.Equal(t, scans[0].ID, scan.ID)
	})

	t.Run("next", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		// Create 3 scans
		courses := NewTestCourses(t, db, 3)
		scans := NewTestScans(t, db, courses)

		// Update the the first scan to processing
		err := UpdateScanStatus(ctx, db, scans[0], types.ScanStatusProcessing)
		require.Nil(t, err)

		// Get the next scan
		scan, err := NextScan(ctx, db)
		require.Nil(t, err)
		assert.Equal(t, scans[1].ID, scan.ID)
	})

	t.Run("empty", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		scan, err := NextScan(ctx, db)
		require.Nil(t, err)
		assert.Nil(t, scan)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, ctx, teardown := setup(t)
		defer teardown(t)

		_, err := db.DB().NewDropTable().Model(&Scan{}).Exec(ctx)
		require.Nil(t, err)

		_, err = NextScan(ctx, db)
		require.ErrorContains(t, err, "no such table: scans")
	})
}
