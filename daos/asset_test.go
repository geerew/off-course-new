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
		require.Zero(t, count)
	})

	t.Run("entries", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		NewTestBuilder(t).Db(db).Courses(5).Assets(1).Build()

		count, err := dao.Count(nil)
		require.Nil(t, err)
		require.Equal(t, count, 5)
	})

	t.Run("where", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(3).Assets(2).Build()

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		count, err := dao.Count(&database.DatabaseParams{Where: squirrel.Eq{dao.Table() + ".id": testData[0].Assets[1].ID}})
		require.Nil(t, err)
		require.Equal(t, 1, count)

		// ----------------------------
		// NOT EQUALS ID
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.NotEq{dao.Table() + ".id": testData[0].Assets[1].ID}})
		require.Nil(t, err)
		require.Equal(t, 5, count)

		// ----------------------------
		// EQUALS COURSE_ID
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.Eq{dao.Table() + ".course_id": testData[1].ID}})
		require.Nil(t, err)
		require.Equal(t, 2, count)

		// ----------------------------
		// ERROR
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		require.Zero(t, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
		require.Nil(t, err)

		_, err = dao.Count(nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		testData := NewTestBuilder(t).Courses(1).Assets(1).Build()

		// Create the course
		courseDao := NewCourseDao(db)
		require.Nil(t, courseDao.Create(testData[0].Course))

		// Create the asset
		err := dao.Create(testData[0].Assets[0])
		require.Nil(t, err)

		newA, err := dao.Get(testData[0].Assets[0].ID, nil, nil)
		require.Nil(t, err)
		require.Equal(t, testData[0].Assets[0].ID, newA.ID)
		require.Equal(t, testData[0].Assets[0].CourseID, newA.CourseID)
		require.Equal(t, testData[0].Assets[0].Title, newA.Title)
		require.Equal(t, testData[0].Assets[0].Prefix, newA.Prefix)
		require.Equal(t, testData[0].Assets[0].Chapter, newA.Chapter)
		require.Equal(t, testData[0].Assets[0].Type, newA.Type)
		require.Equal(t, testData[0].Assets[0].Path, newA.Path)
		require.False(t, newA.CreatedAt.IsZero())
		require.False(t, newA.UpdatedAt.IsZero())

		// Progress
		require.Zero(t, newA.VideoPos)
		require.False(t, newA.Completed)
		require.True(t, newA.CompletedAt.IsZero())
	})

	t.Run("duplicate paths", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Assets(1).Build()

		// Create the asset (again)
		err := dao.Create(testData[0].Assets[0])
		require.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.path", dao.Table()))
	})

	t.Run("constraints", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Build()

		// No course ID
		asset := &models.Asset{}
		require.ErrorContains(t, dao.Create(asset), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", dao.Table()))
		asset.CourseID = ""
		require.ErrorContains(t, dao.Create(asset), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", dao.Table()))
		asset.CourseID = "1234"

		// No title
		require.ErrorContains(t, dao.Create(asset), fmt.Sprintf("NOT NULL constraint failed: %s.title", dao.Table()))
		asset.Title = ""
		require.ErrorContains(t, dao.Create(asset), fmt.Sprintf("NOT NULL constraint failed: %s.title", dao.Table()))
		asset.Title = "Course 1"

		// No/invalid prefix
		require.ErrorContains(t, dao.Create(asset), "NOT NULL constraint failed: assets.prefix")
		asset.Prefix = sql.NullInt16{Int16: -1, Valid: true}
		require.ErrorContains(t, dao.Create(asset), "prefix must be greater than 0")
		asset.Prefix = sql.NullInt16{Int16: 1, Valid: true}

		// No type
		require.ErrorContains(t, dao.Create(asset), "NOT NULL constraint failed: assets.type")
		asset.Type = types.Asset{}
		require.ErrorContains(t, dao.Create(asset), "NOT NULL constraint failed: assets.type")
		asset.Type = *types.NewAsset("mp4")

		// No path
		require.ErrorContains(t, dao.Create(asset), "NOT NULL constraint failed: assets.path")
		asset.Path = ""
		require.ErrorContains(t, dao.Create(asset), "NOT NULL constraint failed: assets.path")
		asset.Path = "/course 1/01 asset"

		// Invalid Course ID
		require.ErrorContains(t, dao.Create(asset), "FOREIGN KEY constraint failed")

		// Success
		asset.CourseID = testData[0].ID
		require.Nil(t, dao.Create(asset))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(2).Assets(1).Attachments(2).Build()

		a, err := dao.Get(testData[0].Assets[0].ID, nil, nil)
		require.Nil(t, err)
		require.Equal(t, testData[0].Assets[0].ID, a.ID)
		require.Nil(t, a.Attachments)

		// ----------------------------
		// Progress
		// ----------------------------
		apDao := NewAssetProgressDao(db)

		require.Zero(t, a.VideoPos)
		require.False(t, a.Completed)
		require.True(t, a.CompletedAt.IsZero())

		// Set video pos
		ap := &models.AssetProgress{
			AssetID:  testData[0].Assets[0].ID,
			CourseID: testData[0].ID,
			VideoPos: 50,
		}

		require.Nil(t, apDao.Update(ap, nil))

		a, err = dao.Get(a.ID, nil, nil)
		require.Nil(t, err)
		require.Equal(t, 50, a.VideoPos)
		require.False(t, a.Completed)
		require.True(t, a.CompletedAt.IsZero())

		// Set completed
		ap.Completed = true
		require.Nil(t, apDao.Update(ap, nil))

		a, err = dao.Get(a.ID, nil, nil)
		require.Nil(t, err)
		require.Equal(t, 50, a.VideoPos)
		require.True(t, a.Completed)
		require.False(t, a.CompletedAt.IsZero())

		// ----------------------------
		// Attachments
		// ----------------------------
		a, err = dao.Get(testData[0].Assets[0].ID, &database.DatabaseParams{IncludeRelations: []string{NewAttachmentDao(dao.db).Table()}}, nil)
		require.Nil(t, err)
		require.Equal(t, testData[0].Assets[0].ID, a.ID)

		require.Len(t, a.Attachments, 2)
		require.Equal(t, testData[0].Assets[0].Attachments[0].ID, a.Attachments[0].ID)
		require.Equal(t, testData[0].Assets[0].Attachments[1].ID, a.Attachments[1].ID)
	})

	t.Run("orderby", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(3).Assets(1).Attachments(2).Build()

		attDao := NewAttachmentDao(db)

		// ----------------------------
		// ATTACHMENTS.CREATED_AT DESC
		// ----------------------------
		dbParams := &database.DatabaseParams{
			OrderBy:          []string{attDao.Table() + ".created_at desc"},
			IncludeRelations: []string{attDao.Table()},
		}

		result, err := dao.Get(testData[0].Assets[0].ID, dbParams, nil)
		require.Nil(t, err)
		require.Equal(t, testData[0].Assets[0].ID, result.ID)
		require.Equal(t, testData[0].Assets[0].Attachments[0].ID, result.Attachments[1].ID)
		require.Equal(t, testData[0].Assets[0].Attachments[1].ID, result.Attachments[0].ID)

		// ----------------------------
		// ATTACHMENTS.CREATED_AT ASC
		// ----------------------------
		dbParams = &database.DatabaseParams{
			OrderBy:          []string{attDao.Table() + ".created_at asc"},
			IncludeRelations: []string{attDao.Table()},
		}

		result, err = dao.Get(testData[0].Assets[0].ID, dbParams, nil)
		require.Nil(t, err)
		require.Equal(t, testData[0].Assets[0].ID, result.ID)
		require.Equal(t, testData[0].Assets[0].Attachments[0].ID, result.Attachments[0].ID)
		require.Equal(t, testData[0].Assets[0].Attachments[1].ID, result.Attachments[1].ID)

		// ----------------------------
		// Error
		// ----------------------------
		dbParams = &database.DatabaseParams{
			OrderBy:          []string{attDao.Table() + ".unit_test desc"},
			IncludeRelations: []string{attDao.Table()},
		}

		result, err = dao.Get(testData[0].Assets[0].ID, dbParams, nil)
		require.ErrorContains(t, err, "no such column")
		require.Nil(t, result)
	})

	t.Run("not found", func(t *testing.T) {
		_, dao, _ := assetSetup(t)

		c, err := dao.Get("1234", nil, nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		require.Nil(t, c)
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, _ := assetSetup(t)

		c, err := dao.Get("", nil, nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		require.Nil(t, c)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
		require.Nil(t, err)

		_, err = dao.Get("1234", nil, nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_List(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, dao, _ := assetSetup(t)

		assets, err := dao.List(nil, nil)
		require.Nil(t, err)
		require.Zero(t, assets)
	})

	t.Run("found", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(5).Assets(2).Attachments(3).Build()

		result, err := dao.List(nil, nil)
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Nil(t, result[0].Attachments)

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
			AssetID:  testData[0].Assets[0].ID,
			CourseID: testData[0].ID,
			VideoPos: 50,
		}

		require.Nil(t, apDao.Update(ap1, nil))
		// Find all started videos
		dbParams := &database.DatabaseParams{
			Where: squirrel.And{
				squirrel.Eq{dao.Table() + ".type": string(types.AssetVideo)},
				squirrel.Gt{apDao.Table() + ".video_pos": 0},
			},
		}
		result, err = dao.List(dbParams, nil)
		require.Nil(t, err)
		require.Len(t, result, 1)
		require.Equal(t, testData[0].Assets[0].ID, result[0].ID)
		require.Equal(t, 50, result[0].VideoPos)

		// Mark the second asset as completed
		ap2 := &models.AssetProgress{
			AssetID:     testData[1].Assets[1].ID,
			CourseID:    testData[1].ID,
			Completed:   true,
			CompletedAt: types.NowDateTime(),
		}

		require.Nil(t, apDao.Update(ap2, nil))

		// Find completed assets
		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Eq{apDao.Table() + ".completed": true}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 1)
		require.Equal(t, testData[1].Assets[1].ID, result[0].ID)
		require.True(t, result[0].Completed)
		require.False(t, result[0].CompletedAt.IsZero())

		// ----------------------------
		// Attachments
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{IncludeRelations: []string{NewAttachmentDao(dao.db).Table()}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 10)

		for _, a := range result {
			require.Len(t, a.Attachments, 3)
		}
	})

	t.Run("orderby", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(3).Assets(1).Attachments(2).Build()

		// ----------------------------
		// CREATED_AT DESC
		// ----------------------------
		dbParams := &database.DatabaseParams{OrderBy: []string{"created_at desc"}}
		result, err := dao.List(dbParams, nil)
		require.Nil(t, err)
		require.Len(t, result, 3)
		require.Equal(t, testData[2].Assets[0].ID, result[0].ID)

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{"created_at asc"}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 3)
		require.Equal(t, testData[0].Assets[0].ID, result[0].ID)

		// ----------------------------
		// CREATED_AT ASC + ATTACHMENTS.CREATED_AT DESC
		// ----------------------------
		attachmentsDao := NewAttachmentDao(db)

		result, err = dao.List(&database.DatabaseParams{
			OrderBy: []string{
				dao.Table() + ".created_at asc",
				attachmentsDao.Table() + ".created_at desc",
			},
			IncludeRelations: []string{attachmentsDao.Table()},
		}, nil)

		require.Nil(t, err)
		require.Len(t, result, 3)
		require.Equal(t, testData[0].Assets[0].ID, result[0].ID)
		require.Equal(t, testData[0].Assets[0].Attachments[1].ID, result[0].Attachments[0].ID)

		// ----------------------------
		// Error
		// ----------------------------
		dbParams = &database.DatabaseParams{OrderBy: []string{"unit_test asc"}}
		result, err = dao.List(dbParams, nil)
		require.ErrorContains(t, err, "no such column")
		require.Nil(t, result)
	})

	t.Run("where", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(3).Assets(2).Build()

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		result, err := dao.List(&database.DatabaseParams{Where: squirrel.Eq{dao.Table() + ".id": testData[0].Assets[1].ID}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 1)
		require.Equal(t, testData[0].Assets[1].ID, result[0].ID)

		// ----------------------------
		// EQUALS ID OR ID
		// ----------------------------
		dbParams := &database.DatabaseParams{
			Where:   squirrel.Or{squirrel.Eq{dao.Table() + ".id": testData[0].Assets[1].ID}, squirrel.Eq{dao.Table() + ".id": testData[1].Assets[1].ID}},
			OrderBy: []string{"created_at asc"},
		}
		result, err = dao.List(dbParams, nil)
		require.Nil(t, err)
		require.Len(t, result, 2)
		require.Equal(t, testData[0].Assets[1].ID, result[0].ID)
		require.Equal(t, testData[1].Assets[1].ID, result[1].ID)

		// ----------------------------
		// ERROR
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Eq{"": ""}}, nil)
		require.ErrorContains(t, err, "syntax error")
		require.Nil(t, result)
	})

	t.Run("pagination", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Assets(17).Build()

		// ----------------------------
		// Page 1 with 10 items
		// ----------------------------
		p := pagination.New(1, 10)

		result, err := dao.List(&database.DatabaseParams{Pagination: p}, nil)
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 17, p.TotalItems())
		require.Equal(t, testData[0].Assets[0].ID, result[0].ID)
		require.Equal(t, testData[0].Assets[9].ID, result[9].ID)

		// ----------------------------
		// Page 2 with 7 items
		// ----------------------------
		p = pagination.New(2, 10)

		result, err = dao.List(&database.DatabaseParams{Pagination: p}, nil)
		require.Nil(t, err)
		require.Len(t, result, 7)
		require.Equal(t, 17, p.TotalItems())
		require.Equal(t, testData[0].Assets[10].ID, result[0].ID)
		require.Equal(t, testData[0].Assets[16].ID, result[6].ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
		require.Nil(t, err)

		_, err = dao.List(nil, nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Assets(1).Build()
		err := dao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": testData[0].Assets[0].ID}}, nil)
		require.Nil(t, err)
	})

	t.Run("no db params", func(t *testing.T) {
		_, dao, _ := assetSetup(t)

		err := dao.Delete(nil, nil)
		require.ErrorIs(t, err, ErrMissingWhere)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := assetSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
		require.Nil(t, err)

		err = dao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": "1234"}}, nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_DeleteCascade(t *testing.T) {
	_, dao, db := assetSetup(t)

	testData := NewTestBuilder(t).Db(db).Courses(1).Assets(1).Build()

	// Delete the course
	courseDao := NewCourseDao(db)
	err := courseDao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": testData[0].ID}}, nil)
	require.Nil(t, err)

	// Check the asset was deleted
	a, err := dao.Get(testData[0].Assets[0].ID, nil, nil)
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.Nil(t, a)
}
