package daos

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func CourseProgressSetup(t *testing.T) (*CourseProgressDao, database.Database) {
	t.Helper()

	dbManager := setup(t)
	cpDao := NewCourseProgressDao(dbManager.DataDb)
	return cpDao, dbManager.DataDb
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseProgress_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, db := CourseProgressSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Build()

		cp, err := dao.Get(testData[0].ID, nil)
		require.Nil(t, err)
		require.False(t, cp.Started)
		require.True(t, cp.StartedAt.IsZero())
		require.Zero(t, cp.Percent)
		require.True(t, cp.CompletedAt.IsZero())
		require.False(t, cp.CreatedAt.IsZero())
		require.False(t, cp.UpdatedAt.IsZero())
	})

	t.Run("duplicate course id", func(t *testing.T) {
		dao, db := CourseProgressSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Build()

		cp, err := dao.Get(testData[0].ID, nil)
		require.Nil(t, err)

		err = dao.Create(cp, nil)
		require.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.course_id", dao.Table()))
	})

	t.Run("constraint errors", func(t *testing.T) {
		dao, db := CourseProgressSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Build()

		// Delete the courses_progress row using squirrel
		query, args, _ := squirrel.StatementBuilder.Delete(dao.Table()).Where(squirrel.Eq{"course_id": testData[0].ID}).ToSql()
		_, err := db.Exec(query, args...)
		require.Nil(t, err)

		// Course ID
		cp := &models.CourseProgress{}
		require.ErrorContains(t, dao.Create(cp, nil), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", dao.Table()))
		cp.CourseID = ""
		require.ErrorContains(t, dao.Create(cp, nil), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", dao.Table()))
		cp.CourseID = "1234"

		// Invalid Course ID
		require.ErrorContains(t, dao.Create(cp, nil), "FOREIGN KEY constraint failed")
		cp.CourseID = testData[0].ID

		// Success
		require.Nil(t, dao.Create(cp, nil))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseProgress_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		dao, db := CourseProgressSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Build()

		cp, err := dao.Get(testData[0].ID, nil)
		require.Nil(t, err)
		require.Equal(t, testData[0].ID, cp.CourseID)
	})

	t.Run("not found", func(t *testing.T) {
		dao, _ := CourseProgressSetup(t)

		cp, err := dao.Get("1234", nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		require.Nil(t, cp)
	})

	t.Run("empty id", func(t *testing.T) {
		dao, _ := CourseProgressSetup(t)

		cp, err := dao.Get("", nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		require.Nil(t, cp)
	})

	t.Run("db error", func(t *testing.T) {
		dao, db := CourseProgressSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
		require.Nil(t, err)

		_, err = dao.Get("1234", nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseProgress_Update(t *testing.T) {
	t.Run("status", func(t *testing.T) {
		dao, db := CourseProgressSetup(t)

		// Create a course with 2 assets
		apDao := NewAssetProgressDao(db)
		testData := NewTestBuilder(t).Db(db).Courses(1).Assets(2).Build()

		// Create an asset progress for the first asset
		assetProgressDao := NewAssetProgressDao(db)
		ap1 := &models.AssetProgress{
			AssetID:  testData[0].Assets[0].ID,
			CourseID: testData[0].ID,
		}

		require.Nil(t, assetProgressDao.Create(ap1, nil))

		// Ensure the course percent is 0, started is false, and the started_at and completed_at are not set
		origCp, err := dao.Get(testData[0].ID, nil)
		require.Nil(t, err)
		require.False(t, origCp.Started)
		require.True(t, origCp.StartedAt.IsZero())
		require.Zero(t, origCp.Percent)
		require.True(t, origCp.CompletedAt.IsZero())

		// ----------------------------
		// Set the first asset to completed
		// ----------------------------
		time.Sleep(1 * time.Millisecond)
		ap1.Completed = true
		require.Nil(t, apDao.Update(ap1, nil))

		// Check the course percent is 50, started is true, started_at is set and completed_at is not set
		updatedCp1, err := dao.Get(origCp.CourseID, nil)
		require.Nil(t, err)
		require.True(t, updatedCp1.Started)
		require.False(t, updatedCp1.StartedAt.IsZero())
		require.Equal(t, 50, updatedCp1.Percent)
		require.True(t, updatedCp1.CompletedAt.IsZero())

		// ----------------------------
		// Set the second asset to completed
		// ----------------------------
		ap2 := &models.AssetProgress{
			AssetID:   testData[0].Assets[1].ID,
			CourseID:  testData[0].ID,
			Completed: true,
		}

		require.Nil(t, apDao.Create(ap2, nil))

		// Check the course percent is 100, started is true, and started_at and completed_at are set
		updatedCp2, err := dao.Get(origCp.CourseID, nil)
		require.Nil(t, err)
		require.True(t, updatedCp2.Started)
		require.False(t, updatedCp2.StartedAt.IsZero())
		require.Equal(t, updatedCp2.StartedAt.String(), updatedCp1.StartedAt.String())
		require.Equal(t, 100, updatedCp2.Percent)
		require.False(t, updatedCp2.CompletedAt.IsZero())

		// ----------------------------
		// Set the second asset as uncompleted
		// ----------------------------
		ap2.Completed = false
		require.Nil(t, apDao.Update(ap2, nil))

		// Check the course percent is 50, started is true, started_at is set and completed_at is not set
		updatedCp3, err := dao.Get(origCp.CourseID, nil)
		require.Nil(t, err)
		require.True(t, updatedCp3.Started)
		require.False(t, updatedCp3.StartedAt.IsZero())
		require.Equal(t, updatedCp3.StartedAt.String(), updatedCp2.StartedAt.String())
		require.Equal(t, 50, updatedCp3.Percent)
		require.True(t, updatedCp3.CompletedAt.IsZero())

		// ----------------------------
		// Set the first asset as uncompleted
		// ----------------------------
		time.Sleep(1 * time.Millisecond)
		ap1.Completed = false
		require.Nil(t, apDao.Update(ap1, nil))

		// Check the percent is 0, started is false and started_at and completed_at are not set
		updatedCp4, err := dao.Get(origCp.CourseID, nil)
		require.Nil(t, err)
		require.False(t, updatedCp4.Started)
		require.True(t, updatedCp4.StartedAt.IsZero())
		require.Zero(t, updatedCp4.Percent)
		require.True(t, updatedCp4.CompletedAt.IsZero())
	})

	t.Run("empty id", func(t *testing.T) {
		dao, db := CourseProgressSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Build()
		origCp, err := dao.Get(testData[0].ID, nil)
		require.Nil(t, err)

		origCp.CourseID = ""

		err = dao.Refresh(origCp.CourseID, nil)
		require.EqualError(t, err, "id cannot be empty")
	})

	t.Run("db error", func(t *testing.T) {
		dao, db := CourseProgressSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Build()
		origCp, err := dao.Get(testData[0].ID, nil)
		require.Nil(t, err)

		_, err = db.Exec("DROP TABLE IF EXISTS " + dao.Table())
		require.Nil(t, err)

		err = dao.Refresh(origCp.CourseID, nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseProgress_DeleteCascade(t *testing.T) {
	dao, db := CourseProgressSetup(t)

	testData := NewTestBuilder(t).Db(db).Courses(1).Build()

	// Delete the course
	courseDao := NewCourseDao(db)
	err := courseDao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": testData[0].ID}}, nil)
	require.Nil(t, err)

	// Check the course progress was deleted
	cp, err := dao.Get(testData[0].ID, nil)
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.Nil(t, cp)
}
