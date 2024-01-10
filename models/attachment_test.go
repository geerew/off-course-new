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

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 1)
		NewTestAttachments(t, db, assets, 1)

		count, err := CountAttachments(db, nil)
		require.Nil(t, err)
		assert.Equal(t, count, 5)
	})

	t.Run("where", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 3)
		assets := NewTestAssets(t, db, courses, 1)
		attachments := NewTestAttachments(t, db, assets, 2)

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		count, err := CountAttachments(db, &database.DatabaseParams{Where: sq.Eq{TableAttachments() + ".id": attachments[2].ID}})
		require.Nil(t, err)
		assert.Equal(t, 1, count)

		// ----------------------------
		// NOT EQUALS ID
		// ----------------------------
		count, err = CountAttachments(db, &database.DatabaseParams{Where: sq.NotEq{TableAttachments() + ".id": attachments[2].ID}})
		require.Nil(t, err)
		assert.Equal(t, 5, count)

		// ----------------------------
		// EQUALS ASSET_ID
		// ----------------------------
		count, err = CountAttachments(db, &database.DatabaseParams{Where: sq.Eq{TableAttachments() + ".asset_id": assets[1].ID}})
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

		courses := NewTestCourses(t, db, 5)
		assets := NewTestAssets(t, db, courses, 1)
		NewTestAttachments(t, db, assets, 1)

		result, err := GetAttachments(db, nil)
		require.Nil(t, err)
		require.Len(t, result, 5)
	})

	t.Run("orderby", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 3)
		assets := NewTestAssets(t, db, courses, 1)
		attachments := NewTestAttachments(t, db, assets, 1)

		// ----------------------------
		// Descending
		// ----------------------------
		dbParams := &database.DatabaseParams{OrderBy: []string{"created_at desc"}}
		result, err := GetAttachments(db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, attachments[2].ID, result[0].ID)

		// ----------------------------
		// Ascending
		// ----------------------------
		result, err = GetAttachments(db, &database.DatabaseParams{OrderBy: []string{"created_at asc"}})
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, attachments[0].ID, result[0].ID)

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

		// Create 3 courses, 2 assets per course, 2 attachments per asset (12 attachments total)
		courses := NewTestCourses(t, db, 3)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		result, err := GetAttachments(db, &database.DatabaseParams{Where: sq.Eq{TableAttachments() + ".id": attachments[1].ID}})
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, attachments[1].ID, result[0].ID)

		// ----------------------------
		// EQUALS ID OR ID
		// ----------------------------
		dbParams := &database.DatabaseParams{
			Where:   sq.Or{sq.Eq{TableAttachments() + ".id": attachments[1].ID}, sq.Eq{TableAttachments() + ".id": attachments[2].ID}},
			OrderBy: []string{"created_at asc"},
		}
		result, err = GetAttachments(db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, attachments[1].ID, result[0].ID)
		assert.Equal(t, attachments[2].ID, result[1].ID)

		// ----------------------------
		// EQUALS ASSET_ID
		// ----------------------------
		dbParams = &database.DatabaseParams{
			Where:   sq.Eq{TableAttachments() + ".asset_id": assets[1].ID},
			OrderBy: []string{"created_at asc"},
		}
		result, err = GetAttachments(db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, attachments[2].ID, result[0].ID)
		assert.Equal(t, attachments[3].ID, result[1].ID)

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

		courses := NewTestCourses(t, db, 1)
		assets := NewTestAssets(t, db, courses, 1)
		attachments := NewTestAttachments(t, db, assets, 17)

		// ----------------------------
		// Page 1 with 10 items
		// ----------------------------
		p := pagination.New(1, 10)

		result, err := GetAttachments(db, &database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, attachments[0].ID, result[0].ID)
		assert.Equal(t, attachments[9].ID, result[9].ID)

		// ----------------------------
		// Page 2 with 7 items
		// ----------------------------
		p = pagination.New(2, 10)

		result, err = GetAttachments(db, &database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 7)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, attachments[10].ID, result[0].ID)
		assert.Equal(t, attachments[16].ID, result[6].ID)
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

		courses := NewTestCourses(t, db, 2)
		assets := NewTestAssets(t, db, courses, 1)
		attachments := NewTestAttachments(t, db, assets, 1)

		a, err := GetAttachment(db, attachments[1].ID)
		require.Nil(t, err)
		assert.Equal(t, attachments[1].ID, a.ID)

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

		courses := NewTestCourses(t, db, 1)
		assets := NewTestAssets(t, db, courses, 1)
		attachment := NewTestAttachments(t, nil, assets, 1)[0]

		err := CreateAttachment(db, attachment)
		require.Nil(t, err)

		newA, err := GetAttachment(db, attachment.ID)
		require.Nil(t, err)
		assert.Equal(t, attachment.ID, newA.ID)
		assert.Equal(t, attachment.AssetID, newA.AssetID)
		assert.Equal(t, attachment.Title, newA.Title)
		assert.Equal(t, attachment.Path, newA.Path)
		assert.False(t, attachment.CreatedAt.IsZero())
		assert.False(t, attachment.UpdatedAt.IsZero())
	})

	t.Run("duplicate paths", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 1)
		assets := NewTestAssets(t, db, courses, 1)
		attachment := NewTestAttachments(t, nil, assets, 1)[0]

		err := CreateAttachment(db, attachment)
		require.Nil(t, err)

		err = CreateAttachment(db, attachment)
		assert.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.path", TableAttachments()))
	})

	t.Run("constraint errors", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 1)
		asset := NewTestAssets(t, db, courses, 1)[0]

		// No asset ID
		attachment := &Attachment{}
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

		// Invalid attachment ID
		assert.ErrorContains(t, CreateAttachment(db, attachment), "FOREIGN KEY constraint failed")

		// Success
		attachment.AssetID = asset.ID
		assert.Nil(t, CreateAttachment(db, attachment))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteAttachment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		courses := NewTestCourses(t, db, 1)
		assets := NewTestAssets(t, db, courses, 1)
		attachments := NewTestAttachments(t, db, assets, 1)

		err := DeleteAttachment(db, attachments[0].ID)
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

		// Create 2 courses, 2 assets per course, 2 attachments per asset (8 attachments total)
		courses := NewTestCourses(t, db, 2)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		err := DeleteCourse(db, courses[0].ID)
		require.Nil(t, err)

		for _, a := range attachments[:4] {
			_, err := GetAttachment(db, a.ID)
			assert.ErrorIs(t, err, sql.ErrNoRows)
		}
	})

	t.Run("delete asset", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		// Create 2 courses, 2 assets per course, 2 attachments per asset (8 attachments total)
		courses := NewTestCourses(t, db, 2)
		assets := NewTestAssets(t, db, courses, 2)
		attachments := NewTestAttachments(t, db, assets, 2)

		err := DeleteAsset(db, assets[0].ID)
		require.Nil(t, err)

		_, err = GetAttachment(db, attachments[0].ID)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		_, err = GetAttachment(db, attachments[1].ID)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}
