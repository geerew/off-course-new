package daos

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func attachmentSetup(t *testing.T) (*appFs.AppFs, *AttachmentDao, database.Database) {
	appFs, db := setup(t)
	attachmentDao := NewAttachmentDao(db)
	return appFs, attachmentDao, db
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachment_Count(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, dao, _ := attachmentSetup(t)

		count, err := dao.Count(nil)
		require.Nil(t, err)
		assert.Zero(t, count)
	})

	t.Run("entries", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		NewTestData(t, db, 5, false, 1, 1)

		count, err := dao.Count(nil)
		require.Nil(t, err)
		assert.Equal(t, count, 5)
	})

	t.Run("where", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		workingData := NewTestData(t, db, 3, false, 1, 2)

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		count, err := dao.Count(&database.DatabaseParams{Where: squirrel.Eq{TableAttachments() + ".id": workingData[1].Assets[0].Attachments[1].ID}})
		require.Nil(t, err)
		assert.Equal(t, 1, count)

		// ----------------------------
		// NOT EQUALS ID
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.NotEq{TableAttachments() + ".id": workingData[1].Assets[0].Attachments[1].ID}})
		require.Nil(t, err)
		assert.Equal(t, 5, count)

		// ----------------------------
		// EQUALS ASSET_ID
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.Eq{TableAttachments() + ".asset_id": workingData[1].Assets[0].ID}})
		require.Nil(t, err)
		assert.Equal(t, 2, count)

		// ----------------------------
		// ERROR
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		assert.Equal(t, 0, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		_, err = dao.Count(nil)
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachment_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		workingData := NewTestData(t, nil, 1, false, 1, 1)

		// Create the course
		courseDao := NewCourseDao(db)
		require.Nil(t, courseDao.Create(workingData[0].Course))

		// Create the asset
		assetDao := NewAssetDao(db)
		err := assetDao.Create(workingData[0].Assets[0])
		require.Nil(t, err)

		// Create the attachment
		err = dao.Create(workingData[0].Assets[0].Attachments[0])
		require.Nil(t, err)

		newA, err := dao.Get(workingData[0].Assets[0].Attachments[0].ID)
		require.Nil(t, err)
		assert.Equal(t, workingData[0].Assets[0].Attachments[0].ID, newA.ID)
		assert.Equal(t, workingData[0].Assets[0].Attachments[0].CourseID, newA.CourseID)
		assert.Equal(t, workingData[0].Assets[0].Attachments[0].AssetID, newA.AssetID)
		assert.Equal(t, workingData[0].Assets[0].Attachments[0].Title, newA.Title)
		assert.Equal(t, workingData[0].Assets[0].Attachments[0].Path, newA.Path)
		assert.False(t, newA.CreatedAt.IsZero())
		assert.False(t, newA.UpdatedAt.IsZero())
	})

	t.Run("duplicate paths", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		workingData := NewTestData(t, db, 1, false, 1, 1)

		// Create the attachment (again)
		err := dao.Create(workingData[0].Assets[0].Attachments[0])
		require.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.path", dao.table))
	})

	t.Run("constraints", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)

		// No course ID
		attachment := &models.Attachment{}
		assert.ErrorContains(t, dao.Create(attachment), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", TableAttachments()))
		attachment.CourseID = ""
		assert.ErrorContains(t, dao.Create(attachment), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", TableAttachments()))
		attachment.CourseID = "1234"

		// No asset ID
		assert.ErrorContains(t, dao.Create(attachment), fmt.Sprintf("NOT NULL constraint failed: %s.asset_id", TableAttachments()))
		attachment.AssetID = ""
		assert.ErrorContains(t, dao.Create(attachment), fmt.Sprintf("NOT NULL constraint failed: %s.asset_id", TableAttachments()))
		attachment.AssetID = "1234"

		// No title
		assert.ErrorContains(t, dao.Create(attachment), fmt.Sprintf("NOT NULL constraint failed: %s.title", TableAttachments()))
		attachment.Title = ""
		assert.ErrorContains(t, dao.Create(attachment), fmt.Sprintf("NOT NULL constraint failed: %s.title", TableAttachments()))
		attachment.Title = "Course 1"

		// No path
		assert.ErrorContains(t, dao.Create(attachment), "NOT NULL constraint failed: attachments.path")
		attachment.Path = ""
		assert.ErrorContains(t, dao.Create(attachment), "NOT NULL constraint failed: attachments.path")
		attachment.Path = "/course 1/01 attachment"

		// Invalid course ID
		assert.ErrorContains(t, dao.Create(attachment), "FOREIGN KEY constraint failed")
		attachment.CourseID = workingData[0].ID

		// Invalid asset ID
		assert.ErrorContains(t, dao.Create(attachment), "FOREIGN KEY constraint failed")
		attachment.AssetID = workingData[0].Assets[0].ID

		// Success
		assert.Nil(t, dao.Create(attachment))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachment_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		workingData := NewTestData(t, db, 2, false, 1, 1)

		a, err := dao.Get(workingData[1].Assets[0].Attachments[0].ID)
		require.Nil(t, err)
		assert.Equal(t, workingData[1].Assets[0].Attachments[0].ID, a.ID)
	})

	t.Run("not found", func(t *testing.T) {
		_, dao, _ := attachmentSetup(t)

		c, err := dao.Get("1234")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, c)
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, _ := attachmentSetup(t)

		c, err := dao.Get("")
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, c)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		_, err = dao.Get("1234")
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachment_List(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, dao, _ := attachmentSetup(t)

		assets, err := dao.List(nil)
		require.Nil(t, err)
		require.Zero(t, assets)
	})

	t.Run("found", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		NewTestData(t, db, 5, false, 1, 1)

		result, err := dao.List(nil)
		require.Nil(t, err)
		require.Len(t, result, 5)

	})

	t.Run("orderby", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		workingData := NewTestData(t, db, 3, false, 1, 1)

		// ----------------------------
		// CREATED_AT DESC
		// ----------------------------
		result, err := dao.List(&database.DatabaseParams{OrderBy: []string{"created_at desc"}})
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, workingData[2].Assets[0].Attachments[0].ID, result[0].ID)
		assert.Equal(t, workingData[1].Assets[0].Attachments[0].ID, result[1].ID)
		assert.Equal(t, workingData[0].Assets[0].Attachments[0].ID, result[2].ID)

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{"created_at asc"}})
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, workingData[0].Assets[0].Attachments[0].ID, result[0].ID)
		assert.Equal(t, workingData[1].Assets[0].Attachments[0].ID, result[1].ID)
		assert.Equal(t, workingData[2].Assets[0].Attachments[0].ID, result[2].ID)

		// ----------------------------
		// Error
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{"unit_test asc"}})
		require.ErrorContains(t, err, "no such column")
		assert.Nil(t, result)
	})

	t.Run("where", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		workingData := NewTestData(t, db, 3, false, 2, 2)

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		result, err := dao.List(&database.DatabaseParams{Where: squirrel.Eq{TableAttachments() + ".id": workingData[1].Assets[1].Attachments[0].ID}})
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, workingData[1].Assets[1].Attachments[0].ID, result[0].ID)

		// ----------------------------
		// EQUALS ID OR ID
		// ----------------------------
		dbParams := &database.DatabaseParams{
			Where: squirrel.Or{
				squirrel.Eq{TableAttachments() + ".id": workingData[1].Assets[1].Attachments[0].ID},
				squirrel.Eq{TableAttachments() + ".id": workingData[2].Assets[0].Attachments[1].ID},
			},
			OrderBy: []string{"created_at asc"},
		}

		result, err = dao.List(dbParams)
		require.Nil(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, workingData[1].Assets[1].Attachments[0].ID, result[0].ID)
		assert.Equal(t, workingData[2].Assets[0].Attachments[1].ID, result[1].ID)

		// ----------------------------
		// ERROR
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		assert.Nil(t, result)
	})

	t.Run("pagination", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		workingData := NewTestData(t, db, 1, false, 1, 17)

		// ----------------------------
		// Page 1 with 10 items
		// ----------------------------
		p := pagination.New(1, 10)

		result, err := dao.List(&database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, workingData[0].Assets[0].Attachments[0].ID, result[0].ID)
		assert.Equal(t, workingData[0].Assets[0].Attachments[9].ID, result[9].ID)

		// ----------------------------
		// Page 2 with 7 items
		// ----------------------------
		p = pagination.New(2, 10)

		result, err = dao.List(&database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 7)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, workingData[0].Assets[0].Attachments[10].ID, result[0].ID)
		assert.Equal(t, workingData[0].Assets[0].Attachments[16].ID, result[6].ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		_, err = dao.List(nil)
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachment_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		workingData := NewTestData(t, db, 1, false, 1, 1)
		err := dao.Delete(workingData[0].Assets[0].Attachments[0].ID)
		require.Nil(t, err)
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, _ := attachmentSetup(t)

		err := dao.Delete("")
		assert.ErrorContains(t, err, "id cannot be empty")
	})

	t.Run("invalid id", func(t *testing.T) {
		_, dao, _ := attachmentSetup(t)

		err := dao.Delete("1234")
		assert.Nil(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		err = dao.Delete("1234")
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachment_DeleteCascade(t *testing.T) {
	t.Run("delete course", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		workingData := NewTestData(t, db, 2, false, 1, 1)

		// Delete the course
		courseDao := NewCourseDao(db)
		err := courseDao.Delete(workingData[0].ID)
		require.Nil(t, err)

		// Check the asset was deleted
		s, err := dao.Get(workingData[0].Assets[0].Attachments[0].ID)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, s)
	})

	t.Run("delete asset", func(t *testing.T) {
		_, dao, db := scanSetup(t)

		workingData := NewTestData(t, db, 2, false, 1, 1)

		// Delete the asset
		assetDao := NewAssetDao(db)
		err := assetDao.Delete(workingData[0].Assets[0].ID)
		require.Nil(t, err)

		_, err = dao.Get(workingData[0].Assets[0].Attachments[0].ID)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}
