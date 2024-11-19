package dao

import (
	"database/sql"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateOrUpdateAssetProgress(t *testing.T) {
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

		assetProgress := &models.AssetProgress{
			AssetID:  asset.ID,
			CourseID: course.ID,
		}
		require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, assetProgress))
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.CreateOrUpdateAssetProgress(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_AssetProgressDeleteCascade(t *testing.T) {
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

	assetProgress := &models.AssetProgress{
		AssetID:  asset.ID,
		CourseID: course.ID,
	}
	require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, assetProgress))

	require.NoError(t, dao.Delete(ctx, asset, nil))

	err := dao.Get(ctx, assetProgress, &database.Options{Where: squirrel.Eq{assetProgress.Table() + ".id": assetProgress.ID}})
	require.ErrorIs(t, err, sql.ErrNoRows)
}
