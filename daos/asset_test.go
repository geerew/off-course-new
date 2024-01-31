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
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func assetSetup(t *testing.T) (*appFs.AppFs, *AssetDao, database.Database) {
	appFs, db := setup(t)
	assetDao := NewAssetDao(db)
	return appFs, assetDao, db
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_Count(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, dao, _ := assetSetup(t)

		count, err := dao.Count(nil)
		require.Nil(t, err)
		assert.Zero(t, count)
	})

	t.Run("entries", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		NewTestData(t, db, 5, false, 1, 0)

		count, err := dao.Count(nil)
		require.Nil(t, err)
		assert.Equal(t, count, 5)
	})

	t.Run("where", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		workingData := NewTestData(t, db, 3, false, 2, 0)

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		count, err := dao.Count(&database.DatabaseParams{Where: squirrel.Eq{TableAssets() + ".id": workingData[0].Assets[1].ID}})
		require.Nil(t, err)
		assert.Equal(t, 1, count)

		// ----------------------------
		// NOT EQUALS ID
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.NotEq{TableAssets() + ".id": workingData[0].Assets[1].ID}})
		require.Nil(t, err)
		assert.Equal(t, 5, count)

		// ----------------------------
		// EQUALS COURSE_ID
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.Eq{TableAssets() + ".course_id": workingData[1].ID}})
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
		_, dao, db := assetSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		_, err = dao.Count(nil)
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		workingData := NewTestData(t, nil, 1, false, 1, 0)

		// Create the course
		courseDao := NewCourseDao(db)
		require.Nil(t, courseDao.Create(workingData[0].Course))

		// Create the asset
		err := dao.Create(workingData[0].Assets[0])
		require.Nil(t, err)

		newA, err := dao.Get(workingData[0].Assets[0].ID, nil)
		require.Nil(t, err)
		assert.Equal(t, workingData[0].Assets[0].ID, newA.ID)
		assert.Equal(t, workingData[0].Assets[0].CourseID, newA.CourseID)
		assert.Equal(t, workingData[0].Assets[0].Title, newA.Title)
		assert.Equal(t, workingData[0].Assets[0].Prefix, newA.Prefix)
		assert.Equal(t, workingData[0].Assets[0].Chapter, newA.Chapter)
		assert.Equal(t, workingData[0].Assets[0].Type, newA.Type)
		assert.Equal(t, workingData[0].Assets[0].Path, newA.Path)
		assert.False(t, newA.CreatedAt.IsZero())
		assert.False(t, newA.UpdatedAt.IsZero())

		// Progress
		assert.Zero(t, newA.VideoPos)
		assert.False(t, newA.Completed)
		assert.True(t, newA.CompletedAt.IsZero())
	})

	t.Run("duplicate paths", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)

		// Create the asset (again)
		err := dao.Create(workingData[0].Assets[0])
		require.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.path", dao.table))
	})

	t.Run("constraints", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)

		// No course ID
		asset := &models.Asset{}
		assert.ErrorContains(t, dao.Create(asset), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", TableAssets()))
		asset.CourseID = ""
		assert.ErrorContains(t, dao.Create(asset), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", TableAssets()))
		asset.CourseID = "1234"

		// No title
		assert.ErrorContains(t, dao.Create(asset), fmt.Sprintf("NOT NULL constraint failed: %s.title", TableAssets()))
		asset.Title = ""
		assert.ErrorContains(t, dao.Create(asset), fmt.Sprintf("NOT NULL constraint failed: %s.title", TableAssets()))
		asset.Title = "Course 1"

		// No/invalid prefix
		assert.ErrorContains(t, dao.Create(asset), "NOT NULL constraint failed: assets.prefix")
		asset.Prefix = sql.NullInt16{Int16: -1, Valid: true}
		assert.ErrorContains(t, dao.Create(asset), "prefix must be greater than 0")
		asset.Prefix = sql.NullInt16{Int16: 1, Valid: true}

		// No type
		assert.ErrorContains(t, dao.Create(asset), "NOT NULL constraint failed: assets.type")
		asset.Type = types.Asset{}
		assert.ErrorContains(t, dao.Create(asset), "NOT NULL constraint failed: assets.type")
		asset.Type = *types.NewAsset("mp4")

		// No path
		assert.ErrorContains(t, dao.Create(asset), "NOT NULL constraint failed: assets.path")
		asset.Path = ""
		assert.ErrorContains(t, dao.Create(asset), "NOT NULL constraint failed: assets.path")
		asset.Path = "/course 1/01 asset"

		// Invalid Course ID
		assert.ErrorContains(t, dao.Create(asset), "FOREIGN KEY constraint failed")

		// Success
		asset.CourseID = workingData[0].ID
		assert.Nil(t, dao.Create(asset))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		workingData := NewTestData(t, db, 2, false, 1, 2)

		a, err := dao.Get(workingData[0].Assets[0].ID, nil)
		require.Nil(t, err)
		assert.Equal(t, workingData[0].Assets[0].ID, a.ID)

		// ----------------------------
		// Progress
		// ----------------------------
		apDao := NewAssetProgressDao(db)

		require.Zero(t, a.VideoPos)
		require.False(t, a.Completed)
		require.True(t, a.CompletedAt.IsZero())

		// Set video pos
		ap := &models.AssetProgress{
			AssetID:  workingData[0].Assets[0].ID,
			CourseID: workingData[0].ID,
			VideoPos: 50,
		}
		ap.RefreshId()

		err = apDao.Update(ap)
		require.Nil(t, err)

		a, err = dao.Get(a.ID, nil)
		require.Nil(t, err)
		assert.Equal(t, 50, a.VideoPos)
		assert.False(t, a.Completed)
		assert.True(t, a.CompletedAt.IsZero())

		// Set completed
		ap.Completed = true
		err = apDao.Update(ap)
		require.Nil(t, err)

		a, err = dao.Get(a.ID, nil)
		require.Nil(t, err)
		assert.Equal(t, 50, a.VideoPos)
		assert.True(t, a.Completed)
		assert.False(t, a.CompletedAt.IsZero())

		// ----------------------------
		// Attachments
		// ----------------------------
		require.Len(t, a.Attachments, 2)
		assert.Equal(t, workingData[0].Assets[0].Attachments[0].ID, a.Attachments[0].ID)
		assert.Equal(t, workingData[0].Assets[0].Attachments[1].ID, a.Attachments[1].ID)
	})

	t.Run("orderby", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		workingData := NewTestData(t, db, 3, false, 1, 2)

		// ----------------------------
		// ATTACHMENTS.CREATED_AT DESC
		// ----------------------------
		result, err := dao.Get(workingData[0].Assets[0].ID, &database.DatabaseParams{OrderBy: []string{TableAttachments() + ".created_at desc"}})
		require.Nil(t, err)
		require.Equal(t, workingData[0].Assets[0].ID, result.ID)
		assert.Equal(t, workingData[0].Assets[0].Attachments[0].ID, result.Attachments[1].ID)
		assert.Equal(t, workingData[0].Assets[0].Attachments[1].ID, result.Attachments[0].ID)

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		result, err = dao.Get(workingData[0].Assets[0].ID, &database.DatabaseParams{OrderBy: []string{TableAttachments() + ".created_at asc"}})
		require.Nil(t, err)
		require.Equal(t, workingData[0].Assets[0].ID, result.ID)
		assert.Equal(t, workingData[0].Assets[0].Attachments[0].ID, result.Attachments[0].ID)
		assert.Equal(t, workingData[0].Assets[0].Attachments[1].ID, result.Attachments[1].ID)

		// ----------------------------
		// Error
		// ----------------------------
		result, err = dao.Get(workingData[0].Assets[0].ID, &database.DatabaseParams{OrderBy: []string{"unit_test asc"}})
		require.ErrorContains(t, err, "no such column")
		assert.Nil(t, result)
	})

	t.Run("not found", func(t *testing.T) {
		_, dao, _ := assetSetup(t)

		c, err := dao.Get("1234", nil)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, c)
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, _ := assetSetup(t)

		c, err := dao.Get("", nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, c)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		_, err = dao.Get("1234", nil)
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_List(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, dao, _ := assetSetup(t)

		assets, err := dao.List(nil)
		require.Nil(t, err)
		require.Zero(t, assets)
	})

	t.Run("found", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		workingData := NewTestData(t, db, 5, false, 2, 3)

		result, err := dao.List(nil)
		require.Nil(t, err)
		require.Len(t, result, 10)

		// ----------------------------
		// Progress
		// ----------------------------
		apDao := NewAssetProgressDao(db)

		for _, a := range result {
			require.Zero(t, a.VideoPos)
			require.False(t, a.Completed)
			require.True(t, a.CompletedAt.IsZero())
		}

		// Update video position for the first asset (This will create the asset progress)
		ap1 := &models.AssetProgress{
			AssetID:  workingData[0].Assets[0].ID,
			CourseID: workingData[0].ID,
			VideoPos: 50,
		}
		ap1.RefreshId()
		require.Nil(t, apDao.Update(ap1))

		// Find all started videos
		dbParams := &database.DatabaseParams{
			Where: squirrel.And{
				squirrel.Eq{TableAssets() + ".type": string(types.AssetVideo)},
				squirrel.Gt{TableAssetsProgress() + ".video_pos": 0},
			},
		}
		result, err = dao.List(dbParams)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, workingData[0].Assets[0].ID, result[0].ID)
		assert.Equal(t, 50, result[0].VideoPos)

		// Mark the second asset as completed
		ap2 := &models.AssetProgress{
			AssetID:     workingData[1].Assets[1].ID,
			CourseID:    workingData[1].ID,
			Completed:   true,
			CompletedAt: types.NowDateTime(),
		}
		ap2.RefreshId()
		require.Nil(t, apDao.Update(ap2))

		// Find completed assets
		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Eq{TableAssetsProgress() + ".completed": true}})
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, workingData[1].Assets[1].ID, result[0].ID)
		assert.True(t, result[0].Completed)
		assert.False(t, result[0].CompletedAt.IsZero())

		// ----------------------------
		// Attachments
		// ----------------------------
		for _, a := range result {
			require.Len(t, a.Attachments, 3)
		}
	})

	t.Run("orderby", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		workingData := NewTestData(t, db, 3, false, 1, 2)

		// ----------------------------
		// CREATED_AT DESC
		// ----------------------------
		dbParams := &database.DatabaseParams{OrderBy: []string{"created_at desc"}}
		result, err := dao.List(dbParams)
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, workingData[2].Assets[0].ID, result[0].ID)

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{"created_at asc"}})
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, workingData[0].Assets[0].ID, result[0].ID)

		// ----------------------------
		// Error
		// ----------------------------
		dbParams = &database.DatabaseParams{OrderBy: []string{"unit_test asc"}}
		result, err = dao.List(dbParams)
		require.ErrorContains(t, err, "no such column")
		assert.Nil(t, result)
	})

	t.Run("where", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		workingData := NewTestData(t, db, 3, false, 2, 0)

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		result, err := dao.List(&database.DatabaseParams{Where: squirrel.Eq{TableAssets() + ".id": workingData[0].Assets[1].ID}})
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, workingData[0].Assets[1].ID, result[0].ID)

		// ----------------------------
		// EQUALS ID OR ID
		// ----------------------------
		dbParams := &database.DatabaseParams{
			Where:   squirrel.Or{squirrel.Eq{TableAssets() + ".id": workingData[0].Assets[1].ID}, squirrel.Eq{TableAssets() + ".id": workingData[1].Assets[1].ID}},
			OrderBy: []string{"created_at asc"},
		}
		result, err = dao.List(dbParams)
		require.Nil(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, workingData[0].Assets[1].ID, result[0].ID)
		assert.Equal(t, workingData[1].Assets[1].ID, result[1].ID)

		// ----------------------------
		// ERROR
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		assert.Nil(t, result)
	})

	t.Run("pagination", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		workingData := NewTestData(t, db, 1, false, 17, 0)

		// ----------------------------
		// Page 1 with 10 items
		// ----------------------------
		p := pagination.New(1, 10)

		result, err := dao.List(&database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, workingData[0].Assets[0].ID, result[0].ID)
		assert.Equal(t, workingData[0].Assets[9].ID, result[9].ID)

		// ----------------------------
		// Page 2 with 7 items
		// ----------------------------
		p = pagination.New(2, 10)

		result, err = dao.List(&database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 7)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, workingData[0].Assets[10].ID, result[0].ID)
		assert.Equal(t, workingData[0].Assets[16].ID, result[6].ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		_, err = dao.List(nil)
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)
		err := dao.Delete(workingData[0].Assets[0].ID)
		require.Nil(t, err)
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, _ := assetSetup(t)

		err := dao.Delete("")
		assert.ErrorContains(t, err, "id cannot be empty")
	})

	t.Run("invalid id", func(t *testing.T) {
		_, dao, _ := assetSetup(t)

		err := dao.Delete("1234")
		assert.Nil(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		err = dao.Delete("1234")
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_DeleteCascade(t *testing.T) {
	_, dao, db := scanSetup(t)

	workingData := NewTestData(t, db, 1, false, 1, 0)

	// Delete the course
	courseDao := NewCourseDao(db)
	err := courseDao.Delete(workingData[0].ID)
	require.Nil(t, err)

	// Check the asset was deleted
	s, err := dao.Get(workingData[0].ID)
	require.ErrorIs(t, err, sql.ErrNoRows)
	assert.Nil(t, s)
}
