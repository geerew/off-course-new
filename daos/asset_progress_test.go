package daos

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func AssetProgressSetup(t *testing.T) (*appFs.AppFs, *AssetProgressDao, database.Database) {
	appFs, db := setup(t)
	apDao := NewAssetProgressDao(db)
	return appFs, apDao, db
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssetProgress_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, db := AssetProgressSetup(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)

		ap := newTestAssetsProgress(t, nil, workingData[0].Assets[0].ID, workingData[0].ID)

		err := dao.Create(ap)
		require.Nil(t, err)
		assert.NotEmpty(t, ap.ID)
		assert.Equal(t, workingData[0].Assets[0].ID, ap.AssetID)
		assert.Zero(t, ap.VideoPos)
		assert.False(t, ap.Completed)
		assert.True(t, ap.CompletedAt.IsZero())
		assert.False(t, ap.CreatedAt.IsZero())
		assert.False(t, ap.UpdatedAt.IsZero())
	})

	t.Run("duplicate asset id", func(t *testing.T) {
		_, dao, db := AssetProgressSetup(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)
		ap := newTestAssetsProgress(t, db, workingData[0].Assets[0].ID, workingData[0].ID)

		// Create (again)
		err := dao.Create(ap)
		require.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.asset_id", dao.table))
	})

	t.Run("constraint errors", func(t *testing.T) {
		_, dao, db := AssetProgressSetup(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)

		// Asset ID
		ap := &models.AssetProgress{}
		require.ErrorContains(t, dao.Create(ap), fmt.Sprintf("NOT NULL constraint failed: %s.asset_id", TableAssetsProgress()))
		ap.AssetID = ""
		require.ErrorContains(t, dao.Create(ap), fmt.Sprintf("NOT NULL constraint failed: %s.asset_id", TableAssetsProgress()))
		ap.AssetID = "1234"

		// Asset ID
		require.ErrorContains(t, dao.Create(ap), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", TableAssetsProgress()))
		ap.CourseID = ""
		require.ErrorContains(t, dao.Create(ap), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", TableAssetsProgress()))
		ap.CourseID = "1234"

		// Invalid asset ID
		require.ErrorContains(t, dao.Create(ap), "FOREIGN KEY constraint failed")
		ap.AssetID = workingData[0].Assets[0].ID

		// Invalid course ID
		require.ErrorContains(t, dao.Create(ap), "FOREIGN KEY constraint failed")
		ap.CourseID = workingData[0].ID

		// Success
		require.Nil(t, dao.Create(ap))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssetProgress_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		_, dao, db := AssetProgressSetup(t)

		workingData := NewTestData(t, db, 3, true, 1, 0)
		for _, tc := range workingData {
			newTestAssetsProgress(t, db, tc.Assets[0].ID, tc.ID)
		}

		ap, err := dao.Get(workingData[1].Assets[0].ID)
		require.Nil(t, err)
		assert.Equal(t, workingData[1].Assets[0].ID, ap.AssetID)
	})

	t.Run("not found", func(t *testing.T) {
		_, dao, _ := AssetProgressSetup(t)

		ap, err := dao.Get("1234")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, ap)
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, _ := AssetProgressSetup(t)

		ap, err := dao.Get("")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, ap)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := AssetProgressSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		_, err = dao.Get("1234")
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssetProgress_Update(t *testing.T) {
	t.Run("update", func(t *testing.T) {
		_, dao, db := AssetProgressSetup(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)

		origAp := newTestAssetsProgress(t, db, workingData[0].Assets[0].ID, workingData[0].ID)
		require.Zero(t, origAp.VideoPos)

		cpDao := NewCourseProgressDao(db)

		// ----------------------------
		// Set to 50
		// ----------------------------
		origAp.VideoPos = 50
		require.Nil(t, dao.Update(origAp))

		updatedAp1, err := dao.Get(origAp.AssetID)
		require.Nil(t, err)
		require.Equal(t, 50, updatedAp1.VideoPos)

		// Ensure the course was set to started
		cp1, err := cpDao.Get(workingData[0].ID)
		require.Nil(t, err)
		assert.True(t, cp1.Started)
		assert.False(t, cp1.StartedAt.IsZero())

		// ----------------------------
		// Set to -10 (should be set to 0)
		// ----------------------------
		updatedAp1.VideoPos = -10
		require.Nil(t, dao.Update(updatedAp1))

		updatedAp2, err := dao.Get(updatedAp1.AssetID)
		require.Nil(t, err)
		require.Zero(t, updatedAp2.VideoPos)

		// Ensure the course is not started
		cp2, err := cpDao.Get(workingData[0].ID)
		require.Nil(t, err)
		assert.False(t, cp2.Started)
		assert.True(t, cp2.StartedAt.IsZero())

		// ----------------------------
		// Set completed
		// ----------------------------
		updatedAp2.Completed = true
		require.Nil(t, dao.Update(updatedAp2))

		updatedAp3, err := dao.Get(updatedAp2.AssetID)
		require.Nil(t, err)
		require.Zero(t, updatedAp3.VideoPos)
		require.True(t, updatedAp3.Completed)
		require.False(t, updatedAp3.CompletedAt.IsZero())

		// Ensure the course is started and completed
		cp3, err := cpDao.Get(workingData[0].ID)
		require.Nil(t, err)
		assert.True(t, cp3.Started)
		assert.False(t, cp3.StartedAt.IsZero())
		assert.Equal(t, 100, cp3.Percent)
		assert.False(t, cp3.CompletedAt.IsZero())
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, db := AssetProgressSetup(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)
		ap := newTestAssetsProgress(t, db, workingData[0].Assets[0].ID, workingData[0].ID)

		ap.AssetID = ""

		err := dao.Update(ap)
		assert.EqualError(t, err, "id cannot be empty")
	})

	t.Run("invalid asset id", func(t *testing.T) {
		_, dao, db := AssetProgressSetup(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)
		ap := newTestAssetsProgress(t, db, workingData[0].Assets[0].ID, workingData[0].ID)

		ap.AssetID = "1234"

		err := dao.Update(ap)
		require.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := AssetProgressSetup(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)
		ap := newTestAssetsProgress(t, db, workingData[0].Assets[0].ID, workingData[0].ID)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		err = dao.Update(ap)
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssetProgress_DeleteCascade(t *testing.T) {
	t.Run("course", func(t *testing.T) {
		_, dao, db := AssetProgressSetup(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)
		ap := newTestAssetsProgress(t, db, workingData[0].Assets[0].ID, workingData[0].ID)

		// Delete the course
		courseDao := NewCourseDao(db)
		err := courseDao.Delete(workingData[0].ID)
		require.Nil(t, err)

		// Check the asset progress was deleted
		_, err = dao.Get(ap.ID)
		require.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("asset", func(t *testing.T) {
		_, dao, db := AssetProgressSetup(t)

		workingData := NewTestData(t, db, 1, false, 1, 0)
		ap := newTestAssetsProgress(t, db, workingData[0].Assets[0].ID, workingData[0].ID)

		// Delete the asset
		assetDao := NewAssetDao(db)
		err := assetDao.Delete(workingData[0].Assets[0].ID)
		require.Nil(t, err)

		// Check the asset progress was deleted
		_, err = dao.Get(ap.ID)
		require.ErrorIs(t, err, sql.ErrNoRows)
	})
}
