package dao

import (
	"database/sql"
	"testing"
	"time"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateAttachment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

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

		attachment := &models.Attachment{
			AssetID: asset.ID,
			Title:   "Attachment 1",
			Path:    "/course-1/01 attachment.txt",
		}
		require.NoError(t, dao.CreateAttachment(ctx, attachment))
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.CreateAttachment(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateAttachment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

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

		originalAttachment := &models.Attachment{
			AssetID: asset.ID,
			Title:   "Attachment 1",
			Path:    "/course-1/01 Attachment 1.txt",
		}
		require.NoError(t, dao.CreateAttachment(ctx, originalAttachment))

		time.Sleep(1 * time.Millisecond)

		newAttachment := &models.Attachment{
			Base:    originalAttachment.Base,
			AssetID: asset.ID,                        // Immutable
			Title:   "Attachment 2",                  // Mutable
			Path:    "/course-1/01 Attachment 2.txt", // Mutable
		}
		require.NoError(t, dao.UpdateAttachment(ctx, newAttachment))

		attachmentResult := &models.Attachment{Base: models.Base{ID: originalAttachment.ID}}
		require.NoError(t, dao.GetById(ctx, attachmentResult))
		require.Equal(t, newAttachment.ID, attachmentResult.ID)                          // No change
		require.Equal(t, newAttachment.AssetID, attachmentResult.AssetID)                // No change
		require.True(t, newAttachment.CreatedAt.Equal(originalAttachment.CreatedAt))     // No change
		require.Equal(t, newAttachment.Title, attachmentResult.Title)                    // Changed
		require.Equal(t, newAttachment.Path, attachmentResult.Path)                      // Changed
		require.False(t, attachmentResult.UpdatedAt.Equal(originalAttachment.UpdatedAt)) // Changed
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

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

		attachment := &models.Attachment{
			AssetID: asset.ID,
			Title:   "Attachment 1",
			Path:    "/course-1/01 attachment.txt",
		}
		require.NoError(t, dao.CreateAttachment(ctx, attachment))

		// Empty ID
		attachment.ID = ""
		require.ErrorIs(t, dao.UpdateAttachment(ctx, attachment), utils.ErrInvalidId)

		// Nil Model
		require.ErrorIs(t, dao.UpdateAttachment(ctx, nil), utils.ErrNilPtr)
	})
}

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestAttachment_List(t *testing.T) {
// 	t.Run("no entries", func(t *testing.T) {
// 		dao, _ := attachmentSetup(t)

// 		assets, err := dao.List(nil, nil)
// 		require.Nil(t, err)
// 		require.Zero(t, assets)
// 	})

// 	t.Run("found", func(t *testing.T) {
// 		dao, db := attachmentSetup(t)

// 		NewTestBuilder(t).Db(db).Courses(5).Assets(1).Attachments(1).Build()

// 		result, err := dao.List(nil, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 5)

// 	})

// 	t.Run("orderby", func(t *testing.T) {
// 		dao, db := attachmentSetup(t)

// 		testData := NewTestBuilder(t).Db(db).Courses(3).Assets(1).Attachments(1).Build()

// 		// ----------------------------
// 		// CREATED_AT DESC
// 		// ----------------------------
// 		result, err := dao.List(&database.DatabaseParams{OrderBy: []string{"created_at desc"}}, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 3)
// 		require.Equal(t, testData[2].Assets[0].Attachments[0].ID, result[0].ID)
// 		require.Equal(t, testData[1].Assets[0].Attachments[0].ID, result[1].ID)
// 		require.Equal(t, testData[0].Assets[0].Attachments[0].ID, result[2].ID)

// 		// ----------------------------
// 		// CREATED_AT ASC
// 		// ----------------------------
// 		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{"created_at asc"}}, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 3)
// 		require.Equal(t, testData[0].Assets[0].Attachments[0].ID, result[0].ID)
// 		require.Equal(t, testData[1].Assets[0].Attachments[0].ID, result[1].ID)
// 		require.Equal(t, testData[2].Assets[0].Attachments[0].ID, result[2].ID)
// 	})

// 	t.Run("where", func(t *testing.T) {
// 		dao, db := attachmentSetup(t)

// 		testData := NewTestBuilder(t).Db(db).Courses(3).Assets(2).Attachments(2).Build()

// 		// ----------------------------
// 		// EQUALS ID
// 		// ----------------------------
// 		result, err := dao.List(&database.DatabaseParams{Where: squirrel.Eq{dao.Table() + ".id": testData[1].Assets[1].Attachments[0].ID}}, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 1)
// 		require.Equal(t, testData[1].Assets[1].Attachments[0].ID, result[0].ID)

// 		// ----------------------------
// 		// EQUALS ID OR ID
// 		// ----------------------------
// 		dbParams := &database.DatabaseParams{
// 			Where: squirrel.Or{
// 				squirrel.Eq{dao.Table() + ".id": testData[1].Assets[1].Attachments[0].ID},
// 				squirrel.Eq{dao.Table() + ".id": testData[2].Assets[0].Attachments[1].ID},
// 			},
// 			OrderBy: []string{"created_at asc"},
// 		}

// 		result, err = dao.List(dbParams, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 2)
// 		require.Equal(t, testData[1].Assets[1].Attachments[0].ID, result[0].ID)
// 		require.Equal(t, testData[2].Assets[0].Attachments[1].ID, result[1].ID)

// 		// ----------------------------
// 		// ERROR
// 		// ----------------------------
// 		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Eq{"": ""}}, nil)
// 		require.ErrorContains(t, err, "syntax error")
// 		require.Nil(t, result)
// 	})

// 	t.Run("pagination", func(t *testing.T) {
// 		dao, db := attachmentSetup(t)

// 		testData := NewTestBuilder(t).Db(db).Courses(1).Assets(1).Attachments(17).Build()

// 		// ----------------------------
// 		// Page 1 with 10 items
// 		// ----------------------------
// 		p := pagination.New(1, 10)

// 		result, err := dao.List(&database.DatabaseParams{Pagination: p}, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 10)
// 		require.Equal(t, 17, p.TotalItems())
// 		require.Equal(t, testData[0].Assets[0].Attachments[0].ID, result[0].ID)
// 		require.Equal(t, testData[0].Assets[0].Attachments[9].ID, result[9].ID)

// 		// ----------------------------
// 		// Page 2 with 7 items
// 		// ----------------------------
// 		p = pagination.New(2, 10)

// 		result, err = dao.List(&database.DatabaseParams{Pagination: p}, nil)
// 		require.Nil(t, err)
// 		require.Len(t, result, 7)
// 		require.Equal(t, 17, p.TotalItems())
// 		require.Equal(t, testData[0].Assets[0].Attachments[10].ID, result[0].ID)
// 		require.Equal(t, testData[0].Assets[0].Attachments[16].ID, result[6].ID)
// 	})

// 	t.Run("db error", func(t *testing.T) {
// 		dao, db := attachmentSetup(t)

// 		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
// 		require.Nil(t, err)

// 		_, err = dao.List(nil, nil)
// 		require.ErrorContains(t, err, "no such table: "+dao.Table())
// 	})
// }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_AttachmentDeleteCascade(t *testing.T) {
	dao, ctx := setup(t)

	course := &models.Course{Title: "Course", Path: "/course"}
	require.NoError(t, dao.CreateCourse(ctx, course))

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

	attachment := &models.Attachment{
		AssetID: asset.ID,
		Title:   "Attachment 1",
		Path:    "/course-1/01 attachment.txt",
	}
	require.NoError(t, dao.CreateAttachment(ctx, attachment))

	require.Nil(t, dao.Delete(ctx, asset, nil))

	err := dao.GetById(ctx, attachment)
	require.ErrorIs(t, err, sql.ErrNoRows)
}
