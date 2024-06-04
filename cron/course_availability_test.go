package cron

import (
	"fmt"
	"testing"

	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/mocks"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseAvailability_Run(t *testing.T) {
	t.Run("single update", func(t *testing.T) {
		dbManager, appFs, logger, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Build()

		// ----------------------------
		// true -> false
		// ----------------------------

		// Mark the course as available
		testData[0].Course.Available = true
		require.Nil(t, daos.NewCourseDao(dbManager.DataDb).Update(testData[0].Course, nil))

		ca := &courseAvailability{
			db:        dbManager.DataDb,
			appFs:     appFs,
			logger:    logger,
			batchSize: 1,
		}

		err := ca.run()
		require.Nil(t, err)

		// Check the course is marked as unavailable
		course, err := daos.NewCourseDao(dbManager.DataDb).Get(testData[0].Course.ID, nil, nil)
		require.Nil(t, err)
		require.False(t, course.Available)

		// ----------------------------
		// false -> true
		// ----------------------------

		// Create course directory
		require.Nil(t, appFs.Fs.MkdirAll(course.Path, 0755))

		err = ca.run()
		require.Nil(t, err)

		// Check the course is marked as available
		course, err = daos.NewCourseDao(dbManager.DataDb).Get(testData[0].Course.ID, nil, nil)
		require.Nil(t, err)
		require.True(t, course.Available)
	})

	t.Run("multi update", func(t *testing.T) {
		dbManager, appFs, logger, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(3).Build()

		// ----------------------------
		// true -> false
		// ----------------------------

		// Mark the courses as available
		for _, data := range testData {
			data.Course.Available = true
			require.Nil(t, daos.NewCourseDao(dbManager.DataDb).Update(data.Course, nil))
		}

		ca := &courseAvailability{
			db:        dbManager.DataDb,
			appFs:     appFs,
			logger:    logger,
			batchSize: 2,
		}

		err := ca.run()
		require.Nil(t, err)

		// Check the courses are marked as unavailable
		for _, data := range testData {
			course, err := daos.NewCourseDao(dbManager.DataDb).Get(data.Course.ID, nil, nil)
			require.Nil(t, err)
			require.False(t, course.Available)
		}

		// ----------------------------
		// false -> true
		// ----------------------------

		// Create course directories
		for _, data := range testData {
			require.Nil(t, appFs.Fs.MkdirAll(data.Course.Path, 0755))
		}

		err = ca.run()
		require.Nil(t, err)

		// Check the courses are marked as available
		for _, data := range testData {
			course, err := daos.NewCourseDao(dbManager.DataDb).Get(data.Course.ID, nil, nil)
			require.Nil(t, err)
			require.True(t, course.Available)
		}
	})

	t.Run("stat error", func(t *testing.T) {
		dbManager, _, logger, logs := setup(t)

		daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Build()

		fsWithError := &mocks.MockFsWithError{
			Fs:          afero.NewMemMapFs(),
			ErrToReturn: fmt.Errorf("stat error"),
		}

		caAppFs := appFs.NewAppFs(fsWithError, logger)

		ca := &courseAvailability{
			db:        dbManager.DataDb,
			appFs:     caAppFs,
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
		dbManager, appFs, logger, logs := setup(t)

		// Drop the table
		_, err := dbManager.DataDb.Exec("DROP TABLE IF EXISTS " + daos.NewCourseDao(dbManager.DataDb).Table())
		require.Nil(t, err)

		ca := &courseAvailability{
			db:        dbManager.DataDb,
			appFs:     appFs,
			logger:    logger,
			batchSize: 1,
		}

		err = ca.run()
		require.ErrorContains(t, err, "no such table: "+daos.NewCourseDao(dbManager.DataDb).Table())

		// Check the logger
		require.Len(t, *logs, 2)
		require.Equal(t, "Failed to fetch courses", (*logs)[1].Message)
	})
}
