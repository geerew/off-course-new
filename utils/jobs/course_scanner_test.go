package jobs

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"sync"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// setupCourseScanner initializes the CourseScanner and its dependencies
func setupCourseScanner(t *testing.T) (*CourseScanner, *database.DatabaseManager, *[]*logger.Log, *sync.Mutex) {
	t.Helper()

	// Logger
	var logs []*logger.Log
	var logsMux sync.Mutex
	logger, _, err := logger.InitLogger(&logger.BatchOptions{
		BatchSize: 1,
		WriteFn:   logger.TestWriteFn(&logs, &logsMux),
	})
	require.NoError(t, err, "Failed to initialize logger")

	// Filesystem
	appFs := appFs.NewAppFs(afero.NewMemMapFs(), logger)

	// Db
	dbManager, err := database.NewSqliteDBManager(&database.DatabaseConfig{
		IsDebug:  false,
		DataDir:  "./oc_data",
		AppFs:    appFs,
		InMemory: true,
	})

	require.Nil(t, err)
	require.NotNil(t, dbManager)

	// Course scanner
	courseScanner := NewCourseScanner(&CourseScannerConfig{
		Db:     dbManager.DataDb,
		AppFs:  appFs,
		Logger: logger,
	})

	return courseScanner, dbManager, &logs, &logsMux
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseScanner_Add(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Build()

		scan, err := scanner.Add(testData[0].ID)
		require.Nil(t, err)
		require.Equal(t, testData[0].ID, scan.CourseID)
	})

	t.Run("duplicate", func(t *testing.T) {
		scanner, dbManager, logs, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Build()

		scan, err := scanner.Add(testData[0].ID)
		require.Nil(t, err)
		require.Equal(t, testData[0].ID, scan.CourseID)

		// Add the same course again
		scan, err = scanner.Add(testData[0].ID)
		require.Nil(t, err)
		require.Nil(t, scan)
		require.NotEmpty(t, *logs)
		require.Equal(t, "Scan already in progress", (*logs)[len(*logs)-1].Message)
		require.Equal(t, slog.LevelDebug, (*logs)[len(*logs)-1].Level)
	})

	t.Run("invalid course", func(t *testing.T) {
		scanner, _, _, _ := setupCourseScanner(t)

		scan, err := scanner.Add("test")
		require.ErrorIs(t, err, sql.ErrNoRows)
		require.Nil(t, scan)
	})

	t.Run("not blocked", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(2).Build()

		scan1, err := scanner.Add(testData[0].ID)
		require.Nil(t, err)
		require.Equal(t, testData[0].ID, scan1.CourseID)

		scan2, err := scanner.Add(testData[1].ID)
		require.Nil(t, err)
		require.Equal(t, testData[1].ID, scan2.CourseID)
	})

	t.Run("db error", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Build()
		scanDao := daos.NewScanDao(dbManager.DataDb)

		_, err := dbManager.DataDb.Exec("DROP TABLE IF EXISTS " + scanDao.Table())
		require.Nil(t, err)

		scan, err := scanner.Add(testData[0].ID)
		require.ErrorContains(t, err, fmt.Sprintf("no such table: %s", scanDao.Table()))
		require.Nil(t, scan)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseScanner_Worker(t *testing.T) {
	t.Run("single job", func(t *testing.T) {
		scanner, dbManager, logs, logsMux := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Build()

		// Start the worker
		var processingDone = make(chan bool, 1)
		go scanner.Worker(func(*CourseScanner, *models.Scan) error {
			return nil
		}, processingDone)

		// Add the job
		scan, err := scanner.Add(testData[0].ID)
		require.Nil(t, err)
		require.Equal(t, scan.CourseID, testData[0].ID)

		// Wait for the worker to finish
		<-processingDone

		// Assert the scan job was deleted from the DB
		s, err := scanner.scanDao.Get(testData[0].ID, nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		require.Nil(t, s)

		logsMux.Lock()
		defer logsMux.Unlock()

		require.NotEmpty(t, *logs)
		require.Equal(t, "Finished processing all jobs", (*logs)[len(*logs)-1].Message)
		require.Equal(t, slog.LevelDebug, (*logs)[len(*logs)-1].Level)
	})

	t.Run("several jobs", func(t *testing.T) {
		scanner, dbManager, logs, logsMux := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(3).Build()

		for _, course := range testData {
			_, err := scanner.Add(course.ID)
			require.Nil(t, err)
		}

		// Start the worker
		var processingDone = make(chan bool, 1)
		go scanner.Worker(func(*CourseScanner, *models.Scan) error {
			return nil
		}, processingDone)

		// Wait for the worker to finish
		<-processingDone

		// Assert the scan job was deleted from the DB
		for _, course := range testData {
			s, err := scanner.scanDao.Get(course.ID, nil)
			require.ErrorIs(t, err, sql.ErrNoRows)
			require.Nil(t, s)
		}

		logsMux.Lock()
		defer logsMux.Unlock()

		require.NotEmpty(t, *logs)
		require.Equal(t, "Finished processing all jobs", (*logs)[len(*logs)-1].Message)
		require.Equal(t, slog.LevelDebug, (*logs)[len(*logs)-1].Level)
	})

	t.Run("error processing", func(t *testing.T) {
		scanner, dbManager, logs, logsMux := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Build()

		// Start the worker
		var processingDone = make(chan bool, 1)
		go scanner.Worker(func(*CourseScanner, *models.Scan) error {
			return errors.New("processing error")
		}, processingDone)

		scan, err := scanner.Add(testData[0].ID)
		require.Nil(t, err)
		require.Equal(t, scan.CourseID, testData[0].ID)

		// Wait for the worker to finish
		<-processingDone

		logsMux.Lock()
		defer logsMux.Unlock()

		require.NotEmpty(t, *logs)
		require.Greater(t, len(*logs), 2)
		require.Equal(t, "Failed to process scan job", (*logs)[len(*logs)-2].Message)
		require.Equal(t, slog.LevelError, (*logs)[len(*logs)-2].Level)
		require.Equal(t, "Finished processing all jobs", (*logs)[len(*logs)-1].Message)
		require.Equal(t, slog.LevelDebug, (*logs)[len(*logs)-1].Level)
	})

	t.Run("scan error", func(t *testing.T) {
		scanner, dbManager, logs, logsMux := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Build()
		scanDao := daos.NewScanDao(dbManager.DataDb)

		// Add the job
		scan, err := scanner.Add(testData[0].ID)
		require.Nil(t, err)
		require.Equal(t, scan.CourseID, testData[0].ID)

		// Drop the DB
		_, err = dbManager.DataDb.Exec("DROP TABLE IF EXISTS " + scanDao.Table())
		require.Nil(t, err)

		// Start the worker
		var processingDone = make(chan bool, 1)
		go scanner.Worker(func(*CourseScanner, *models.Scan) error {
			return nil
		}, processingDone)

		// Wait for the worker to finish
		<-processingDone

		logsMux.Lock()
		defer logsMux.Unlock()

		require.NotEmpty(t, *logs)
		require.Equal(t, "Failed to look up next scan job", (*logs)[len(*logs)-1].Message)
		require.Equal(t, slog.LevelError, (*logs)[len(*logs)-1].Level)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseScanner_CourseProcessor(t *testing.T) {
	t.Run("scan nil", func(t *testing.T) {
		scanner, _, _, _ := setupCourseScanner(t)

		err := CourseProcessor(scanner, nil)
		require.EqualError(t, err, "scan cannot be empty")
	})

	t.Run("error getting course", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Scan().Build()
		courseDao := daos.NewCourseDao(dbManager.DataDb)

		// Drop the table
		_, err := dbManager.DataDb.Exec("DROP TABLE IF EXISTS " + courseDao.Table())
		require.Nil(t, err)

		err = CourseProcessor(scanner, testData[0].Scan)
		require.ErrorContains(t, err, fmt.Sprintf("no such table: %s", courseDao.Table()))
	})

	t.Run("course unavailable", func(t *testing.T) {
		scanner, dbManager, logs, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Scan().Build()

		// Mark the course as available
		testData[0].Available = true
		err := scanner.courseDao.Update(testData[0].Course, nil)
		require.Nil(t, err)

		err = CourseProcessor(scanner, testData[0].Scan)
		require.Nil(t, err)

		require.NotEmpty(t, *logs)
		require.Equal(t, "Ignoring scan job as the course path does not exist", (*logs)[len(*logs)-1].Message)
		require.Equal(t, slog.LevelDebug, (*logs)[len(*logs)-1].Level)
	})

	t.Run("card", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		// ----------------------------
		// Found card
		// ----------------------------
		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Scan().Build()
		require.Empty(t, testData[0].CardPath)

		scanner.appFs.Fs.Mkdir(testData[0].Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/card.jpg", testData[0].Path))

		err := CourseProcessor(scanner, testData[0].Scan)
		require.Nil(t, err)

		c, err := scanner.courseDao.Get(testData[0].ID, nil, nil)
		require.Nil(t, err)
		require.Equal(t, fmt.Sprintf("%s/card.jpg", testData[0].Path), c.CardPath)

		// ----------------------------
		// Ignore card in chapter
		// ----------------------------
		testData = daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Scan().Build()
		require.Empty(t, testData[0].CardPath)

		scanner.appFs.Fs.Mkdir(testData[0].Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/chapter 1/card.jpg", testData[0].Path))

		err = CourseProcessor(scanner, testData[0].Scan)
		require.Nil(t, err)

		c, err = scanner.courseDao.Get(testData[0].ID, nil, nil)
		require.Nil(t, err)
		require.Empty(t, c.CardPath)
		require.Empty(t, testData[0].CardPath)

		// ----------------------------
		// Multiple cards types
		// ----------------------------
		testData = daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Scan().Build()
		require.Empty(t, testData[0].CardPath)

		scanner.appFs.Fs.Mkdir(testData[0].Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/card.jpg", testData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/card.png", testData[0].Path))

		err = CourseProcessor(scanner, testData[0].Scan)
		require.Nil(t, err)

		c, err = scanner.courseDao.Get(testData[0].ID, nil, nil)
		require.Nil(t, err)
		require.Equal(t, fmt.Sprintf("%s/card.jpg", testData[0].Path), c.CardPath)
	})

	t.Run("card error", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Scan().Build()
		require.Empty(t, testData[0].CardPath)

		scanner.appFs.Fs.Mkdir(testData[0].Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/card.jpg", testData[0].Path))

		// Rename the card_path column
		courseDao := daos.NewCourseDao(dbManager.DataDb)
		_, err := dbManager.DataDb.Exec(fmt.Sprintf("ALTER TABLE %s RENAME COLUMN card_path TO ignore_card_path", courseDao.Table()))
		require.Nil(t, err)

		err = CourseProcessor(scanner, testData[0].Scan)
		require.ErrorContains(t, err, "no such column: card_path")
	})

	t.Run("ignore files", func(t *testing.T) {
		scanner, dbManager, logs, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Scan().Build()

		scanner.appFs.Fs.Mkdir(testData[0].Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/file 1", testData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/file.file", testData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/file.avi", testData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/ - file.avi", testData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/- - file.avi", testData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/.avi", testData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/-1 - file.avi", testData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/a - file.avi", testData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/1.1 - file.avi", testData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/2.3-file.avi", testData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/1file.avi", testData[0].Path))

		err := CourseProcessor(scanner, testData[0].Scan)
		require.Nil(t, err)

		assetDao := daos.NewAssetDao(dbManager.DataDb)

		assets, err := scanner.assetDao.List(&database.DatabaseParams{Where: squirrel.Eq{assetDao.Table() + ".course_id": testData[0].ID}}, nil)
		require.Nil(t, err)
		require.Zero(t, len(assets))
		require.NotEmpty(t, *logs)
		require.Equal(t, "Ignoring file during scan job", (*logs)[len(*logs)-1].Message)
		require.Equal(t, slog.LevelDebug, (*logs)[len(*logs)-1].Level)
	})

	t.Run("assets", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Scan().Build()
		assetDao := daos.NewAssetDao(dbManager.DataDb)

		dbParams := &database.DatabaseParams{
			OrderBy: []string{assetDao.Table() + ".chapter asc", assetDao.Table() + ".prefix asc"},
			Where:   squirrel.Eq{assetDao.Table() + ".course_id": testData[0].ID},
		}

		// ----------------------------
		// Add 2 assets
		// ----------------------------
		scanner.appFs.Fs.Mkdir(testData[0].Path, os.ModePerm)

		afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 file 1.mkv", testData[0].Path), []byte("file 1"), os.ModePerm)
		afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/02 file 2.html", testData[0].Path), []byte("file 2"), os.ModePerm)
		afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/should ignore", testData[0].Path), []byte("ignore"), os.ModePerm)

		err := CourseProcessor(scanner, testData[0].Scan)
		require.Nil(t, err)

		assets, err := scanner.assetDao.List(dbParams, nil)
		require.Nil(t, err)
		require.Len(t, assets, 2)

		require.Equal(t, "file 1", assets[0].Title)
		require.Equal(t, testData[0].ID, assets[0].CourseID)
		require.Equal(t, 1, int(assets[0].Prefix.Int16))
		require.Empty(t, assets[0].Chapter)
		require.True(t, assets[0].Type.IsVideo())
		require.Equal(t, "ca934260de4b6eb696e4e9912447bc7f2bd7b614da6879b7addef8e03dca71d1", assets[0].Hash)

		require.Equal(t, "file 2", assets[1].Title)
		require.Equal(t, testData[0].ID, assets[1].CourseID)
		require.Equal(t, 2, int(assets[1].Prefix.Int16))
		require.Empty(t, assets[1].Chapter)
		require.True(t, assets[1].Type.IsHTML())
		require.Equal(t, "21b5bfe70ae6b203182d12bdde12f6f086000e37c894187a47b664ea7ec2331a", assets[1].Hash)

		// ----------------------------
		// Delete asset
		// ----------------------------
		scanner.appFs.Fs.Remove(fmt.Sprintf("%s/01 file 1.mkv", testData[0].Path))

		err = CourseProcessor(scanner, testData[0].Scan)
		require.Nil(t, err)

		assets, err = scanner.assetDao.List(dbParams, nil)
		require.Nil(t, err)
		require.Len(t, assets, 1)

		require.Equal(t, "file 2", assets[0].Title)
		require.Equal(t, testData[0].ID, assets[0].CourseID)
		require.Equal(t, 2, int(assets[0].Prefix.Int16))
		require.Empty(t, assets[0].Chapter)
		require.True(t, assets[0].Type.IsHTML())
		require.Equal(t, "21b5bfe70ae6b203182d12bdde12f6f086000e37c894187a47b664ea7ec2331a", assets[0].Hash)

		// ----------------------------
		// Add chapter asset
		// ----------------------------
		afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 Chapter 1/01 file 3.pdf", testData[0].Path), []byte("file 3"), os.ModePerm)

		err = CourseProcessor(scanner, testData[0].Scan)
		require.Nil(t, err)

		assets, err = scanner.assetDao.List(dbParams, nil)
		require.Nil(t, err)
		require.Len(t, assets, 2)

		require.Equal(t, "file 2", assets[0].Title)
		require.Equal(t, testData[0].ID, assets[0].CourseID)
		require.Equal(t, 2, int(assets[0].Prefix.Int16))
		require.Empty(t, assets[0].Chapter)
		require.True(t, assets[0].Type.IsHTML())
		require.Equal(t, "21b5bfe70ae6b203182d12bdde12f6f086000e37c894187a47b664ea7ec2331a", assets[0].Hash)

		require.Equal(t, "file 3", assets[1].Title)
		require.Equal(t, testData[0].ID, assets[1].CourseID)
		require.Equal(t, 1, int(assets[1].Prefix.Int16))
		require.Equal(t, "01 Chapter 1", assets[1].Chapter)
		require.True(t, assets[1].Type.IsPDF())
		require.Equal(t, "333940e348f410361b399939d5e120c72896843ad2bea2e5a961cba6818a9ad9", assets[1].Hash)
	})

	t.Run("assets error", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Scan().Build()
		assetDao := daos.NewAssetDao(dbManager.DataDb)

		scanner.appFs.Fs.Mkdir(testData[0].Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 video.mkv", testData[0].Path))

		_, err := dbManager.DataDb.Exec("DROP TABLE IF EXISTS " + assetDao.Table())
		require.Nil(t, err)

		err = CourseProcessor(scanner, testData[0].Scan)
		require.ErrorContains(t, err, "no such table: "+assetDao.Table())
	})

	t.Run("attachments", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Scan().Build()
		assetDao := daos.NewAssetDao(dbManager.DataDb)
		attachmentDao := daos.NewAttachmentDao(dbManager.DataDb)

		scanner.appFs.Fs.Mkdir(testData[0].Path, os.ModePerm)

		assetDbParams := &database.DatabaseParams{
			Where:            squirrel.Eq{assetDao.Table() + ".course_id": testData[0].ID},
			IncludeRelations: []string{attachmentDao.Table()},
		}

		attachmentDbParams := &database.DatabaseParams{
			OrderBy: []string{"created_at asc"},
			Where:   squirrel.Eq{attachmentDao.Table() + ".course_id": testData[0].ID},
		}

		// ----------------------------
		// Add 1 asset with 1 attachment
		// ----------------------------
		afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 video.mp4", testData[0].Path), []byte("video"), os.ModePerm)
		afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 info.txt", testData[0].Path), []byte("info"), os.ModePerm)

		err := CourseProcessor(scanner, testData[0].Scan)
		require.Nil(t, err)

		assets, err := scanner.assetDao.List(assetDbParams, nil)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		require.Equal(t, "video", assets[0].Title)
		require.Equal(t, testData[0].ID, assets[0].CourseID)
		require.Equal(t, 1, int(assets[0].Prefix.Int16))
		require.Equal(t, fmt.Sprintf("%s/01 video.mp4", testData[0].Path), assets[0].Path)
		require.True(t, assets[0].Type.IsVideo())
		require.Equal(t, "e56ca866bff1691433766c60304a96583c1a410e53b33ef7d89cb29eac2a97ab", assets[0].Hash)

		attachments, err := scanner.attachmentDao.List(attachmentDbParams, nil)
		require.Nil(t, err)
		require.Len(t, attachments, 1)
		require.Equal(t, "info.txt", attachments[0].Title)
		require.Equal(t, fmt.Sprintf("%s/01 info.txt", testData[0].Path), attachments[0].Path)

		// ----------------------------
		// Add another attachment
		// ----------------------------
		afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 code.zip", testData[0].Path), []byte("code"), os.ModePerm)

		err = CourseProcessor(scanner, testData[0].Scan)
		require.Nil(t, err)

		// Assert the asset
		assets, err = scanner.assetDao.List(assetDbParams, nil)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		require.Len(t, assets[0].Attachments, 2)

		attachments, err = scanner.attachmentDao.List(attachmentDbParams, nil)
		require.Nil(t, err)
		require.Len(t, attachments, 2)
		require.Equal(t, "info.txt", attachments[0].Title)
		require.Equal(t, fmt.Sprintf("%s/01 info.txt", testData[0].Path), attachments[0].Path)
		require.Equal(t, "code.zip", attachments[1].Title)
		require.Equal(t, fmt.Sprintf("%s/01 code.zip", testData[0].Path), attachments[1].Path)

		// ----------------------------
		// Delete first attachment
		// ----------------------------
		scanner.appFs.Fs.Remove(fmt.Sprintf("%s/01 info.txt", testData[0].Path))

		err = CourseProcessor(scanner, testData[0].Scan)
		require.Nil(t, err)

		// Assert the asset
		assets, err = scanner.assetDao.List(assetDbParams, nil)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		require.Equal(t, "video", assets[0].Title)
		require.Len(t, assets[0].Attachments, 1)

		attachments, err = scanner.attachmentDao.List(attachmentDbParams, nil)
		require.Nil(t, err)
		require.Len(t, attachments, 1)
		require.Equal(t, "code.zip", attachments[0].Title)
		require.Equal(t, fmt.Sprintf("%s/01 code.zip", testData[0].Path), attachments[0].Path)
	})

	t.Run("attachments error", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Scan().Build()
		attachmentDao := daos.NewAttachmentDao(dbManager.DataDb)

		scanner.appFs.Fs.Mkdir(testData[0].Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 video.mkv", testData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 info", testData[0].Path))

		// Drop the attachments table
		_, err := dbManager.DataDb.Exec("DROP TABLE IF EXISTS " + attachmentDao.Table())
		require.Nil(t, err)

		err = CourseProcessor(scanner, testData[0].Scan)
		require.ErrorContains(t, err, "no such table: "+attachmentDao.Table())
	})

	t.Run("asset priority", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		// ----------------------------
		// Priority is VIDEO -> HTML -> PDF
		// ----------------------------

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Scan().Build()
		assetDao := daos.NewAssetDao(dbManager.DataDb)
		attachmentDao := daos.NewAttachmentDao(dbManager.DataDb)

		scanner.appFs.Fs.Mkdir(testData[0].Path, os.ModePerm)

		assetDbParams := &database.DatabaseParams{
			Where:            squirrel.Eq{assetDao.Table() + ".course_id": testData[0].ID},
			IncludeRelations: []string{attachmentDao.Table()},
		}

		attachmentDbParams := &database.DatabaseParams{
			OrderBy: []string{"created_at asc"},
			Where:   squirrel.Eq{attachmentDao.Table() + ".course_id": testData[0].ID},
		}

		// ----------------------------
		// Add PDF (asset)
		// ----------------------------
		pdfFile := fmt.Sprintf("%s/01 doc 1.pdf", testData[0].Path)
		afero.WriteFile(scanner.appFs.Fs, pdfFile, []byte("doc 1"), os.ModePerm)

		err := CourseProcessor(scanner, testData[0].Scan)
		require.Nil(t, err)

		assets, err := scanner.assetDao.List(assetDbParams, nil)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		require.Equal(t, pdfFile, assets[0].Path)
		require.True(t, assets[0].Type.IsPDF())
		require.Equal(t, "61363a1cb5bf5514e3f9e983b6a96aeb12dd1ccff1b19938231d6b798d5832f9", assets[0].Hash)
		require.Empty(t, assets[0].Attachments)

		// ----------------------------
		// Add HTML (asset)
		// ----------------------------
		htmlFile := fmt.Sprintf("%s/01 example.html", testData[0].Path)
		afero.WriteFile(scanner.appFs.Fs, htmlFile, []byte("example"), os.ModePerm)

		err = CourseProcessor(scanner, testData[0].Scan)
		require.Nil(t, err)

		assets, err = scanner.assetDao.List(assetDbParams, nil)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		require.Equal(t, htmlFile, assets[0].Path)
		require.True(t, assets[0].Type.IsHTML())
		require.Equal(t, "43aeba97fea3bfc61a897ca37b73e79c74b2ff6ea792446764a1daf65784c971", assets[0].Hash)
		require.Len(t, assets[0].Attachments, 1)

		attachments, err := scanner.attachmentDao.List(attachmentDbParams, nil)
		require.Nil(t, err)
		require.Len(t, attachments, 1)
		require.Equal(t, pdfFile, attachments[0].Path)

		// ----------------------------
		// Add VIDEO (asset)
		// ----------------------------
		videoFile := fmt.Sprintf("%s/01 video.mp4", testData[0].Path)
		afero.WriteFile(scanner.appFs.Fs, videoFile, []byte("video"), os.ModePerm)

		err = CourseProcessor(scanner, testData[0].Scan)
		require.Nil(t, err)

		assets, err = scanner.assetDao.List(assetDbParams, nil)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		require.Equal(t, videoFile, assets[0].Path)
		require.True(t, assets[0].Type.IsVideo())
		require.Equal(t, "e56ca866bff1691433766c60304a96583c1a410e53b33ef7d89cb29eac2a97ab", assets[0].Hash)
		require.Len(t, assets[0].Attachments, 2)

		attachments, err = scanner.attachmentDao.List(attachmentDbParams, nil)
		require.Nil(t, err)
		require.Len(t, attachments, 2)
		require.Equal(t, pdfFile, attachments[0].Path)
		require.Equal(t, htmlFile, attachments[1].Path)

		// ----------------------------
		// Add PDF file (attachment)
		// ----------------------------
		pdfFile2 := fmt.Sprintf("%s/01 - e.pdf", testData[0].Path)
		afero.WriteFile(scanner.appFs.Fs, pdfFile2, []byte("e"), os.ModePerm)

		err = CourseProcessor(scanner, testData[0].Scan)
		require.Nil(t, err)

		assets, err = scanner.assetDao.List(assetDbParams, nil)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		require.Equal(t, videoFile, assets[0].Path)
		require.True(t, assets[0].Type.IsVideo())
		require.Equal(t, "e56ca866bff1691433766c60304a96583c1a410e53b33ef7d89cb29eac2a97ab", assets[0].Hash)
		require.Len(t, assets[0].Attachments, 3)

		attachments, err = scanner.attachmentDao.List(attachmentDbParams, nil)
		require.Nil(t, err)
		require.Len(t, attachments, 3)
		require.Equal(t, pdfFile, attachments[0].Path)
		require.Equal(t, htmlFile, attachments[1].Path)
		require.Equal(t, pdfFile2, attachments[2].Path)
	})

	t.Run("course updated", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Scan().Build()

		scanner.appFs.Fs.Mkdir(testData[0].Path, os.ModePerm)

		err := CourseProcessor(scanner, testData[0].Scan)
		require.Nil(t, err)

		updatedCourse, err := scanner.courseDao.Get(testData[0].ID, nil, nil)
		require.Nil(t, err)
		require.NotEqual(t, testData[0].UpdatedAt, updatedCourse.UpdatedAt)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseScanner_BuildFileInfo(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		var tests = []string{
			// No prefix
			"file",
			"file.file",
			"file.avi",
			" - file.avi",
			"- - file.avi",
			".avi",
			// Invalid prefix
			"-1 - file.avi",
			"a - file.avi",
			"1.1 - file.avi",
			"2.3-file.avi",
			"1file.avi",
		}

		for _, tt := range tests {
			fb := parseFileName(tt)
			require.Nil(t, fb)
		}
	})

	t.Run("assets", func(t *testing.T) {
		var tests = []struct {
			in       string
			expected *parsedFileName
		}{
			// Video (with varied filenames)
			{"0    file 0.avi", &parsedFileName{prefix: 0, title: "file 0", ext: "avi", attachmentTitle: "file 0.avi", asset: types.NewAsset("avi")}},
			{"001 file 1.mp4", &parsedFileName{prefix: 1, title: "file 1", ext: "mp4", attachmentTitle: "file 1.mp4", asset: types.NewAsset("mp4")}},
			{"1-file.ogg", &parsedFileName{prefix: 1, title: "file", ext: "ogg", attachmentTitle: "file.ogg", asset: types.NewAsset("ogg")}},
			{"2 - file.webm", &parsedFileName{prefix: 2, title: "file", ext: "webm", attachmentTitle: "file.webm", asset: types.NewAsset("webm")}},
			{"3 -file.m4a", &parsedFileName{prefix: 3, title: "file", ext: "m4a", attachmentTitle: "file.m4a", asset: types.NewAsset("m4a")}},
			{"4- file.opus", &parsedFileName{prefix: 4, title: "file", ext: "opus", attachmentTitle: "file.opus", asset: types.NewAsset("opus")}},
			{"5000 --- file.wav", &parsedFileName{prefix: 5000, title: "file", ext: "wav", attachmentTitle: "file.wav", asset: types.NewAsset("wav")}},
			{"0100 file.mp3", &parsedFileName{prefix: 100, title: "file", ext: "mp3", attachmentTitle: "file.mp3", asset: types.NewAsset("mp3")}},
			// PDF
			{"1 - doc.pdf", &parsedFileName{prefix: 1, title: "doc", ext: "pdf", attachmentTitle: "doc.pdf", asset: types.NewAsset("pdf")}},
			// HTML
			{"1 index.html", &parsedFileName{prefix: 1, title: "index", ext: "html", attachmentTitle: "index.html", asset: types.NewAsset("html")}},
		}

		for _, tt := range tests {
			fb := parseFileName(tt.in)
			require.Equal(t, tt.expected, fb, fmt.Sprintf("error for [%s]", tt.in))
		}
	})

	t.Run("attachments", func(t *testing.T) {
		var tests = []struct {
			in       string
			expected *parsedFileName
		}{
			// No title
			{"01", &parsedFileName{prefix: 1, title: "", attachmentTitle: "01"}},
			{"200.pdf", &parsedFileName{prefix: 200, title: "", attachmentTitle: "200.pdf"}},
			{"1 -.txt", &parsedFileName{prefix: 1, title: "", attachmentTitle: "1 -.txt"}},
			{"1 .txt", &parsedFileName{prefix: 1, title: "", attachmentTitle: "1 .txt"}},
			{"1     .pdf", &parsedFileName{prefix: 1, title: "", attachmentTitle: "1     .pdf"}},
			// No extension (fileName should have no prefix)
			{"0    file 0", &parsedFileName{prefix: 0, title: "file 0", attachmentTitle: "file 0"}},
			{"001    file 1", &parsedFileName{prefix: 1, title: "file 1", attachmentTitle: "file 1"}},
			{"1001 - file", &parsedFileName{prefix: 1001, title: "file", attachmentTitle: "file"}},
			{"0123-file", &parsedFileName{prefix: 123, title: "file", attachmentTitle: "file"}},
			{"1 --- file", &parsedFileName{prefix: 1, title: "file", attachmentTitle: "file"}},
			// Non-asset extension (fileName should have no prefix)
			{"1 file.txt", &parsedFileName{prefix: 1, title: "file", ext: "txt", attachmentTitle: "file.txt"}},
		}

		for _, tt := range tests {
			fb := parseFileName(tt.in)
			require.Equal(t, tt.expected, fb, fmt.Sprintf("error for [%s]", tt.in))
		}
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseScanner_IsCard(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		var tests = []string{
			"card",
			"1234",
			"1234.jpg",
			"jpg",
			"card.test.jpg",
			"card.txt",
		}

		for _, tt := range tests {
			require.False(t, isCard(tt))
		}
	})

	t.Run("valid", func(t *testing.T) {
		var tests = []string{
			"card.jpg",
			"card.jpeg",
			"card.png",
			"card.webp",
			"card.tiff",
		}

		for _, tt := range tests {
			require.True(t, isCard(tt))
		}
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseScanner_UpdateAssets(t *testing.T) {
	t.Run("nothing added or deleted", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Assets(10).Build()
		assetDao := daos.NewAssetDao(dbManager.DataDb)

		err := updateAssets(scanner.assetDao, nil, testData[0].ID, testData[0].Assets)
		require.Nil(t, err)

		dbParams := &database.DatabaseParams{Where: squirrel.Eq{assetDao.Table() + ".course_id": testData[0].ID}}
		count, err := scanner.assetDao.Count(dbParams, nil)
		require.Nil(t, err)
		require.Equal(t, 10, count)
	})

	t.Run("add", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Assets(12).Build()
		assetDao := daos.NewAssetDao(dbManager.DataDb)

		// Delete the assets (so we can add them again)
		for _, a := range testData[0].Assets {
			require.Nil(t, scanner.assetDao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": a.ID}}, nil))
		}

		dbParams := &database.DatabaseParams{Where: squirrel.Eq{assetDao.Table() + ".course_id": testData[0].ID}}

		// ----------------------------
		// Add 10 assets
		// ----------------------------
		err := updateAssets(scanner.assetDao, nil, testData[0].ID, testData[0].Assets[:10])
		require.Nil(t, err)

		count, err := scanner.assetDao.Count(dbParams, nil)
		require.Nil(t, err)
		require.Equal(t, 10, count)

		// ----------------------------
		// Add another 2 assets
		// ----------------------------
		testData[0].Assets[10].ID = ""
		testData[0].Assets[11].ID = ""

		err = updateAssets(scanner.assetDao, nil, testData[0].ID, testData[0].Assets)
		require.Nil(t, err)

		count, err = scanner.assetDao.Count(dbParams, nil)
		require.Nil(t, err)
		require.Equal(t, 12, count)

		// Ensure all assets have an ID
		for _, a := range testData[0].Assets {
			require.NotEmpty(t, a.ID)
		}
	})

	t.Run("delete", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Assets(12).Build()

		assetDao := daos.NewAssetDao(dbManager.DataDb)

		dbParams := &database.DatabaseParams{Where: squirrel.Eq{assetDao.Table() + ".course_id": testData[0].ID}}

		// ----------------------------
		// Remove 2 assets
		// ----------------------------
		testData[0].Assets = testData[0].Assets[2:]

		err := updateAssets(scanner.assetDao, nil, testData[0].ID, testData[0].Assets)
		require.Nil(t, err)

		count, err := scanner.assetDao.Count(dbParams, nil)
		require.Nil(t, err)
		require.Equal(t, 10, count)

		// ----------------------------
		// Remove another 2 assets
		// ----------------------------
		testData[0].Assets = testData[0].Assets[2:]

		err = updateAssets(scanner.assetDao, nil, testData[0].ID, testData[0].Assets)
		require.Nil(t, err)

		count, err = scanner.assetDao.Count(dbParams, nil)
		require.Nil(t, err)
		require.Equal(t, 8, count)
	})

	t.Run("rename", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Assets(12).Build()
		assetDao := daos.NewAssetDao(dbManager.DataDb)

		// Delete the assets (so we can add them again)
		for _, a := range testData[0].Assets {
			require.Nil(t, scanner.assetDao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": a.ID}}, nil))
		}

		dbParams := &database.DatabaseParams{Where: squirrel.Eq{assetDao.Table() + ".course_id": testData[0].ID}}

		// ----------------------------
		// Add assets
		// ----------------------------
		err := updateAssets(scanner.assetDao, nil, testData[0].ID, testData[0].Assets)
		require.Nil(t, err)

		count, err := scanner.assetDao.Count(dbParams, nil)
		require.Nil(t, err)
		require.Equal(t, 12, count)

		// ----------------------------
		// Rename 2 assets
		// ----------------------------
		testData[0].Assets[10].Prefix = sql.NullInt16{Int16: int16(rand.Intn(100-1) + 1), Valid: true}
		testData[0].Assets[10].Title = security.PseudorandomString(7)
		testData[0].Assets[10].Chapter = security.PseudorandomString(7)
		testData[0].Assets[10].Path = fmt.Sprintf("%s/%s/%d %s.mp4", testData[0].Path, testData[0].Assets[10].Chapter, testData[0].Assets[10].Prefix.Int16, testData[0].Assets[10].Title)

		testData[0].Assets[11].Prefix = sql.NullInt16{Int16: int16(rand.Intn(100-1) + 1), Valid: true}
		testData[0].Assets[11].Title = security.PseudorandomString(7)
		testData[0].Assets[11].Chapter = security.PseudorandomString(7)
		testData[0].Assets[11].Path = fmt.Sprintf("%s/%s/%d %s.mp4", testData[0].Path, testData[0].Assets[11].Chapter, testData[0].Assets[11].Prefix.Int16, testData[0].Assets[11].Title)

		err = updateAssets(scanner.assetDao, nil, testData[0].ID, testData[0].Assets)
		require.Nil(t, err)

		a, err := scanner.assetDao.List(dbParams, nil)
		require.Nil(t, err)
		require.Equal(t, 12, len(a))

		// Ensure the assets were updated
		//
		// Note: Order is not guaranteed, so we have to validate this way
		for _, asset := range a {
			if asset.ID == testData[0].Assets[10].ID {
				require.Equal(t, testData[0].Assets[10].Title, asset.Title)
				require.Equal(t, testData[0].Assets[10].Prefix, asset.Prefix)
				require.Equal(t, testData[0].Assets[10].Chapter, asset.Chapter)
				require.Equal(t, testData[0].Assets[10].Path, asset.Path)
			} else if asset.ID == testData[0].Assets[11].ID {
				require.Equal(t, testData[0].Assets[11].Title, asset.Title)
				require.Equal(t, testData[0].Assets[11].Prefix, asset.Prefix)
				require.Equal(t, testData[0].Assets[11].Chapter, asset.Chapter)
				require.Equal(t, testData[0].Assets[11].Path, asset.Path)
			}
		}
	})

	t.Run("db error", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)
		assetDao := daos.NewAssetDao(dbManager.DataDb)

		// Drop the table
		_, err := dbManager.DataDb.Exec("DROP TABLE IF EXISTS " + assetDao.Table())
		require.Nil(t, err)

		err = updateAssets(scanner.assetDao, nil, "1234", []*models.Asset{})
		require.ErrorContains(t, err, fmt.Sprintf("no such table: %s", assetDao.Table()))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseScanner_UpdateAttachments(t *testing.T) {
	t.Run("nothing added or delete)", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Assets(1).Attachments(10).Build()
		attDao := daos.NewAttachmentDao(dbManager.DataDb)

		err := updateAttachments(scanner.attachmentDao, nil, testData[0].ID, testData[0].Assets[0].Attachments)
		require.Nil(t, err)

		count, err := scanner.attachmentDao.Count(&database.DatabaseParams{Where: squirrel.Eq{attDao.Table() + ".course_id": testData[0].ID}}, nil)
		require.Nil(t, err)
		require.Equal(t, 10, count)
	})

	t.Run("add", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Assets(1).Attachments(12).Build()
		attDao := daos.NewAttachmentDao(dbManager.DataDb)

		// Delete the attachments (so we can add them again)
		for _, a := range testData[0].Assets[0].Attachments {
			require.Nil(t, scanner.attachmentDao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": a.ID}}, nil))
		}

		dbParams := &database.DatabaseParams{Where: squirrel.Eq{attDao.Table() + ".course_id": testData[0].ID}}

		// ----------------------------
		// Add 10 attachments
		// ----------------------------
		err := updateAttachments(scanner.attachmentDao, nil, testData[0].ID, testData[0].Assets[0].Attachments[:10])
		require.Nil(t, err)

		count, err := scanner.attachmentDao.Count(dbParams, nil)
		require.Nil(t, err)
		require.Equal(t, 10, count)

		// ----------------------------
		// Add another 2 attachments
		// ----------------------------
		err = updateAttachments(scanner.attachmentDao, nil, testData[0].ID, testData[0].Assets[0].Attachments)
		require.Nil(t, err)

		count, err = scanner.attachmentDao.Count(dbParams, nil)
		require.Nil(t, err)
		require.Equal(t, 12, count)
	})

	t.Run("delete", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		testData := daos.NewTestBuilder(t).Db(dbManager.DataDb).Courses(1).Assets(1).Attachments(12).Build()
		attachmentDao := daos.NewAttachmentDao(dbManager.DataDb)

		dbParams := &database.DatabaseParams{Where: squirrel.Eq{attachmentDao.Table() + ".course_id": testData[0].ID}}

		// ----------------------------
		// Remove 2 attachments
		// ----------------------------
		testData[0].Assets[0].Attachments = testData[0].Assets[0].Attachments[2:]

		err := updateAttachments(scanner.attachmentDao, nil, testData[0].ID, testData[0].Assets[0].Attachments)
		require.Nil(t, err)

		count, err := scanner.attachmentDao.Count(dbParams, nil)
		require.Nil(t, err)
		require.Equal(t, 10, count)

		// ----------------------------
		// Remove another 2 attachments
		// ----------------------------
		testData[0].Assets[0].Attachments = testData[0].Assets[0].Attachments[2:]

		err = updateAttachments(scanner.attachmentDao, nil, testData[0].ID, testData[0].Assets[0].Attachments)
		require.Nil(t, err)

		count, err = scanner.attachmentDao.Count(dbParams, nil)
		require.Nil(t, err)
		require.Equal(t, 8, count)
	})

	t.Run("db error", func(t *testing.T) {
		scanner, dbManager, _, _ := setupCourseScanner(t)

		attachmentDao := daos.NewAttachmentDao(dbManager.DataDb)

		// Drop the table
		_, err := dbManager.DataDb.Exec("DROP TABLE IF EXISTS " + attachmentDao.Table())
		require.Nil(t, err)

		err = updateAttachments(scanner.attachmentDao, nil, "1234", []*models.Attachment{})
		require.ErrorContains(t, err, fmt.Sprintf("no such table: %s", attachmentDao.Table()))
	})
}
