package coursescan

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setup(t *testing.T) (*CourseScan, context.Context, *[]*logger.Log) {
	t.Helper()

	// Logger
	var logs []*logger.Log
	var logsMux sync.Mutex
	logger, _, err := logger.InitLogger(&logger.BatchOptions{
		BatchSize: 1,
		WriteFn:   logger.TestWriteFn(&logs, &logsMux),
	})
	require.NoError(t, err, "Failed to initialize logger")

	appFs := appFs.NewAppFs(afero.NewMemMapFs(), logger)

	dbManager, err := database.NewSqliteDBManager(&database.DatabaseConfig{
		IsDebug:  false,
		DataDir:  "./oc_data",
		AppFs:    appFs,
		InMemory: true,
	})

	require.NoError(t, err)
	require.NotNil(t, dbManager)

	courseScan := NewCourseScan(&CourseScanConfig{
		Db:     dbManager.DataDb,
		AppFs:  appFs,
		Logger: logger,
	})

	return courseScan, context.Background(), &logs
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanner_Add(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course1 := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course1))

		scan1, err := scanner.Add(ctx, course1.ID)
		require.NoError(t, err)
		require.Equal(t, course1.ID, scan1.CourseID)

		course2 := &models.Course{Title: "Course 2", Path: "/course-2"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course2))

		scan2, err := scanner.Add(ctx, course2.ID)
		require.NoError(t, err)
		require.Equal(t, course2.ID, scan2.CourseID)
	})

	t.Run("duplicate", func(t *testing.T) {
		scanner, ctx, logs := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		first, err := scanner.Add(ctx, course.ID)
		require.NoError(t, err)
		require.Equal(t, course.ID, first.CourseID)

		// Add again
		second, err := scanner.Add(ctx, course.ID)
		require.NoError(t, err)
		require.Equal(t, second.ID, first.ID)
		require.NotEmpty(t, *logs)
		require.Equal(t, "Scan already in progress", (*logs)[len(*logs)-1].Message)
		require.Equal(t, slog.LevelDebug, (*logs)[len(*logs)-1].Level)
	})

	t.Run("invalid course", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		scan, err := scanner.Add(ctx, "1234")
		require.ErrorIs(t, err, utils.ErrInvalidId)
		require.Nil(t, scan)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanner_Worker(t *testing.T) {
	t.Run("jobs", func(t *testing.T) {
		scanner, ctx, logs := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course-%d", i)}
			require.NoError(t, scanner.dao.CreateCourse(ctx, course))
			courses = append(courses, course)
		}

		var processingDone = make(chan bool, 1)
		go scanner.Worker(ctx, func(context.Context, *CourseScan, *models.Scan) error {
			time.Sleep(1 * time.Millisecond)
			return nil
		}, processingDone)

		// Add the courses
		for i := range 3 {
			scan, err := scanner.Add(ctx, courses[i].ID)
			require.NoError(t, err)
			require.Equal(t, scan.CourseID, courses[i].ID)
		}

		// Wait for the worker to finish
		<-processingDone

		// Sometimes the delete is slow to happen
		time.Sleep(20 * time.Millisecond)

		count, err := scanner.dao.Count(ctx, &models.Scan{}, nil)
		require.NoError(t, err)
		require.Zero(t, count)

		require.NotEmpty(t, *logs)
		require.Equal(t, "Finished processing all scan jobs", (*logs)[len(*logs)-1].Message)
		require.Equal(t, slog.LevelDebug, (*logs)[len(*logs)-1].Level)

		// Add the first 2 courses (again)
		for i := range 2 {
			scan, err := scanner.Add(ctx, courses[i].ID)
			require.NoError(t, err)
			require.Equal(t, scan.CourseID, courses[i].ID)
		}

		// Wait for the worker to finish
		<-processingDone

		count, err = scanner.dao.Count(ctx, &models.Scan{}, nil)
		require.NoError(t, err)
		require.Zero(t, count)

		require.NotEmpty(t, *logs)
		require.Equal(t, "Finished processing all scan jobs", (*logs)[len(*logs)-1].Message)
		require.Equal(t, slog.LevelDebug, (*logs)[len(*logs)-1].Level)
	})

	t.Run("error processing", func(t *testing.T) {
		scanner, ctx, logs := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		var processingDone = make(chan bool, 1)
		go scanner.Worker(ctx, func(context.Context, *CourseScan, *models.Scan) error {
			time.Sleep(1 * time.Millisecond)
			return errors.New("processing error")
		}, processingDone)

		scan, err := scanner.Add(ctx, course.ID)
		require.NoError(t, err)
		require.Equal(t, scan.CourseID, course.ID)

		// Wait for the worker to finish
		<-processingDone

		require.NotEmpty(t, *logs)
		require.Greater(t, len(*logs), 2)
		require.Equal(t, "Failed to process scan job", (*logs)[len(*logs)-2].Message)
		require.Equal(t, slog.LevelError, (*logs)[len(*logs)-2].Level)
		require.Equal(t, "Finished processing all scan jobs", (*logs)[len(*logs)-1].Message)
		require.Equal(t, slog.LevelDebug, (*logs)[len(*logs)-1].Level)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanner_Processor(t *testing.T) {
	t.Run("scan nil", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		err := Processor(ctx, scanner, nil)
		require.ErrorIs(t, err, ErrNilScan)
	})

	t.Run("error getting course", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		// Drop the table
		_, err := scanner.db.Exec("DROP TABLE IF EXISTS " + course.Table())
		require.NoError(t, err)

		err = Processor(ctx, scanner, scan)
		require.ErrorContains(t, err, fmt.Sprintf("no such table: %s", course.Table()))
	})

	t.Run("course unavailable", func(t *testing.T) {
		scanner, ctx, logs := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1", Available: true}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		err := Processor(ctx, scanner, scan)
		require.NoError(t, err)

		require.NotEmpty(t, *logs)
		require.Equal(t, "Skipping as the course is unavailable", (*logs)[len(*logs)-1].Message)
		require.Equal(t, slog.LevelDebug, (*logs)[len(*logs)-1].Level)
	})

	t.Run("mark course available", func(t *testing.T) {
		scanner, ctx, logs := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1", Available: false}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)

		err := Processor(ctx, scanner, scan)
		require.NoError(t, err)

		require.NotEmpty(t, *logs)
		require.Equal(t, "Setting unavailable course as available", (*logs)[len(*logs)-1].Message)
		require.Equal(t, slog.LevelDebug, (*logs)[len(*logs)-1].Level)
	})

	t.Run("card", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		scanner.appFs.Fs.Create(filepath.Join(course.Path, "card.jpg"))

		err := Processor(ctx, scanner, scan)
		require.NoError(t, err)

		courseResult := &models.Course{Base: models.Base{ID: course.ID}}
		err = scanner.dao.GetById(ctx, courseResult)
		require.NoError(t, err)
		require.Equal(t, filepath.Join(course.Path, "card.jpg"), courseResult.CardPath)

		// Ignore card in chapter
		scanner.appFs.Fs.Remove(filepath.Join(course.Path, "card.jpg"))
		scanner.appFs.Fs.Create(filepath.Join(course.Path, "01 Chapter 1", "card.jpg"))

		err = Processor(ctx, scanner, scan)
		require.NoError(t, err)

		courseResult = &models.Course{Base: models.Base{ID: course.ID}}
		err = scanner.dao.GetById(ctx, courseResult)
		require.NoError(t, err)
		require.Empty(t, courseResult.CardPath)

		// Ignore additional cards at the root
		scanner.appFs.Fs.Create(filepath.Join(course.Path, "card.jpg"))
		scanner.appFs.Fs.Create(filepath.Join(course.Path, "card.png"))

		err = Processor(ctx, scanner, scan)
		require.NoError(t, err)

		courseResult = &models.Course{Base: models.Base{ID: course.ID}}
		err = scanner.dao.GetById(ctx, courseResult)
		require.NoError(t, err)
		require.Equal(t, filepath.Join(course.Path, "card.jpg"), courseResult.CardPath)
	})

	t.Run("ignore files", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/file 1", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/file.file", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/file.avi", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/ - file.avi", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/- - file.avi", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/.avi", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/-1 - file.avi", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/a - file.avi", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/1.1 - file.avi", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/2.3-file.avi", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/1file.avi", course.Path))

		err := Processor(ctx, scanner, scan)
		require.NoError(t, err)

		count, err := scanner.dao.Count(ctx, &models.Asset{}, nil)
		require.NoError(t, err)
		require.Zero(t, count)
	})

	t.Run("assets", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		// Add 2 assets + 1 to ignore
		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 file 1.mkv", course.Path), []byte("file 1"), os.ModePerm)
		afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/02 file 2.html", course.Path), []byte("file 2"), os.ModePerm)
		afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/should ignore", course.Path), []byte("ignore"), os.ModePerm)

		err := Processor(ctx, scanner, scan)
		require.NoError(t, err)

		options := &database.Options{
			OrderBy: []string{models.ASSET_TABLE + ".chapter asc", models.ASSET_TABLE + ".prefix asc"},
			Where:   squirrel.Eq{models.ASSET_TABLE + ".course_id": course.ID},
		}

		assets := []*models.Asset{}
		err = scanner.dao.List(ctx, &assets, options)
		require.NoError(t, err)
		require.Len(t, assets, 2)

		require.Equal(t, "file 1", assets[0].Title)
		require.Equal(t, course.ID, assets[0].CourseID)
		require.Equal(t, 1, int(assets[0].Prefix.Int16))
		require.Empty(t, assets[0].Chapter)
		require.True(t, assets[0].Type.IsVideo())
		require.Equal(t, "ca934260de4b6eb696e4e9912447bc7f2bd7b614da6879b7addef8e03dca71d1", assets[0].Hash)

		require.Equal(t, "file 2", assets[1].Title)
		require.Equal(t, course.ID, assets[1].CourseID)
		require.Equal(t, 2, int(assets[1].Prefix.Int16))
		require.Empty(t, assets[1].Chapter)
		require.True(t, assets[1].Type.IsHTML())
		require.Equal(t, "21b5bfe70ae6b203182d12bdde12f6f086000e37c894187a47b664ea7ec2331a", assets[1].Hash)

		// Delete asset 1
		scanner.appFs.Fs.Remove(fmt.Sprintf("%s/01 file 1.mkv", course.Path))

		err = Processor(ctx, scanner, scan)
		require.NoError(t, err)

		err = scanner.dao.List(ctx, &assets, options)
		require.NoError(t, err)
		require.Len(t, assets, 1)

		require.Equal(t, "file 2", assets[0].Title)
		require.Equal(t, course.ID, assets[0].CourseID)
		require.Equal(t, 2, int(assets[0].Prefix.Int16))
		require.Empty(t, assets[0].Chapter)
		require.True(t, assets[0].Type.IsHTML())
		require.Equal(t, "21b5bfe70ae6b203182d12bdde12f6f086000e37c894187a47b664ea7ec2331a", assets[0].Hash)

		// Add asset in chapter
		afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 Chapter 1/01 file 3.pdf", course.Path), []byte("file 3"), os.ModePerm)

		err = Processor(ctx, scanner, scan)
		require.NoError(t, err)

		err = scanner.dao.List(ctx, &assets, options)
		require.NoError(t, err)
		require.Len(t, assets, 2)

		require.Equal(t, "file 2", assets[0].Title)
		require.Equal(t, course.ID, assets[0].CourseID)
		require.Equal(t, 2, int(assets[0].Prefix.Int16))
		require.Empty(t, assets[0].Chapter)
		require.True(t, assets[0].Type.IsHTML())
		require.Equal(t, "21b5bfe70ae6b203182d12bdde12f6f086000e37c894187a47b664ea7ec2331a", assets[0].Hash)

		require.Equal(t, "file 3", assets[1].Title)
		require.Equal(t, course.ID, assets[1].CourseID)
		require.Equal(t, 1, int(assets[1].Prefix.Int16))
		require.Equal(t, "01 Chapter 1", assets[1].Chapter)
		require.True(t, assets[1].Type.IsPDF())
		require.Equal(t, "333940e348f410361b399939d5e120c72896843ad2bea2e5a961cba6818a9ad9", assets[1].Hash)
	})

	t.Run("attachments", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		// Add asset
		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 file 1.mkv", course.Path), []byte("file 1"), os.ModePerm)

		err := Processor(ctx, scanner, scan)
		require.NoError(t, err)

		assetOptions := &database.Options{
			OrderBy: []string{models.ASSET_TABLE + ".chapter asc", models.ASSET_TABLE + ".prefix asc"},
			Where:   squirrel.Eq{models.ASSET_TABLE + ".course_id": course.ID},
		}

		assets := []*models.Asset{}
		err = scanner.dao.List(ctx, &assets, assetOptions)
		require.NoError(t, err)
		require.Len(t, assets, 1)

		require.Equal(t, "file 1", assets[0].Title)
		require.Equal(t, course.ID, assets[0].CourseID)
		require.Equal(t, 1, int(assets[0].Prefix.Int16))
		require.Empty(t, assets[0].Chapter)
		require.True(t, assets[0].Type.IsVideo())
		require.Equal(t, "ca934260de4b6eb696e4e9912447bc7f2bd7b614da6879b7addef8e03dca71d1", assets[0].Hash)

		// Add attachment
		afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 file 1.txt", course.Path), []byte("file 1"), os.ModePerm)

		err = Processor(ctx, scanner, scan)
		require.NoError(t, err)

		attachmentOptions := &database.Options{
			OrderBy: []string{models.ATTACHMENT_TABLE + ".created_at asc"},
			Where:   squirrel.Eq{models.ATTACHMENT_TABLE + ".course_id": course.ID},
		}

		attachments := []*models.Attachment{}
		err = scanner.dao.List(ctx, &attachments, attachmentOptions)
		require.NoError(t, err)
		require.Len(t, attachments, 1)

		require.Equal(t, "file 1.txt", attachments[0].Title)
		require.Equal(t, course.ID, attachments[0].CourseID)
		require.Equal(t, filepath.Join(course.Path, "01 file 1.txt"), attachments[0].Path)

		// Add another attachment
		afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 file 2.txt", course.Path), []byte("file 2"), os.ModePerm)

		err = Processor(ctx, scanner, scan)
		require.NoError(t, err)

		err = scanner.dao.List(ctx, &attachments, attachmentOptions)
		require.NoError(t, err)
		require.Len(t, attachments, 2)

		require.Equal(t, "file 1.txt", attachments[0].Title)
		require.Equal(t, course.ID, attachments[0].CourseID)
		require.Equal(t, filepath.Join(course.Path, "01 file 1.txt"), attachments[0].Path)

		require.Equal(t, "file 2.txt", attachments[1].Title)
		require.Equal(t, course.ID, attachments[1].CourseID)
		require.Equal(t, filepath.Join(course.Path, "01 file 2.txt"), attachments[1].Path)

		// Delete attachment
		scanner.appFs.Fs.Remove(fmt.Sprintf("%s/01 file 1.txt", course.Path))

		err = Processor(ctx, scanner, scan)
		require.NoError(t, err)

		err = scanner.dao.List(ctx, &attachments, attachmentOptions)
		require.NoError(t, err)
		require.Len(t, attachments, 1)

		require.Equal(t, "file 2.txt", attachments[0].Title)
		require.Equal(t, course.ID, attachments[0].CourseID)
		require.Equal(t, filepath.Join(course.Path, "01 file 2.txt"), attachments[0].Path)
	})

	t.Run("asset priority", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		// ----------------------------
		// Priority is VIDEO -> HTML -> PDF
		// ----------------------------

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		// Add PDF asset
		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 doc 1.pdf", course.Path), []byte("doc 1"), os.ModePerm)

		err := Processor(ctx, scanner, scan)
		require.NoError(t, err)

		assetOptions := &database.Options{
			OrderBy: []string{models.ASSET_TABLE + ".chapter asc", models.ASSET_TABLE + ".prefix asc"},
			Where:   squirrel.Eq{models.ASSET_TABLE + ".course_id": course.ID},
		}

		assets := []*models.Asset{}
		err = scanner.dao.List(ctx, &assets, assetOptions)
		require.NoError(t, err)
		require.Len(t, assets, 1)

		require.Equal(t, "doc 1", assets[0].Title)
		require.Equal(t, course.ID, assets[0].CourseID)
		require.Equal(t, 1, int(assets[0].Prefix.Int16))
		require.Empty(t, assets[0].Chapter)
		require.True(t, assets[0].Type.IsPDF())
		require.Equal(t, "61363a1cb5bf5514e3f9e983b6a96aeb12dd1ccff1b19938231d6b798d5832f9", assets[0].Hash)
		require.Len(t, assets[0].Attachments, 0)

		// Add HTML asset
		afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 index.html", course.Path), []byte("index"), os.ModePerm)

		err = Processor(ctx, scanner, scan)
		require.NoError(t, err)

		err = scanner.dao.List(ctx, &assets, assetOptions)
		require.NoError(t, err)
		require.Len(t, assets, 1)

		require.Equal(t, "index", assets[0].Title)
		require.Equal(t, course.ID, assets[0].CourseID)
		require.Equal(t, 1, int(assets[0].Prefix.Int16))
		require.Empty(t, assets[0].Chapter)
		require.True(t, assets[0].Type.IsHTML())
		require.Equal(t, "935600bd3714c889e3d03a3196cf0e90b4a6aa51af8a73f7867c8a421a1106ba", assets[0].Hash)
		require.Len(t, assets[0].Attachments, 1)

		attachmentOptions := &database.Options{
			OrderBy: []string{models.ATTACHMENT_TABLE + ".created_at asc"},
			Where:   squirrel.Eq{models.ATTACHMENT_TABLE + ".course_id": course.ID},
		}

		attachments := []*models.Attachment{}
		err = scanner.dao.List(ctx, &attachments, attachmentOptions)
		require.NoError(t, err)
		require.Len(t, attachments, 1)
		require.Equal(t, filepath.Join(course.Path, "01 doc 1.pdf"), attachments[0].Path)

		// Add VIDEO asset
		afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 video.mp4", course.Path), []byte("video"), os.ModePerm)

		err = Processor(ctx, scanner, scan)
		require.NoError(t, err)

		err = scanner.dao.List(ctx, &assets, assetOptions)
		require.NoError(t, err)
		require.Len(t, assets, 1)

		require.Equal(t, "video", assets[0].Title)
		require.Equal(t, course.ID, assets[0].CourseID)
		require.Equal(t, 1, int(assets[0].Prefix.Int16))
		require.Empty(t, assets[0].Chapter)
		require.True(t, assets[0].Type.IsVideo())
		require.Equal(t, "e56ca866bff1691433766c60304a96583c1a410e53b33ef7d89cb29eac2a97ab", assets[0].Hash)
		require.Len(t, assets[0].Attachments, 2)

		err = scanner.dao.List(ctx, &attachments, attachmentOptions)
		require.NoError(t, err)
		require.Len(t, attachments, 2)
		require.Equal(t, filepath.Join(course.Path, "01 doc 1.pdf"), attachments[0].Path)
		require.Equal(t, filepath.Join(course.Path, "01 index.html"), attachments[1].Path)

		// Add another PDF asset
		afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 doc 2.pdf", course.Path), []byte("doc 2"), os.ModePerm)

		err = Processor(ctx, scanner, scan)
		require.NoError(t, err)

		err = scanner.dao.List(ctx, &assets, assetOptions)
		require.NoError(t, err)
		require.Len(t, assets, 1)

		require.Equal(t, "video", assets[0].Title)
		require.Equal(t, course.ID, assets[0].CourseID)
		require.Equal(t, 1, int(assets[0].Prefix.Int16))
		require.Empty(t, assets[0].Chapter)
		require.True(t, assets[0].Type.IsVideo())
		require.Equal(t, "e56ca866bff1691433766c60304a96583c1a410e53b33ef7d89cb29eac2a97ab", assets[0].Hash)
		require.Len(t, assets[0].Attachments, 3)

		err = scanner.dao.List(ctx, &attachments, attachmentOptions)
		require.NoError(t, err)
		require.Len(t, attachments, 3)
		require.Equal(t, filepath.Join(course.Path, "01 doc 1.pdf"), attachments[0].Path)
		require.Equal(t, filepath.Join(course.Path, "01 index.html"), attachments[1].Path)
		require.Equal(t, filepath.Join(course.Path, "01 doc 2.pdf"), attachments[2].Path)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanner_parseFilename(t *testing.T) {
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
			fb := parseFilename(tt)
			require.Nil(t, fb)
		}
	})

	t.Run("assets", func(t *testing.T) {
		var tests = []struct {
			in       string
			expected *parsedFilename
		}{
			// Video (with varied filenames)
			{"0    file 0.avi", &parsedFilename{prefix: 0, title: "file 0", asset: types.NewAsset("avi")}},
			{"001 file 1.mp4", &parsedFilename{prefix: 1, title: "file 1", asset: types.NewAsset("mp4")}},
			{"1-file.ogg", &parsedFilename{prefix: 1, title: "file", asset: types.NewAsset("ogg")}},
			{"2 - file.webm", &parsedFilename{prefix: 2, title: "file", asset: types.NewAsset("webm")}},
			{"3 -file.m4a", &parsedFilename{prefix: 3, title: "file", asset: types.NewAsset("m4a")}},
			{"4- file.opus", &parsedFilename{prefix: 4, title: "file", asset: types.NewAsset("opus")}},
			{"5000 --- file.wav", &parsedFilename{prefix: 5000, title: "file", asset: types.NewAsset("wav")}},
			{"0100 file.mp3", &parsedFilename{prefix: 100, title: "file", asset: types.NewAsset("mp3")}},
			// PDF
			{"1 - doc.pdf", &parsedFilename{prefix: 1, title: "doc", asset: types.NewAsset("pdf")}},
			// HTML
			{"1 index.html", &parsedFilename{prefix: 1, title: "index", asset: types.NewAsset("html")}},
		}

		for _, tt := range tests {
			fb := parseFilename(tt.in)
			require.Equal(t, tt.expected, fb, fmt.Sprintf("error for [%s]", tt.in))
		}
	})

	t.Run("attachments", func(t *testing.T) {
		var tests = []struct {
			in       string
			expected *parsedFilename
		}{
			// No title
			{"01", &parsedFilename{prefix: 1, title: "01"}},
			{"200.pdf", &parsedFilename{prefix: 200, title: "200.pdf"}},
			{"1 -.txt", &parsedFilename{prefix: 1, title: "1 -.txt"}},
			{"1 .txt", &parsedFilename{prefix: 1, title: "1 .txt"}},
			{"1     .pdf", &parsedFilename{prefix: 1, title: "1     .pdf"}},
			// No extension (fileName should have no prefix)
			{"0    file 0", &parsedFilename{prefix: 0, title: "file 0"}},
			{"001    file 1", &parsedFilename{prefix: 1, title: "file 1"}},
			{"1001 - file", &parsedFilename{prefix: 1001, title: "file"}},
			{"0123-file", &parsedFilename{prefix: 123, title: "file"}},
			{"1 --- file", &parsedFilename{prefix: 1, title: "file"}},
			// Non-asset extension (fileName should have no prefix)
			{"1 file.txt", &parsedFilename{prefix: 1, title: "file.txt"}},
		}

		for _, tt := range tests {
			fb := parseFilename(tt.in)
			require.Equal(t, tt.expected, fb, fmt.Sprintf("error for [%s]", tt.in))
		}
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanner_IsCard(t *testing.T) {
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

func TestScanner_UpdateAssets(t *testing.T) {
	t.Run("nothing added or deleted", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		numAssets := 10
		batchSize := 3

		assets := []*models.Asset{}
		for i := range numAssets {
			asset := &models.Asset{
				CourseID: course.ID,
				Title:    fmt.Sprintf("Asset %d", i+1),
				Prefix:   sql.NullInt16{Int16: int16(i + 1), Valid: true},
				Chapter:  fmt.Sprintf("Chapter %d", i/batchSize+1),
				Type:     *types.NewAsset("mp4"),
				Path:     fmt.Sprintf("/course-1/Chapter %d/%d asset.mp4", i/batchSize+1, (i%batchSize)+1),
				Hash:     security.RandomString(64),
			}
			require.NoError(t, scanner.dao.CreateAsset(ctx, asset))
			assets = append(assets, asset)
		}

		err := updateAssets(ctx, scanner.dao, course.ID, assets)
		require.NoError(t, err)

		count, err := scanner.dao.Count(ctx, &models.Asset{}, nil)
		require.NoError(t, err)
		require.Equal(t, numAssets, count)
	})

	t.Run("add", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		numAssets := 10
		batchSize := 3

		assets := []*models.Asset{}
		for i := range numAssets {
			asset := &models.Asset{
				CourseID: course.ID,
				Title:    fmt.Sprintf("Asset %d", i+1),
				Prefix:   sql.NullInt16{Int16: int16(i + 1), Valid: true},
				Chapter:  fmt.Sprintf("Chapter %d", i/batchSize+1),
				Type:     *types.NewAsset("mp4"),
				Path:     fmt.Sprintf("/course-1/Chapter %d/%d asset.mp4", i/batchSize+1, (i%batchSize)+1),
				Hash:     security.RandomString(64),
			}
			assets = append(assets, asset)
		}

		// Add first 7 assets
		err := updateAssets(ctx, scanner.dao, course.ID, assets[:7])
		require.NoError(t, err)

		count, err := scanner.dao.Count(ctx, &models.Asset{}, nil)
		require.NoError(t, err)
		require.Equal(t, 7, count)

		// Add remaining assets
		err = updateAssets(ctx, scanner.dao, course.ID, assets)
		require.NoError(t, err)

		count, err = scanner.dao.Count(ctx, &models.Asset{}, nil)
		require.NoError(t, err)
		require.Equal(t, numAssets, count)

		// Ensure all assets have an ID
		assetsResult := []*models.Asset{}
		err = scanner.dao.List(ctx, &assetsResult, nil)
		require.NoError(t, err)
	})

	t.Run("delete", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		numAssets := 10
		batchSize := 3

		assets := []*models.Asset{}
		for i := range numAssets {
			asset := &models.Asset{
				CourseID: course.ID,
				Title:    fmt.Sprintf("Asset %d", i+1),
				Prefix:   sql.NullInt16{Int16: int16(i + 1), Valid: true},
				Chapter:  fmt.Sprintf("Chapter %d", i/batchSize+1),
				Type:     *types.NewAsset("mp4"),
				Path:     fmt.Sprintf("/course-1/Chapter %d/%d asset.mp4", i/batchSize+1, (i%batchSize)+1),
				Hash:     security.RandomString(64),
			}
			require.NoError(t, scanner.dao.CreateAsset(ctx, asset))
			assets = append(assets, asset)
		}

		// Delete 2 assets
		err := updateAssets(ctx, scanner.dao, course.ID, assets[2:])
		require.NoError(t, err)

		count, err := scanner.dao.Count(ctx, &models.Asset{}, nil)
		require.NoError(t, err)
		require.Equal(t, 8, count)

		// Delete another 2 assets
		err = updateAssets(ctx, scanner.dao, course.ID, assets[4:])
		require.NoError(t, err)

		count, err = scanner.dao.Count(ctx, &models.Asset{}, nil)
		require.NoError(t, err)
		require.Equal(t, 6, count)
	})

	t.Run("rename", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		numAssets := 10
		batchSize := 3

		assets := []*models.Asset{}
		for i := range numAssets {
			asset := &models.Asset{
				CourseID: course.ID,
				Title:    fmt.Sprintf("Asset %d", i+1),
				Prefix:   sql.NullInt16{Int16: int16(i + 1), Valid: true},
				Chapter:  fmt.Sprintf("Chapter %d", i/batchSize+1),
				Type:     *types.NewAsset("mp4"),
				Path:     fmt.Sprintf("/course-1/Chapter %d/%d asset.mp4", i/batchSize+1, (i%batchSize)+1),
				Hash:     security.RandomString(64),
			}
			assets = append(assets, asset)
		}

		// Add assets
		err := updateAssets(ctx, scanner.dao, course.ID, assets)
		require.NoError(t, err)

		count, err := scanner.dao.Count(ctx, &models.Asset{}, nil)
		require.NoError(t, err)
		require.Equal(t, numAssets, count)

		// Rename 2 assets
		assets[2].Title = "Asset 100"
		assets[2].Prefix = sql.NullInt16{Int16: 100, Valid: true}
		assets[2].Chapter = "Chapter 100"
		assets[2].Path = "/course-1/Chapter 100/100 asset.mp4"

		assets[4].Title = "Asset 200"
		assets[4].Prefix = sql.NullInt16{Int16: 200, Valid: true}
		assets[4].Chapter = "Chapter 200"
		assets[4].Path = "/course-1/Chapter 200/200 asset.mp4"

		err = updateAssets(ctx, scanner.dao, course.ID, assets)
		require.NoError(t, err)

		count, err = scanner.dao.Count(ctx, &models.Asset{}, nil)
		require.NoError(t, err)
		require.Equal(t, numAssets, count)

		// Ensure the assets were updated
		asset2 := &models.Asset{Base: models.Base{ID: assets[2].ID}}
		err = scanner.dao.GetById(ctx, asset2)
		require.NoError(t, err)
		require.Equal(t, "Asset 100", asset2.Title)
		require.Equal(t, int16(100), asset2.Prefix.Int16)
		require.Equal(t, "Chapter 100", asset2.Chapter)
		require.Equal(t, "/course-1/Chapter 100/100 asset.mp4", asset2.Path)

		asset4 := &models.Asset{Base: models.Base{ID: assets[4].ID}}
		err = scanner.dao.GetById(ctx, asset4)
		require.NoError(t, err)
		require.Equal(t, "Asset 200", asset4.Title)
		require.Equal(t, int16(200), asset4.Prefix.Int16)
		require.Equal(t, "Chapter 200", asset4.Chapter)
		require.Equal(t, "/course-1/Chapter 200/200 asset.mp4", asset4.Path)
	})

	t.Run("swap", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		numAssets := 2

		assets := []*models.Asset{}
		for i := range numAssets {
			asset := &models.Asset{
				CourseID: course.ID,
				Title:    fmt.Sprintf("Asset %d", i+1),
				Prefix:   sql.NullInt16{Int16: int16(i + 1), Valid: true},
				Chapter:  fmt.Sprintf("Chapter %d", i+1),
				Type:     *types.NewAsset("mp4"),
				Path:     fmt.Sprintf("/course-1/Chapter %d/%d asset.mp4", i+1, i+1),
				Hash:     security.RandomString(64),
			}

			require.NoError(t, scanner.dao.CreateAsset(ctx, asset))
			assets = append(assets, asset)
		}

		// Swap assets title and path
		assets[0].Title, assets[1].Title = assets[1].Title, assets[0].Title
		assets[0].Path, assets[1].Path = assets[1].Path, assets[0].Path

		err := updateAssets(ctx, scanner.dao, course.ID, assets)
		require.NoError(t, err)

		asset1 := &models.Asset{Base: models.Base{ID: assets[0].ID}}
		err = scanner.dao.GetById(ctx, asset1)
		require.NoError(t, err)
		require.Equal(t, "Asset 2", asset1.Title)
		require.Equal(t, "/course-1/Chapter 2/2 asset.mp4", asset1.Path)

		asset2 := &models.Asset{Base: models.Base{ID: assets[1].ID}}
		err = scanner.dao.GetById(ctx, asset2)
		require.NoError(t, err)
		require.Equal(t, "Asset 1", asset2.Title)
		require.Equal(t, "/course-1/Chapter 1/1 asset.mp4", asset2.Path)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanner_UpdateAttachments(t *testing.T) {
	t.Run("nothing added or deleted", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/Chapter 1/1 Asset 1.mp4",
			Hash:     security.RandomString(64),
		}
		require.NoError(t, scanner.dao.CreateAsset(ctx, asset))

		numAttachments := 10

		attachments := []*models.Attachment{}
		for i := range numAttachments {
			attachment := &models.Attachment{
				CourseID: course.ID,
				AssetID:  asset.ID,
				Title:    fmt.Sprintf("Attachment %d", i+1),
				Path:     fmt.Sprintf("/course-1/Chapter 1/1 Attachment %d.pdf", i+1),
			}
			require.NoError(t, scanner.dao.CreateAttachment(ctx, attachment))
			attachments = append(attachments, attachment)
		}

		err := updateAttachments(ctx, scanner.dao, course.ID, attachments)
		require.NoError(t, err)

		count, err := scanner.dao.Count(ctx, &models.Attachment{}, nil)
		require.NoError(t, err)
		require.Equal(t, numAttachments, count)
	})

	t.Run("add", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/Chapter 1/1 Asset 1.mp4",
			Hash:     security.RandomString(64),
		}
		require.NoError(t, scanner.dao.CreateAsset(ctx, asset))

		numAttachments := 10

		attachments := []*models.Attachment{}
		for i := range numAttachments {
			attachment := &models.Attachment{
				CourseID: course.ID,
				AssetID:  asset.ID,
				Title:    fmt.Sprintf("Attachment %d", i+1),
				Path:     fmt.Sprintf("/course-1/Chapter 1/1 Attachment %d.pdf", i+1),
			}
			attachments = append(attachments, attachment)
		}

		// Add first 7 attachments
		err := updateAttachments(ctx, scanner.dao, course.ID, attachments[:7])
		require.NoError(t, err)

		count, err := scanner.dao.Count(ctx, &models.Attachment{}, nil)
		require.NoError(t, err)
		require.Equal(t, 7, count)

		// Add remaining attachments
		err = updateAttachments(ctx, scanner.dao, course.ID, attachments)
		require.NoError(t, err)

		count, err = scanner.dao.Count(ctx, &models.Attachment{}, nil)
		require.NoError(t, err)
		require.Equal(t, numAttachments, count)
	})

	t.Run("delete", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/Chapter 1/1 Asset 1.mp4",
			Hash:     security.RandomString(64),
		}
		require.NoError(t, scanner.dao.CreateAsset(ctx, asset))

		numAttachments := 10

		attachments := []*models.Attachment{}
		for i := range numAttachments {
			attachment := &models.Attachment{
				CourseID: course.ID,
				AssetID:  asset.ID,
				Title:    fmt.Sprintf("Attachment %d", i+1),
				Path:     fmt.Sprintf("/course-1/Chapter 1/1 Attachment %d.pdf", i+1),
			}
			require.NoError(t, scanner.dao.CreateAttachment(ctx, attachment))
			attachments = append(attachments, attachment)
		}

		// Delete 2 attachments
		err := updateAttachments(ctx, scanner.dao, course.ID, attachments[2:])
		require.NoError(t, err)

		count, err := scanner.dao.Count(ctx, &models.Attachment{}, nil)
		require.NoError(t, err)
		require.Equal(t, 8, count)

		// Delete another 2 attachments
		err = updateAttachments(ctx, scanner.dao, course.ID, attachments[4:])
		require.NoError(t, err)

		count, err = scanner.dao.Count(ctx, &models.Attachment{}, nil)
		require.NoError(t, err)
		require.Equal(t, 6, count)
	})
}
