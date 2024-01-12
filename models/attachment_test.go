package models

import (
	"database/sql"
	"fmt"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CountAttachments(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		count, err := CountAttachments(db, nil)
		require.Nil(t, err)
		assert.Zero(t, count)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		NewTestData(t, db, 5, false, 1, 1)

		count, err := CountAttachments(db, nil)
		require.Nil(t, err)
		assert.Equal(t, count, 5)
	})

	t.Run("where", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 3, false, 1, 2)

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		count, err := CountAttachments(db, &database.DatabaseParams{Where: sq.Eq{TableAttachments() + ".id": workingData[1].Assets[0].Attachments[1].ID}})
		require.Nil(t, err)
		assert.Equal(t, 1, count)

		// ----------------------------
		// NOT EQUALS ID
		// ----------------------------
		count, err = CountAttachments(db, &database.DatabaseParams{Where: sq.NotEq{TableAttachments() + ".id": workingData[1].Assets[0].Attachments[1].ID}})
		require.Nil(t, err)
		assert.Equal(t, 5, count)

		// ----------------------------
		// EQUALS ASSET_ID
		// ----------------------------
		count, err = CountAttachments(db, &database.DatabaseParams{Where: sq.Eq{TableAttachments() + ".asset_id": workingData[1].Assets[0].ID}})
		require.Nil(t, err)
		assert.Equal(t, 2, count)

		// ----------------------------
		// ERROR
		// ----------------------------
		count, err = CountAttachments(db, &database.DatabaseParams{Where: sq.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		assert.Equal(t, -1, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableAttachments())
		require.Nil(t, err)

		_, err = CountAttachments(db, nil)
		require.ErrorContains(t, err, "no such table: "+TableAttachments())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetAttachments(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		assets, err := GetAttachments(db, nil)
		require.Nil(t, err)
		require.Zero(t, assets)
	})

	t.Run("found", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		NewTestData(t, db, 5, false, 1, 1)

		result, err := GetAttachments(db, nil)
		require.Nil(t, err)
		require.Len(t, result, 5)
	})

	t.Run("orderby", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 3, false, 1, 1)

		// ----------------------------
		// Descending
		// ----------------------------
		dbParams := &database.DatabaseParams{OrderBy: []string{"created_at desc"}}
		result, err := GetAttachments(db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, workingData[2].Assets[0].Attachments[0].ID, result[0].ID)

		// ----------------------------
		// Ascending
		// ----------------------------
		result, err = GetAttachments(db, &database.DatabaseParams{OrderBy: []string{"created_at asc"}})
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, workingData[0].Assets[0].Attachments[0].ID, result[0].ID)

		// ----------------------------
		// Error
		// ----------------------------
		dbParams = &database.DatabaseParams{OrderBy: []string{"unit_test asc"}}
		result, err = GetAttachments(db, dbParams)
		require.ErrorContains(t, err, "no such column")
		assert.Nil(t, result)
	})

	t.Run("where", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 3, false, 2, 2)

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		result, err := GetAttachments(db, &database.DatabaseParams{Where: sq.Eq{TableAttachments() + ".id": workingData[1].Assets[1].Attachments[0].ID}})
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, workingData[1].Assets[1].Attachments[0].ID, result[0].ID)

		// ----------------------------
		// EQUALS ID OR ID
		// ----------------------------
		dbParams := &database.DatabaseParams{
			Where: sq.Or{
				sq.Eq{TableAttachments() + ".id": workingData[1].Assets[1].Attachments[0].ID},
				sq.Eq{TableAttachments() + ".id": workingData[2].Assets[0].Attachments[1].ID},
			},
			OrderBy: []string{"created_at asc"},
		}

		result, err = GetAttachments(db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, workingData[1].Assets[1].Attachments[0].ID, result[0].ID)
		assert.Equal(t, workingData[2].Assets[0].Attachments[1].ID, result[1].ID)

		// ----------------------------
		// EQUALS ASSET_ID
		// ----------------------------
		dbParams = &database.DatabaseParams{
			Where:   sq.Eq{TableAttachments() + ".asset_id": workingData[1].Assets[1].ID},
			OrderBy: []string{"created_at asc"},
		}
		result, err = GetAttachments(db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, workingData[1].Assets[1].Attachments[0].ID, result[0].ID)
		assert.Equal(t, workingData[1].Assets[1].Attachments[1].ID, result[1].ID)

		// ----------------------------
		// ERROR
		// ----------------------------
		result, err = GetAttachments(db, &database.DatabaseParams{Where: sq.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		assert.Nil(t, result)
	})

	t.Run("pagination", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 1, 17)

		// ----------------------------
		// Page 1 with 10 items
		// ----------------------------
		p := pagination.New(1, 10)

		result, err := GetAttachments(db, &database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, workingData[0].Assets[0].Attachments[0].ID, result[0].ID)
		assert.Equal(t, workingData[0].Assets[0].Attachments[9].ID, result[9].ID)

		// ----------------------------
		// Page 2 with 7 items
		// ----------------------------
		p = pagination.New(2, 10)

		result, err = GetAttachments(db, &database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 7)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, workingData[0].Assets[0].Attachments[10].ID, result[0].ID)
		assert.Equal(t, workingData[0].Assets[0].Attachments[16].ID, result[6].ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableAttachments())
		require.Nil(t, err)

		_, err = GetAttachments(db, nil)
		require.ErrorContains(t, err, "no such table: "+TableAttachments())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetAttachment(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		a, err := GetAttachment(db, "1234")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, a)
	})

	t.Run("found", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 2, false, 1, 1)

		a, err := GetAttachment(db, workingData[1].Assets[0].Attachments[0].ID)
		require.Nil(t, err)
		assert.Equal(t, workingData[1].Assets[0].Attachments[0].ID, a.ID)

	})

	t.Run("empty id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		c, err := GetAttachment(db, "")
		require.EqualError(t, err, "id cannot be empty")
		assert.Nil(t, c)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableAttachments())
		require.Nil(t, err)

		_, err = GetAttachment(db, "1234")
		require.ErrorContains(t, err, "no such table: "+TableAttachments())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateAttachment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)
		attachment := newTestAttachments(t, nil, workingData[0].Assets[0], 1)[0]

		err := CreateAttachment(db, attachment)
		require.Nil(t, err)

		newA, err := GetAttachment(db, attachment.ID)
		require.Nil(t, err)
		assert.Equal(t, attachment.ID, newA.ID)
		assert.Equal(t, attachment.CourseID, newA.CourseID)
		assert.Equal(t, attachment.AssetID, newA.AssetID)
		assert.Equal(t, attachment.Title, newA.Title)
		assert.Equal(t, attachment.Path, newA.Path)
		assert.False(t, attachment.CreatedAt.IsZero())
		assert.False(t, attachment.UpdatedAt.IsZero())
	})

	t.Run("duplicate paths", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)
		attachment := newTestAttachments(t, nil, workingData[0].Assets[0], 1)[0]

		err := CreateAttachment(db, attachment)
		require.Nil(t, err)

		err = CreateAttachment(db, attachment)
		assert.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.path", TableAttachments()))
	})

	t.Run("constraint errors", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)

		// No course ID
		attachment := &Attachment{}
		assert.ErrorContains(t, CreateAttachment(db, attachment), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", TableAttachments()))
		attachment.CourseID = ""
		assert.ErrorContains(t, CreateAttachment(db, attachment), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", TableAttachments()))
		attachment.CourseID = "1234"

		// No asset ID
		assert.ErrorContains(t, CreateAttachment(db, attachment), fmt.Sprintf("NOT NULL constraint failed: %s.asset_id", TableAttachments()))
		attachment.AssetID = ""
		assert.ErrorContains(t, CreateAttachment(db, attachment), fmt.Sprintf("NOT NULL constraint failed: %s.asset_id", TableAttachments()))
		attachment.AssetID = "1234"

		// No title
		assert.ErrorContains(t, CreateAttachment(db, attachment), fmt.Sprintf("NOT NULL constraint failed: %s.title", TableAttachments()))
		attachment.Title = ""
		assert.ErrorContains(t, CreateAttachment(db, attachment), fmt.Sprintf("NOT NULL constraint failed: %s.title", TableAttachments()))
		attachment.Title = "Course 1"

		// No path
		assert.ErrorContains(t, CreateAttachment(db, attachment), "NOT NULL constraint failed: attachments.path")
		attachment.Path = ""
		assert.ErrorContains(t, CreateAttachment(db, attachment), "NOT NULL constraint failed: attachments.path")
		attachment.Path = "/course 1/01 attachment"

		// Invalid course ID
		assert.ErrorContains(t, CreateAttachment(db, attachment), "FOREIGN KEY constraint failed")
		attachment.CourseID = workingData[0].ID

		// Invalid asset ID
		assert.ErrorContains(t, CreateAttachment(db, attachment), "FOREIGN KEY constraint failed")
		attachment.AssetID = workingData[0].Assets[0].ID

		// Success
		assert.Nil(t, CreateAttachment(db, attachment))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteAttachment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 1, 1)

		err := DeleteAttachment(db, workingData[0].Assets[0].Attachments[0].ID)
		require.Nil(t, err)
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		err := DeleteAttachment(db, "")
		assert.ErrorContains(t, err, "id cannot be empty")
	})

	t.Run("invalid id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		err := DeleteAttachment(db, "1234")
		assert.Nil(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableAttachments())
		require.Nil(t, err)

		err = DeleteAttachment(db, "1234")
		require.ErrorContains(t, err, "no such table: "+TableAttachments())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteAttachmentCascade(t *testing.T) {
	t.Run("delete course", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 2, false, 2, 2)

		err := DeleteCourse(db, workingData[0].ID)
		require.Nil(t, err)

		count, err := CountAttachments(db, nil)
		require.Nil(t, err)
		assert.Equal(t, 4, count)

	})

	t.Run("delete asset", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 2, false, 2, 2)

		err := DeleteAsset(db, workingData[0].Assets[0].ID)
		require.Nil(t, err)

		_, err = GetAttachment(db, workingData[0].Assets[0].Attachments[0].ID)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		_, err = GetAttachment(db, workingData[0].Assets[0].Attachments[1].ID)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}
