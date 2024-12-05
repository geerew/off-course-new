package dao

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setup(tb testing.TB) (*DAO, context.Context) {
	tb.Helper()

	// Logger
	var logs []*logger.Log
	var logsMux sync.Mutex
	logger, _, err := logger.InitLogger(&logger.BatchOptions{
		BatchSize: 1,
		WriteFn:   logger.TestWriteFn(&logs, &logsMux),
	})
	require.NoError(tb, err, "Failed to initialize logger")

	// DB
	dbManager, err := database.NewSqliteDBManager(&database.DatabaseConfig{
		DataDir:  "./oc_data",
		AppFs:    appFs.NewAppFs(afero.NewMemMapFs(), logger),
		InMemory: true,
	})

	require.NoError(tb, err)
	require.NotNil(tb, dbManager)

	dao := &DAO{db: dbManager.DataDb}

	return dao, context.Background()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Count(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{}
		count, err := dao.Count(ctx, course, nil)
		require.NoError(t, err)
		require.Zero(t, count)
	})

	t.Run("entries", func(t *testing.T) {
		dao, ctx := setup(t)

		for i := range 5 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, dao.CreateCourse(ctx, course))
		}

		course := &models.Course{}
		count, err := dao.Count(ctx, course, nil)
		require.NoError(t, err)
		require.Equal(t, count, 5)
	})

	t.Run("where", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, dao.Create(ctx, course))
			courses = append(courses, course)
		}

		course := &models.Course{}

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		count, err := dao.Count(ctx, course, &database.Options{Where: squirrel.Eq{course.Table() + ".id": courses[1].ID}})
		require.NoError(t, err)
		require.Equal(t, 1, count)

		// ----------------------------
		// NOT EQUALS ID
		// ----------------------------
		count, err = dao.Count(ctx, course, &database.Options{Where: squirrel.NotEq{course.Table() + ".id": courses[1].ID}})
		require.NoError(t, err)
		require.Equal(t, 2, count)

		// ----------------------------
		// ERROR
		// ----------------------------
		count, err = dao.Count(ctx, course, &database.Options{Where: squirrel.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		require.Zero(t, count)
	})

	t.Run("invalid model", func(t *testing.T) {
		dao, ctx := setup(t)

		count, err := dao.Count(ctx, nil, nil)
		require.ErrorIs(t, err, utils.ErrNilPtr)
		require.Zero(t, count)
	})

	t.Run("db error", func(t *testing.T) {
		dao, ctx := setup(t)
		course := &models.Course{}

		_, err := dao.db.Exec("DROP TABLE IF EXISTS " + course.Table())
		require.NoError(t, err)

		_, err = dao.Count(ctx, course, nil)
		require.ErrorContains(t, err, "no such table: "+course.Table())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		dao, ctx := setup(t)

		// Create course
		course := &models.Course{Title: "Course 1", Path: "/course-1", Available: true, CardPath: "/course-1/card-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		// Get course
		courseResult := &models.Course{}
		require.NoError(t, dao.Get(ctx, courseResult, nil))
		require.Equal(t, course.ID, courseResult.ID)
		require.True(t, courseResult.CreatedAt.Equal(course.CreatedAt))
		require.True(t, courseResult.UpdatedAt.Equal(course.UpdatedAt))
		require.Equal(t, course.Title, courseResult.Title)
		require.Equal(t, course.Path, courseResult.Path)
		require.Equal(t, course.CardPath, courseResult.CardPath)
		require.True(t, courseResult.Available)
		require.Empty(t, courseResult.ScanStatus)
		require.NotEmpty(t, courseResult.Progress.ID)
		require.Equal(t, course.ID, courseResult.Progress.CourseID)
		require.False(t, courseResult.Progress.Started)
		require.Zero(t, courseResult.Progress.Percent)
		require.True(t, courseResult.Progress.StartedAt.IsZero())
		require.True(t, courseResult.Progress.CompletedAt.IsZero())

		// Create scan
		scan := &models.Scan{CourseID: course.ID}
		require.NoError(t, dao.CreateScan(ctx, scan))

		// Get scan
		scanResult := &models.Scan{}
		require.NoError(t, dao.Get(ctx, scanResult, nil))
		require.Equal(t, scan.ID, scanResult.ID)
		require.True(t, scanResult.CreatedAt.Equal(scan.CreatedAt))
		require.True(t, scanResult.UpdatedAt.Equal(scan.UpdatedAt))
		require.Equal(t, scan.CourseID, scanResult.CourseID)
		require.True(t, scanResult.Status.IsWaiting())
		require.Equal(t, course.Path, scanResult.CoursePath)

		// Get course (again)
		courseResult = &models.Course{}
		require.NoError(t, dao.Get(ctx, courseResult, nil))
		require.Equal(t, course.ID, courseResult.ID)
		require.True(t, courseResult.ScanStatus.IsWaiting())

		// Create asset
		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		// Get asset
		assetResult := &models.Asset{}
		require.NoError(t, dao.Get(ctx, assetResult, nil))
		require.Equal(t, asset.ID, assetResult.ID)
		require.True(t, assetResult.CreatedAt.Equal(asset.CreatedAt))
		require.True(t, assetResult.UpdatedAt.Equal(asset.UpdatedAt))
		require.Equal(t, asset.CourseID, assetResult.CourseID)
		require.Equal(t, asset.Title, assetResult.Title)
		require.Equal(t, asset.Prefix, assetResult.Prefix)
		require.Equal(t, asset.Chapter, assetResult.Chapter)
		require.Equal(t, asset.Type, assetResult.Type)
		require.Equal(t, asset.Path, assetResult.Path)
		require.Equal(t, asset.Hash, assetResult.Hash)
		require.Len(t, assetResult.Attachments, 0)

		// Create attachment
		attachment := &models.Attachment{
			AssetID: asset.ID,
			Title:   "Attachment 1",
			Path:    "/course-1/01 Attachment 1.txt",
		}
		require.NoError(t, dao.CreateAttachment(ctx, attachment))

		// Get attachment
		attachmentResult := &models.Attachment{}
		require.NoError(t, dao.Get(ctx, attachmentResult, nil))
		require.Equal(t, attachment.ID, attachmentResult.ID)
		require.True(t, attachmentResult.CreatedAt.Equal(attachment.CreatedAt))
		require.True(t, attachmentResult.UpdatedAt.Equal(attachment.UpdatedAt))
		require.Equal(t, attachment.AssetID, attachmentResult.AssetID)
		require.Equal(t, attachment.Title, attachmentResult.Title)
		require.Equal(t, attachment.Path, attachmentResult.Path)

		// Get asset (again)
		assetResult = &models.Asset{}
		require.NoError(t, dao.Get(ctx, assetResult, nil))
		require.Equal(t, asset.ID, assetResult.ID)
		require.Len(t, assetResult.Attachments, 1)
		require.Equal(t, attachment.Title, assetResult.Attachments[0].Title)

		// Create tag
		tag := &models.Tag{Tag: "Tag 1"}
		require.NoError(t, dao.CreateTag(ctx, tag))

		// Get tag
		tagResult := &models.Tag{}
		require.NoError(t, dao.Get(ctx, tagResult, nil))
		require.Equal(t, tag.ID, tagResult.ID)
		require.True(t, tagResult.CreatedAt.Equal(tag.CreatedAt))
		require.True(t, tagResult.UpdatedAt.Equal(tag.UpdatedAt))
		require.Equal(t, tag.Tag, tagResult.Tag)
		require.Len(t, tagResult.CourseTags, 0)

		// Create course tag
		courseTag := &models.CourseTag{TagID: tag.ID, CourseID: course.ID}
		require.NoError(t, dao.CreateCourseTag(ctx, courseTag))

		// Get course tag
		courseTagResult := &models.CourseTag{}
		require.NoError(t, dao.Get(ctx, courseTagResult, nil))
		require.Equal(t, courseTag.ID, courseTagResult.ID)
		require.True(t, courseTagResult.CreatedAt.Equal(courseTag.CreatedAt))
		require.True(t, courseTagResult.UpdatedAt.Equal(courseTag.UpdatedAt))
		require.Equal(t, courseTag.TagID, courseTagResult.TagID)
		require.Equal(t, courseTag.CourseID, courseTagResult.CourseID)
		require.Equal(t, course.Title, courseTagResult.Course)
		require.Equal(t, tag.Tag, courseTagResult.Tag)

		// Get tag (again)
		tagResult = &models.Tag{}
		require.NoError(t, dao.Get(ctx, tagResult, nil))
		require.Equal(t, tag.ID, tagResult.ID)
		require.Len(t, tagResult.CourseTags, 1)

		// Create user
		user := &models.User{Username: "user1", PasswordHash: "1234", Role: types.UserRoleAdmin}
		require.NoError(t, dao.CreateUser(ctx, user))

		// Get user
		userResult := &models.User{}
		require.NoError(t, dao.Get(ctx, userResult, nil))
		require.Equal(t, user.ID, userResult.ID)
		require.True(t, userResult.CreatedAt.Equal(user.CreatedAt))
		require.True(t, userResult.UpdatedAt.Equal(user.UpdatedAt))
		require.Equal(t, user.Username, userResult.Username)
		require.Equal(t, user.PasswordHash, userResult.PasswordHash)
		require.Equal(t, user.Role, userResult.Role)
	})

	t.Run("not found", func(t *testing.T) {
		dao, ctx := setup(t)
		course := &models.Course{}
		err := dao.Get(ctx, course, &database.Options{Where: squirrel.Eq{course.Table() + ".path": "1234"}})
		require.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("where", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, dao.Create(ctx, course))
			courses = append(courses, course)
			time.Sleep(1 * time.Millisecond)
		}

		courseResult := &models.Course{}
		require.NoError(t, dao.Get(ctx, courseResult, &database.Options{Where: squirrel.Eq{models.COURSE_TABLE + ".path": courses[1].Path}}))
		require.Equal(t, courses[1].ID, courseResult.ID)
	})

	t.Run("orderby", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, dao.Create(ctx, course))
			courses = append(courses, course)
			time.Sleep(1 * time.Millisecond)
		}

		result := &models.Course{}
		options := &database.Options{OrderBy: []string{fmt.Sprintf("%s.title DESC", models.COURSE_TABLE)}}
		require.NoError(t, dao.Get(ctx, result, options))
		require.Equal(t, courses[2].ID, result.ID)
	})

	t.Run("invalid model", func(t *testing.T) {
		dao, ctx := setup(t)
		err := dao.Get(ctx, nil, nil)
		require.ErrorIs(t, err, utils.ErrNilPtr)
	})

	t.Run("invalid where", func(t *testing.T) {
		dao, ctx := setup(t)
		err := dao.Get(ctx, &models.Course{}, &database.Options{Where: squirrel.Eq{"`": "`"}})
		require.ErrorContains(t, err, "SQL logic error: unrecognized token")
	})

	t.Run("db error", func(t *testing.T) {
		dao, ctx := setup(t)
		course := &models.Course{}

		_, err := dao.db.Exec("DROP TABLE IF EXISTS " + course.Table())
		require.NoError(t, err)

		err = dao.Get(ctx, course, nil)
		require.ErrorContains(t, err, "no such table: "+course.Table())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetById(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1", Available: true, CardPath: "/course-1/card-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		courseResult := &models.Course{Base: models.Base{ID: course.ID}}
		require.NoError(t, dao.GetById(ctx, courseResult))
		require.Equal(t, course.ID, courseResult.ID)
	})

	t.Run("not found", func(t *testing.T) {
		dao, ctx := setup(t)
		err := dao.GetById(ctx, &models.Course{Base: models.Base{ID: "1234"}})
		require.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("invalid model", func(t *testing.T) {
		dao, ctx := setup(t)
		err := dao.GetById(ctx, nil)
		require.ErrorIs(t, err, utils.ErrNilPtr)
	})

	t.Run("invalid id", func(t *testing.T) {
		dao, ctx := setup(t)
		err := dao.GetById(ctx, &models.Course{})
		require.ErrorIs(t, err, utils.ErrInvalidId)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_List(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		err := dao.List(ctx, &courses, nil)
		require.NoError(t, err)
		require.Empty(t, courses)
	})

	t.Run("entries", func(t *testing.T) {
		dao, ctx := setup(t)

		for i := range 5 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, dao.Create(ctx, course))
		}

		courses := []*models.Course{}
		err := dao.List(ctx, &courses, nil)
		require.NoError(t, err)
		require.Len(t, courses, 5)
	})

	t.Run("pagination", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 17 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, dao.Create(ctx, course))
			courses = append(courses, course)
			time.Sleep(1 * time.Millisecond)
		}

		coursesResult := []*models.Course{}

		// Page 1 (10 items)
		p := pagination.New(1, 10)
		require.NoError(t, dao.List(ctx, &coursesResult, &database.Options{Pagination: p}))
		require.Len(t, coursesResult, 10)
		require.Equal(t, 17, p.TotalItems())
		require.Equal(t, courses[0].ID, coursesResult[0].ID)
		require.Equal(t, courses[9].ID, coursesResult[9].ID)

		// Page 2 (7 items)
		p = pagination.New(2, 10)
		require.NoError(t, dao.List(ctx, &coursesResult, &database.Options{Pagination: p}))
		require.Len(t, coursesResult, 7)
		require.Equal(t, 17, p.TotalItems())
		require.Equal(t, courses[10].ID, coursesResult[0].ID)
		require.Equal(t, courses[16].ID, coursesResult[6].ID)
	})

	t.Run("orderby", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, dao.Create(ctx, course))
			courses = append(courses, course)
			time.Sleep(1 * time.Millisecond)
		}

		// CREATED_AT DESC
		coursesResult := []*models.Course{}
		options := &database.Options{OrderBy: []string{models.COURSE_TABLE + ".title DESC"}}
		require.NoError(t, dao.List(ctx, &coursesResult, options))
		require.Len(t, coursesResult, 3)
		require.Equal(t, courses[2].ID, coursesResult[0].ID)

		// 	// // ----------------------------
		// 	// // SCAN_STATUS DESC
		// 	// // ----------------------------

		// 	// // Create a scan for course 2 and 3
		// 	// scanDao := NewScanDao(db)

		// 	// testData[1].Scan = &models.Scan{CourseID: testData[1].ID}
		// 	// require.Nil(t, scanDao.Create(testData[1].Scan, nil))
		// 	// testData[2].Scan = &models.Scan{CourseID: testData[2].ID}
		// 	// require.Nil(t, scanDao.Create(testData[2].Scan, nil))

		// 	// // Set course 3 to processing
		// 	// testData[2].Scan.Status = types.NewScanStatus(types.ScanStatusProcessing)
		// 	// require.Nil(t, scanDao.Update(testData[2].Scan, nil))

		// 	// result, err = dao.List(&database.DatabaseParams{OrderBy: []string{dao.Table() + ".scan_status desc"}}, nil)
		// 	// require.Nil(t, err)
		// 	// require.Len(t, result, 3)

		// 	// require.Equal(t, testData[0].ID, result[2].ID)
		// 	// require.Equal(t, testData[1].ID, result[1].ID)
		// 	// require.Equal(t, testData[2].ID, result[0].ID)

		// 	// // ----------------------------
		// 	// // SCAN_STATUS ASC
		// 	// // ----------------------------
		// 	// result, err = dao.List(&database.DatabaseParams{OrderBy: []string{dao.Table() + ".scan_status asc"}}, nil)
		// 	// require.Nil(t, err)
		// 	// require.Len(t, result, 3)

		// 	// require.Equal(t, testData[0].ID, result[0].ID)
		// 	// require.Equal(t, testData[1].ID, result[1].ID)
		// 	// require.Equal(t, testData[2].ID, result[2].ID)

	})

	t.Run("where", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, dao.Create(ctx, course))
			courses = append(courses, course)
			time.Sleep(1 * time.Millisecond)
		}

		// Equals ID or ID
		coursesResult := []*models.Course{}
		options := &database.Options{
			Where: squirrel.Or{
				squirrel.Eq{models.COURSE_TABLE + ".id": courses[1].ID},
				squirrel.Eq{models.COURSE_TABLE + ".id": courses[2].ID},
			},
			OrderBy: []string{models.COURSE_TABLE + ".created_at ASC"},
		}
		require.NoError(t, dao.List(ctx, &coursesResult, options))
		require.Len(t, coursesResult, 2)
		require.Equal(t, courses[1].ID, coursesResult[0].ID)
		require.Equal(t, courses[2].ID, coursesResult[1].ID)
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		// Nil
		require.ErrorIs(t, dao.List(ctx, nil, nil), utils.ErrNilPtr)

		// Not a pointer
		require.ErrorIs(t, dao.List(ctx, []*models.Course{}, nil), utils.ErrNotPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ListPluck(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		dao, ctx := setup(t)

		ids, err := dao.ListPluck(ctx, &models.Course{}, nil, models.BASE_ID)
		require.NoError(t, err)
		require.Empty(t, ids)
	})

	t.Run("entries", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 5 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, dao.Create(ctx, course))
			courses = append(courses, course)
			time.Sleep(1 * time.Millisecond)
		}

		options := &database.Options{OrderBy: []string{models.COURSE_TABLE + ".created_at ASC"}}

		// Course IDs
		ids, err := dao.ListPluck(ctx, &models.Course{}, options, models.BASE_ID)
		require.NoError(t, err)
		require.Len(t, ids, 5)
		for i := range 5 {
			require.Equal(t, courses[i].ID, ids[i])
		}

		// Course paths
		paths, err := dao.ListPluck(ctx, &models.Course{}, options, models.COURSE_PATH)
		require.NoError(t, err)
		require.Len(t, paths, 5)
		for i := range 5 {
			require.Equal(t, courses[i].Path, paths[i])
		}
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.Create(ctx, course))

		require.NoError(t, dao.Delete(ctx, course, &database.Options{Where: squirrel.Eq{course.Table() + ".path": course.Path}}))
	})

	t.Run("nil model", func(t *testing.T) {
		dao, ctx := setup(t)
		err := dao.Delete(ctx, nil, &database.Options{Where: squirrel.Eq{"path": "1234"}})
		require.ErrorIs(t, err, utils.ErrNilPtr)
	})

	t.Run("nil where", func(t *testing.T) {
		dao, ctx := setup(t)
		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.Create(ctx, course))
		require.NoError(t, dao.Delete(ctx, course, nil))

		// Check if it was deleted
		courseResult := &models.Course{Base: models.Base{ID: course.ID}}
		err := dao.GetById(ctx, courseResult)
		require.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("db error", func(t *testing.T) {
		dao, ctx := setup(t)
		course := &models.Course{}

		_, err := dao.db.Exec("DROP TABLE IF EXISTS " + course.Table())
		require.NoError(t, err)

		err = dao.Delete(ctx, course, &database.Options{Where: squirrel.Eq{"path": "1234"}})
		require.ErrorContains(t, err, "no such table: "+course.Table())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Benchmark_GetById(b *testing.B) {
	dao, ctx := setup(b)

	for i := 0; i < 1000; i++ {
		course := &models.Course{}
		course.ID = fmt.Sprintf("%d", i)
		course.Title = fmt.Sprintf("Course %d", i)
		course.Path = fmt.Sprintf("/course-%d", i)
		require.NoError(b, dao.CreateCourse(ctx, course))

		courseProgress := &models.CourseProgress{}
		require.NoError(b, dao.Get(ctx, courseProgress, &database.Options{Where: squirrel.Eq{courseProgress.Table() + ".course_id": course.ID}}))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		courseResult := &models.Course{Base: models.Base{ID: fmt.Sprintf("%d", (i % 1000))}}
		require.NoError(b, dao.GetById(ctx, courseResult))
	}
}
