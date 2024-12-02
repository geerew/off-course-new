package cron

import (
	"context"
	"fmt"
	"testing"

	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/mocks"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseAvailability_Run(t *testing.T) {
	t.Run("update", func(t *testing.T) {
		db, appFs, logger, _ := setup(t)

		dao := dao.NewDAO(db)
		ctx := context.Background()

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course-%d", i), Available: false}
			require.NoError(t, dao.CreateCourse(ctx, course))
			courses = append(courses, course)
		}

		ca := &courseAvailability{
			db:        db,
			dao:       dao,
			appFs:     appFs,
			logger:    logger,
			batchSize: 2,
		}

		err := ca.run()
		require.NoError(t, err)

		for _, course := range courses {
			require.Nil(t, appFs.Fs.MkdirAll(course.Path, 0755))
		}

		err = ca.run()
		require.NoError(t, err)

		for _, course := range courses {
			err := dao.GetById(ctx, course)
			require.NoError(t, err)
			require.True(t, course.Available)
		}
	})

	t.Run("stat error", func(t *testing.T) {
		db, _, logger, logs := setup(t)

		dao := dao.NewDAO(db)
		ctx := context.Background()

		course := &models.Course{Title: "course 1", Path: "/course-1", Available: false}
		require.NoError(t, dao.CreateCourse(ctx, course))

		fsWithError := &mocks.MockFsWithError{
			Fs:          afero.NewMemMapFs(),
			ErrToReturn: fmt.Errorf("stat error"),
		}

		ca := &courseAvailability{
			db:        db,
			dao:       dao,
			appFs:     appFs.NewAppFs(fsWithError, logger),
			logger:    logger,
			batchSize: 1,
		}

		err := ca.run()
		require.Equal(t, fmt.Errorf("stat error"), err)

		// Check the logger
		require.Len(t, *logs, 2)
		require.Equal(t, "Failed to stat course", (*logs)[1].Message)
	})

	t.Run("db error", func(t *testing.T) {
		db, appFs, logger, logs := setup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + models.COURSE_TABLE)
		require.NoError(t, err)

		ca := &courseAvailability{
			db:        db,
			dao:       dao.NewDAO(db),
			appFs:     appFs,
			logger:    logger,
			batchSize: 1,
		}

		err = ca.run()
		require.ErrorContains(t, err, "no such table: "+models.COURSE_TABLE)

		// Check the logger
		require.Len(t, *logs, 2)
		require.Equal(t, "Failed to fetch courses", (*logs)[1].Message)
	})
}
