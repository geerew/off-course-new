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

func Test_CountAssetsProgress(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		count, err := CountAssetsProgress(db, nil)
		require.Nil(t, err)
		assert.Zero(t, count)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 5, false, 1, 0)
		for _, tc := range workingData {
			newTestAssetsProgress(t, db, tc.Assets[0].ID)
		}

		count, err := CountAssetsProgress(db, nil)
		require.Nil(t, err)
		assert.Equal(t, count, 5)
	})

	t.Run("where", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 3, false, 1, 0)
		aps := []*AssetProgress{}
		for _, tc := range workingData {
			aps = append(aps, newTestAssetsProgress(t, db, tc.Assets[0].ID))
		}

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		count, err := CountAssetsProgress(db, &database.DatabaseParams{Where: sq.Eq{TableAssetsProgress() + ".id": aps[2].ID}})
		require.Nil(t, err)
		assert.Equal(t, 1, count)

		// ----------------------------
		// NOT EQUALS ID
		// ----------------------------
		count, err = CountAssetsProgress(db, &database.DatabaseParams{Where: sq.NotEq{TableAssetsProgress() + ".id": aps[2].ID}})
		require.Nil(t, err)
		assert.Equal(t, 2, count)

		// ----------------------------
		// ERROR
		// ----------------------------
		count, err = CountAssetsProgress(db, &database.DatabaseParams{Where: sq.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		assert.Equal(t, -1, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableAssetsProgress())
		require.Nil(t, err)

		_, err = CountAssetsProgress(db, nil)
		require.ErrorContains(t, err, "no such table: "+TableAssetsProgress())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetAssetsProgress(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		assets, err := GetAssetsProgress(db, nil)
		require.Nil(t, err)
		assert.Zero(t, assets)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 5, false, 1, 0)
		for _, tc := range workingData {
			newTestAssetsProgress(t, db, tc.Assets[0].ID)
		}

		result, err := GetAssetsProgress(db, nil)
		require.Nil(t, err)
		assert.Len(t, result, 5)
	})

	t.Run("orderby", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 3, false, 1, 0)
		aps := []*AssetProgress{}
		for _, tc := range workingData {
			aps = append(aps, newTestAssetsProgress(t, db, tc.Assets[0].ID))
		}

		// ----------------------------
		// Descending
		// ----------------------------
		dbParams := &database.DatabaseParams{OrderBy: []string{"created_at desc"}}
		result, err := GetAssetsProgress(db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, aps[2].ID, result[0].ID)

		// ----------------------------
		// Ascending
		// ----------------------------
		result, err = GetAssetsProgress(db, &database.DatabaseParams{OrderBy: []string{"created_at asc"}})
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, aps[0].ID, result[0].ID)

		// ----------------------------
		// Error
		// ----------------------------
		dbParams = &database.DatabaseParams{OrderBy: []string{"unit_test asc"}}
		result, err = GetAssetsProgress(db, dbParams)
		require.ErrorContains(t, err, "no such column")
		assert.Nil(t, result)
	})

	t.Run("where", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 3, false, 1, 0)
		aps := []*AssetProgress{}
		for _, tc := range workingData {
			aps = append(aps, newTestAssetsProgress(t, db, tc.Assets[0].ID))
		}

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		result, err := GetAssetsProgress(db, &database.DatabaseParams{Where: sq.Eq{TableAssetsProgress() + ".id": aps[2].ID}})
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, aps[2].ID, result[0].ID)

		// ----------------------------
		// EQUALS ID OR ID
		// ----------------------------
		dbParams := &database.DatabaseParams{
			Where:   sq.Or{sq.Eq{TableAssetsProgress() + ".id": aps[1].ID}, sq.Eq{TableAssetsProgress() + ".id": aps[2].ID}},
			OrderBy: []string{"created_at asc"},
		}
		result, err = GetAssetsProgress(db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, aps[1].ID, result[0].ID)
		assert.Equal(t, aps[2].ID, result[1].ID)

		// ----------------------------
		// ERROR
		// ----------------------------
		result, err = GetAssetsProgress(db, &database.DatabaseParams{Where: sq.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		assert.Nil(t, result)
	})

	t.Run("pagination", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 17, false, 1, 0)
		aps := []*AssetProgress{}
		for _, tc := range workingData {
			aps = append(aps, newTestAssetsProgress(t, db, tc.Assets[0].ID))
		}

		// ----------------------------
		// Page 1 with 10 items
		// ----------------------------
		p := pagination.New(1, 10)

		result, err := GetAssetsProgress(db, &database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, aps[0].ID, result[0].ID)
		assert.Equal(t, aps[9].ID, result[9].ID)

		// ----------------------------
		// Page 2 with 7 items
		// ----------------------------
		p = pagination.New(2, 10)

		result, err = GetAssetsProgress(db, &database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 7)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, aps[10].ID, result[0].ID)
		assert.Equal(t, aps[16].ID, result[6].ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableAssetsProgress())
		require.Nil(t, err)

		_, err = GetAssetsProgress(db, nil)
		require.ErrorContains(t, err, "no such table: "+TableAssetsProgress())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetAssetProgress(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		ap, err := GetAssetProgress(db, "1234")
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, ap)
	})

	t.Run("found", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 3, false, 1, 0)
		for _, tc := range workingData {
			newTestAssetsProgress(t, db, tc.Assets[0].ID)
		}

		result, err := GetAssetProgress(db, workingData[1].Assets[0].ID)
		require.Nil(t, err)
		assert.Equal(t, workingData[1].Assets[0].ID, result.AssetID)
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		result, err := GetAssetProgress(db, "")
		require.EqualError(t, err, "id cannot be empty")
		assert.Nil(t, result)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableAssetsProgress())
		require.Nil(t, err)

		_, err = GetAssetProgress(db, "1234")
		require.ErrorContains(t, err, "no such table: "+TableAssetsProgress())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateAssetProgress(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)
		ap := newTestAssetsProgress(t, nil, workingData[0].Assets[0].ID)

		err := CreateAssetProgress(db, ap)
		require.Nil(t, err)
		assert.NotEmpty(t, ap.ID)
		assert.Equal(t, workingData[0].Assets[0].ID, ap.AssetID)
		assert.Zero(t, ap.VideoPos)
		assert.False(t, ap.Completed)
		assert.True(t, ap.CompletedAt.IsZero())
		assert.False(t, ap.CreatedAt.IsZero())
		assert.False(t, ap.UpdatedAt.IsZero())
	})

	t.Run("duplicate course id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)
		ap := newTestAssetsProgress(t, nil, workingData[0].Assets[0].ID)

		err := CreateAssetProgress(db, ap)
		require.Nil(t, err)

		err = CreateAssetProgress(db, ap)
		require.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.asset_id", TableAssetsProgress()))
	})

	t.Run("constraint errors", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)

		// Asset ID
		ap := &AssetProgress{}
		require.ErrorContains(t, CreateAssetProgress(db, ap), fmt.Sprintf("NOT NULL constraint failed: %s.asset_id", TableAssetsProgress()))
		ap.AssetID = ""
		require.ErrorContains(t, CreateAssetProgress(db, ap), fmt.Sprintf("NOT NULL constraint failed: %s.asset_id", TableAssetsProgress()))
		ap.AssetID = "1234"

		// Invalid Course ID
		require.ErrorContains(t, CreateAssetProgress(db, ap), "FOREIGN KEY constraint failed")
		ap.AssetID = workingData[0].Assets[0].ID

		// Success
		require.Nil(t, CreateAssetProgress(db, ap))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateAssetProgressVideoPos(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)
		origAp := newTestAssetsProgress(t, db, workingData[0].Assets[0].ID)
		require.Zero(t, origAp.VideoPos)

		// Set to 50
		updatedAp, err := UpdateAssetProgressVideoPos(db, origAp.AssetID, 50)
		require.Nil(t, err)
		assert.Equal(t, 50, updatedAp.VideoPos)
		assert.NotEqual(t, origAp.UpdatedAt.String(), updatedAp.UpdatedAt.String())
	})

	t.Run("create new", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)

		ap1, err := UpdateAssetProgressVideoPos(db, workingData[0].Assets[0].ID, 50)
		require.Nil(t, err)
		assert.Equal(t, 50, ap1.VideoPos)
	})

	t.Run("normalize position", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)
		origAp := newTestAssetsProgress(t, db, workingData[0].Assets[0].ID)
		require.Zero(t, origAp.VideoPos)

		// Set to -1
		updatedAp, err := UpdateAssetProgressVideoPos(db, origAp.AssetID, -1)
		require.Nil(t, err)
		assert.Zero(t, updatedAp.VideoPos)
		assert.Equal(t, origAp.UpdatedAt.String(), updatedAp.UpdatedAt.String())
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		ap, err := UpdateAssetProgressVideoPos(db, "", 50)
		require.EqualError(t, err, "id cannot be empty")
		assert.Nil(t, ap)
	})

	t.Run("no change", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)
		origAp := newTestAssetsProgress(t, db, workingData[0].Assets[0].ID)
		require.Zero(t, origAp.VideoPos)

		updatedAp, err := UpdateAssetProgressVideoPos(db, origAp.AssetID, 0)
		require.Nil(t, err)
		assert.Zero(t, updatedAp.VideoPos)
		assert.Equal(t, origAp.UpdatedAt.String(), updatedAp.UpdatedAt.String())
	})

	t.Run("invalid course id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		ap, err := UpdateAssetProgressVideoPos(db, "1234", 50)
		assert.EqualError(t, err, "constraint failed: FOREIGN KEY constraint failed (787)")
		assert.Nil(t, ap)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableAssetsProgress())
		require.Nil(t, err)

		_, err = UpdateAssetProgressVideoPos(db, "1234", 50)
		require.ErrorContains(t, err, "no such table: "+TableAssetsProgress())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateAssetProgressCompleted(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)
		origAp := newTestAssetsProgress(t, db, workingData[0].Assets[0].ID)
		require.False(t, origAp.Completed)
		require.True(t, origAp.CompletedAt.IsZero())

		// ----------------------------
		// Set to true
		// ----------------------------
		updatedAp1, err := UpdateAssetProgressCompleted(db, origAp.AssetID, true)
		require.Nil(t, err)
		assert.True(t, updatedAp1.Completed)
		assert.False(t, updatedAp1.CompletedAt.IsZero())
		assert.NotEqual(t, origAp.UpdatedAt, updatedAp1.UpdatedAt)

		// ----------------------------
		// Set to false
		// ----------------------------
		updatedAp2, err := UpdateAssetProgressCompleted(db, origAp.AssetID, false)
		require.Nil(t, err)
		assert.False(t, updatedAp2.Completed)
		assert.True(t, updatedAp2.CompletedAt.IsZero())
		assert.NotEqual(t, updatedAp1.UpdatedAt, updatedAp2.UpdatedAt)
	})

	t.Run("create new", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)

		// ----------------------------
		// Set to true
		// ----------------------------
		ap1, err := UpdateAssetProgressCompleted(db, workingData[0].Assets[0].ID, true)
		require.Nil(t, err)
		assert.True(t, ap1.Completed)
		assert.False(t, ap1.CompletedAt.IsZero())

		// ----------------------------
		// Set to false
		// ----------------------------
		ap2, err := UpdateAssetProgressCompleted(db, workingData[0].Assets[0].ID, false)
		require.Nil(t, err)
		assert.False(t, ap2.Completed)
		assert.True(t, ap2.CompletedAt.IsZero())
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		ap, err := UpdateAssetProgressCompleted(db, "", true)
		require.EqualError(t, err, "id cannot be empty")
		assert.Nil(t, ap)
	})

	t.Run("no change", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)
		origAp := newTestAssetsProgress(t, db, workingData[0].Assets[0].ID)
		require.False(t, origAp.Completed)
		require.True(t, origAp.CompletedAt.IsZero())

		updatedAp, err := UpdateAssetProgressCompleted(db, origAp.AssetID, false)
		require.Nil(t, err)
		assert.False(t, updatedAp.Completed)
		assert.True(t, updatedAp.CompletedAt.IsZero())
		assert.Equal(t, origAp.UpdatedAt.String(), updatedAp.UpdatedAt.String())
	})

	t.Run("invalid course id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		ap, err := UpdateAssetProgressCompleted(db, "1234", true)
		require.EqualError(t, err, "constraint failed: FOREIGN KEY constraint failed (787)")
		assert.Nil(t, ap)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableAssetsProgress())
		require.Nil(t, err)

		_, err = UpdateAssetProgressCompleted(db, "1234", true)
		require.ErrorContains(t, err, "no such table: "+TableAssetsProgress())
	})
}
