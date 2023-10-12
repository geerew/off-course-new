package jobs

// import (
// 	"errors"
// 	"fmt"
// 	"os"
// 	"sort"
// 	"strconv"
// 	"testing"
// 	"time"

// 	"github.com/geerew/off-course/database"
// 	"github.com/geerew/off-course/models"
// 	"github.com/geerew/off-course/utils/appFs"
// 	"github.com/geerew/off-course/utils/types"
// 	"github.com/rs/zerolog"
// 	"github.com/rs/zerolog/log"
// 	"github.com/rzajac/zltest"
// 	"github.com/spf13/afero"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func Test_Add(t *testing.T) {
// 	t.Run("success", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		// Create a course
// 		course := &models.Course{ID: "course1", Title: "Course 1", Path: "/course1"}
// 		require.Nil(t, scanner.db.DB().Create(course).Error)

// 		scan, err := scanner.Add(course.ID)
// 		require.Nil(t, err)
// 		assert.Equal(t, scan.CourseID, course.ID)
// 	})

// 	t.Run("duplicate", func(t *testing.T) {
// 		scanner, lh, teardown := setup(t)
// 		defer teardown(t)

// 		// Create a course
// 		course := &models.Course{ID: "course1", Title: "Course 1", Path: "/course1"}
// 		require.Nil(t, scanner.db.DB().Create(course).Error)

// 		scan, err := scanner.Add(course.ID)
// 		assert.Nil(t, err)
// 		assert.Equal(t, scan.CourseID, course.ID)

// 		// Add the same course again
// 		scan, err = scanner.Add(course.ID)
// 		require.Nil(t, err)
// 		require.NotNil(t, lh.LastEntry())
// 		require.Nil(t, scan)
// 		lh.LastEntry().ExpMsg("scan job already exists")
// 		lh.LastEntry().ExpLevel(zerolog.DebugLevel)
// 	})

// 	t.Run("invalid course", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		scan, err := scanner.Add("bob")
// 		require.EqualError(t, err, "no rows in result set")
// 		assert.Nil(t, scan)

// 		scans, err := models.GetScans(scanner.db, nil)
// 		require.Nil(t, err)
// 		require.Len(t, scans, 0)
// 	})

// 	t.Run("not blocked", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		// Create 2 scans
// 		course1 := &models.Course{ID: "course1", Title: "Course 1", Path: "/course 1"}
// 		course2 := &models.Course{ID: "course2", Title: "Course 2", Path: "/course 2"}
// 		require.Nil(t, scanner.db.DB().Create(course1).Error)
// 		require.Nil(t, scanner.db.DB().Create(course2).Error)

// 		scan1, err := scanner.Add(course1.ID)
// 		require.Nil(t, err)
// 		assert.Equal(t, scan1.CourseID, course1.ID)

// 		scan2, err := scanner.Add(course2.ID)
// 		require.Nil(t, err)
// 		assert.Equal(t, scan2.CourseID, course2.ID)
// 	})

// 	t.Run("scan lookup error", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		err := scanner.db.DB().Migrator().DropTable(&models.Scan{})
// 		require.Nil(t, err)

// 		scan, err := scanner.Add("")
// 		require.EqualError(t, err, "no such table: scans")
// 		assert.Nil(t, scan)
// 	})
// }

// func Test_Worker(t *testing.T) {
// 	t.Run("single job", func(t *testing.T) {
// 		scanner, lh, teardown := setup(t)
// 		defer teardown(t)

// 		// Create a course
// 		course := &models.Course{ID: "course1", Title: "Course 1", Path: "/course1"}
// 		require.Nil(t, scanner.db.DB().Create(course).Error)

// 		// Start the worker
// 		go scanner.Worker(func(scan *models.Scan, db database.Database, fs *appFs.AppFs) error {
// 			time.Sleep(time.Millisecond * 1)
// 			return nil
// 		})

// 		scan, err := scanner.Add(course.ID)
// 		require.Nil(t, err)
// 		assert.Equal(t, scan.CourseID, course.ID)

// 		// Give time for the worker to finish
// 		time.Sleep(time.Millisecond * 5)

// 		// Assert the scan job was deleted from the DB
// 		count, err := models.CountScans(scanner.db, nil)
// 		require.Nil(t, err)
// 		assert.Equal(t, count, int64(0))

// 		// Validate the logs
// 		require.NotNil(t, lh.LastEntry())
// 		lh.Entries().Get()[lh.Len()-2].ExpMsg("finished processing scan job")
// 		lh.Entries().Get()[lh.Len()-2].ExpStr("path", course.Path)
// 		lh.LastEntry().ExpLevel(zerolog.InfoLevel)
// 		lh.LastEntry().ExpMsg("finished processing all scan jobs")
// 	})

// 	t.Run("several jobs", func(t *testing.T) {
// 		scanner, lh, teardown := setup(t)
// 		defer teardown(t)

// 		// Create 5 courses and add scan jobs
// 		numCourses := 3
// 		courses := []*models.Course{}
// 		for i := 0; i < numCourses; i++ {
// 			s := strconv.Itoa(i)
// 			course := &models.Course{ID: "id" + s, Title: "Course " + s, Path: "/course " + s}
// 			require.Nil(t, scanner.db.DB().Create(course).Error)
// 			courses = append(courses, course)

// 			scan, err := scanner.Add(course.ID)
// 			require.Nil(t, err)
// 			require.Equal(t, scan.CourseID, course.ID)
// 		}

// 		// Start the worker
// 		go scanner.Worker(func(scan *models.Scan, db database.Database, fs *appFs.AppFs) error {
// 			time.Sleep(time.Millisecond * 1)
// 			return nil
// 		})

// 		// Give time for the worker to finish
// 		time.Sleep(time.Millisecond * 10)

// 		// Assert the scan job was deleted from the DB
// 		count, err := models.CountScans(scanner.db, nil)
// 		require.Nil(t, err)
// 		assert.Equal(t, count, int64(0))

// 		// Validate the logs. This will process the 3 courses, with the last being course 2
// 		require.NotNil(t, lh.LastEntry())
// 		lh.Entries().Get()[lh.Len()-2].ExpMsg("finished processing scan job")
// 		lh.Entries().Get()[lh.Len()-2].ExpStr("path", courses[2].Path)
// 		lh.LastEntry().ExpLevel(zerolog.InfoLevel)
// 		lh.LastEntry().ExpMsg("finished processing all scan jobs")
// 	})

// 	t.Run("error processing", func(t *testing.T) {
// 		scanner, lh, teardown := setup(t)
// 		defer teardown(t)

// 		// Create a course
// 		course := &models.Course{ID: "course1", Title: "Course 1", Path: "/course1"}
// 		require.Nil(t, scanner.db.DB().Create(course).Error)

// 		// Start the worker
// 		go scanner.Worker(func(scan *models.Scan, db database.Database, fs *appFs.AppFs) error {
// 			return errors.New("processing error")
// 		})

// 		scan, err := scanner.Add(course.ID)
// 		require.Nil(t, err)
// 		assert.Equal(t, scan.CourseID, course.ID)

// 		// Give time for the worker to finish
// 		time.Sleep(time.Millisecond * 5)

// 		// Validate the logs
// 		require.NotNil(t, lh.LastEntry())
// 		lh.LastEntry().ExpMsg("error processing scan job")
// 		lh.LastEntry().ExpErr(errors.New("processing error"))
// 		lh.LastEntry().ExpLevel(zerolog.ErrorLevel)
// 	})

// 	t.Run("scan error", func(t *testing.T) {
// 		scanner, lh, teardown := setup(t)
// 		defer teardown(t)

// 		// Create a course
// 		course := &models.Course{ID: "course1", Title: "Course 1", Path: "/course1"}
// 		require.Nil(t, scanner.db.DB().Create(course).Error)

// 		// Add the job
// 		scan, err := scanner.Add(course.ID)
// 		require.Nil(t, err)
// 		assert.Equal(t, scan.CourseID, course.ID)

// 		// Drop the DB
// 		require.Nil(t, scanner.db.DB().Migrator().DropTable(&models.Scan{}))

// 		// Start the worker
// 		go scanner.Worker(func(scan *models.Scan, db database.Database, fs *appFs.AppFs) error {
// 			return nil
// 		})

// 		// Give time for the worker to finish
// 		time.Sleep(time.Millisecond * 5)

// 		// Validate the logs
// 		require.NotNil(t, lh.LastEntry())
// 		lh.LastEntry().ExpMsg("error looking up next scan job")
// 		lh.LastEntry().ExpLevel(zerolog.ErrorLevel)
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func Test_ScanCourse(t *testing.T) {
// 	t.Run("nil", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		err := scanCourse(nil, scanner.db, scanner.appFs)
// 		require.EqualError(t, err, "course cannot be empty")
// 	})

// 	t.Run("path error", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		course := models.CreateTestCourses(t, scanner.db, 1)[0]

// 		err := scanCourse(course, scanner.db, scanner.appFs)
// 		require.EqualError(t, err, "unable to open path")
// 	})

// 	t.Run("found card", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		// Create course
// 		course := models.CreateTestCourses(t, scanner.db, 1)[0]
// 		require.Empty(t, course.CardPath)

// 		// Create card
// 		cardPath := fmt.Sprintf("%s/card.jpg", course.Path)
// 		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
// 		scanner.appFs.Fs.Create(cardPath)

// 		err := scanCourse(course, scanner.db, scanner.appFs)
// 		require.Nil(t, err)

// 		// Assert the course card
// 		foundCourse, err := models.GetCourse(scanner.db, course.ID, nil)
// 		require.Nil(t, err)
// 		assert.Equal(t, cardPath, foundCourse.CardPath)
// 		assert.Equal(t, cardPath, course.CardPath)
// 	})

// 	t.Run("ignore card in chapter", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		// Create course
// 		course := models.CreateTestCourses(t, scanner.db, 1)[0]
// 		require.Empty(t, course.CardPath)

// 		// Create a card within a chapter
// 		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
// 		scanner.appFs.Fs.Create(fmt.Sprintf("%s/chapter 1/card.jpg", course.Path))

// 		err := scanCourse(course, scanner.db, scanner.appFs)
// 		require.Nil(t, err)

// 		// Assert the course card
// 		foundCourse, err := models.GetCourse(scanner.db, course.ID, nil)
// 		require.Nil(t, err)
// 		assert.Empty(t, foundCourse.CardPath)
// 		assert.Empty(t, course.CardPath)
// 	})

// 	t.Run("multiple cards", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		// Create course
// 		course := models.CreateTestCourses(t, scanner.db, 1)[0]
// 		require.Empty(t, course.CardPath)

// 		// Create card
// 		cardPath1 := fmt.Sprintf("%s/card.jpg", course.Path)
// 		cardPath2 := fmt.Sprintf("%s/card.png", course.Path)
// 		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
// 		scanner.appFs.Fs.Create(cardPath1)
// 		scanner.appFs.Fs.Create(cardPath2)

// 		err := scanCourse(course, scanner.db, scanner.appFs)
// 		require.Nil(t, err)

// 		// Assert the course card is the first one
// 		foundCourse, err := models.GetCourse(scanner.db, course.ID, nil)
// 		require.Nil(t, err)
// 		assert.Equal(t, cardPath1, foundCourse.CardPath)
// 		assert.Equal(t, cardPath1, course.CardPath)
// 	})

// 	t.Run("ignore files", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		course := models.CreateTestCourses(t, scanner.db, 1)[0]

// 		// Create a video file that will be ignored due to no prefix or extension
// 		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
// 		scanner.appFs.Fs.Create(fmt.Sprintf("%s/video 1", course.Path))

// 		err := scanCourse(course, scanner.db, scanner.appFs)
// 		require.Nil(t, err)

// 		// Assert the assets
// 		assets, err := models.GetAssets(scanner.db, nil)
// 		require.Nil(t, err)
// 		require.Len(t, assets, 0)
// 	})

// 	t.Run("found video", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		course := models.CreateTestCourses(t, scanner.db, 1)[0]

// 		// Create a video asset
// 		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
// 		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 - video 1.mkv", course.Path))

// 		err := scanCourse(course, scanner.db, scanner.appFs)
// 		require.Nil(t, err)

// 		// Assert the assets
// 		assets, err := models.GetAssets(scanner.db, nil)
// 		require.Nil(t, err)
// 		require.Len(t, assets, 1)

// 		assert.Equal(t, "video 1", assets[0].Title)
// 		assert.Equal(t, course.ID, assets[0].CourseID)
// 		assert.Equal(t, 1, assets[0].Prefix)
// 		assert.Empty(t, assets[0].Chapter)
// 		assert.True(t, assets[0].Type.IsVideo())
// 	})

// 	t.Run("found document", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		course := models.CreateTestCourses(t, scanner.db, 1)[0]

// 		// Create a document asset
// 		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
// 		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 - doc 1.html", course.Path))

// 		err := scanCourse(course, scanner.db, scanner.appFs)
// 		require.Nil(t, err)

// 		// Assert the assets
// 		assets, err := models.GetAssets(scanner.db, nil)
// 		require.Nil(t, err)
// 		require.Len(t, assets, 1)

// 		assert.Equal(t, "doc 1", assets[0].Title)
// 		assert.Equal(t, course.ID, assets[0].CourseID)
// 		assert.Equal(t, 1, assets[0].Prefix)
// 		assert.Empty(t, assets[0].Chapter)
// 		assert.True(t, assets[0].Type.IsHTML())
// 	})

// 	t.Run("found videos and documents", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		course := models.CreateTestCourses(t, scanner.db, 1)[0]

// 		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
// 		scanner.appFs.Fs.Mkdir(fmt.Sprintf("%s/chapter 1", course.Path), os.ModePerm)

// 		// Add 3 video files (2 in root and 1 in chapter) and 1 document
// 		//
// 		// `file2` should be ignore because it shares the same prefix as  `file1`
// 		file1 := fmt.Sprintf("%s/01 - video 1.mp4", course.Path)
// 		file2 := fmt.Sprintf("%s/01 - video 2.avi", course.Path)
// 		file3 := fmt.Sprintf("%s/02 - doc 1.html", course.Path)
// 		file4 := fmt.Sprintf("%s/chapter 1/01 - video 1.mkv", course.Path)
// 		scanner.appFs.Fs.Create(file1)
// 		scanner.appFs.Fs.Create(file2)
// 		scanner.appFs.Fs.Create(file3)
// 		scanner.appFs.Fs.Create(file4)

// 		err := scanCourse(course, scanner.db, scanner.appFs)
// 		require.Nil(t, err)

// 		// Assert the assets
// 		assets, err := models.GetAssets(scanner.db, nil)
// 		require.Nil(t, err)
// 		require.Len(t, assets, 3)

// 		// Sort assets to ease testing
// 		sort.Slice(assets, func(i, j int) bool {
// 			if assets[i].Chapter != assets[j].Chapter {
// 				return assets[i].Chapter < assets[j].Chapter
// 			}
// 			return assets[i].Prefix < assets[j].Prefix
// 		})

// 		// Assert asset 1
// 		assert.Equal(t, "video 1", assets[0].Title)
// 		assert.Equal(t, course.ID, assets[0].CourseID)
// 		assert.Equal(t, 1, assets[0].Prefix)
// 		assert.Empty(t, assets[0].Chapter)
// 		assert.True(t, assets[0].Type.IsVideo())

// 		// Assert assert 2
// 		assert.Equal(t, "doc 1", assets[1].Title)
// 		assert.Equal(t, course.ID, assets[1].CourseID)
// 		assert.Equal(t, 2, assets[1].Prefix)
// 		assert.Empty(t, assets[1].Chapter)
// 		assert.True(t, assets[1].Type.IsHTML())

// 		// Assert assert 3
// 		assert.Equal(t, "video 1", assets[2].Title)
// 		assert.Equal(t, course.ID, assets[2].CourseID)
// 		assert.Equal(t, 1, assets[2].Prefix)
// 		assert.Equal(t, "chapter 1", assets[2].Chapter)
// 		assert.True(t, assets[2].Type.IsVideo())
// 	})

// 	t.Run("video trumps all", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		course := models.CreateTestCourses(t, scanner.db, 1)[0]

// 		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)

// 		// Create an pdf, html and video file with the same prefix
// 		file1 := fmt.Sprintf("%s/01 - a.pdf", course.Path)
// 		file2 := fmt.Sprintf("%s/01 - b.html", course.Path)
// 		file3 := fmt.Sprintf("%s/01 - c.mp4", course.Path)
// 		scanner.appFs.Fs.Create(file1)
// 		scanner.appFs.Fs.Create(file2)
// 		scanner.appFs.Fs.Create(file3)

// 		err := scanCourse(course, scanner.db, scanner.appFs)
// 		require.Nil(t, err)

// 		// Assert the assets
// 		assets, err := models.GetAssets(scanner.db, nil)
// 		require.Nil(t, err)
// 		require.Len(t, assets, 1)

// 		// Assert the video file is the asset
// 		assert.Equal(t, "c", assets[0].Title)
// 		assert.Equal(t, course.ID, assets[0].CourseID)
// 		assert.Equal(t, 1, assets[0].Prefix)
// 		assert.Equal(t, file3, assets[0].Path)
// 		assert.True(t, assets[0].Type.IsVideo())

// 		// TODO: Asserts the pdf and html attachments
// 	})

// 	t.Run("html trumps pdf", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		course := models.CreateTestCourses(t, scanner.db, 1)[0]

// 		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)

// 		// Create a pdf and html file with the same prefix
// 		file1 := fmt.Sprintf("%s/01 - a.pdf", course.Path)
// 		file2 := fmt.Sprintf("%s/01 - b.html", course.Path)
// 		scanner.appFs.Fs.Create(file1)
// 		scanner.appFs.Fs.Create(file2)

// 		err := scanCourse(course, scanner.db, scanner.appFs)
// 		require.Nil(t, err)

// 		// Assert the assets
// 		assets, err := models.GetAssets(scanner.db, nil)
// 		require.Nil(t, err)
// 		require.Len(t, assets, 1)

// 		// Assert the html file is the asset
// 		assert.Equal(t, "b", assets[0].Title)
// 		assert.Equal(t, course.ID, assets[0].CourseID)
// 		assert.Equal(t, 1, assets[0].Prefix)
// 		assert.Equal(t, file2, assets[0].Path)
// 		assert.True(t, assets[0].Type.IsHTML())

// 		// TODO: Asserts the pdf attachment
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func Test_ExtractPrefixAndTitle(t *testing.T) {
// 	t.Run("invalid assets", func(t *testing.T) {
// 		var tests = []string{
// 			"file",          // No prefix
// 			"file.file",     // No prefix
// 			"file.avi",      // No prefix
// 			"a - file.avi",  // Invalid prefix
// 			"1 - .avi",      // No title
// 			"1 .avi",        // No title
// 			"1      .avi",   // No title
// 			"1 --  -- .avi", // No title
// 		}

// 		for _, tt := range tests {
// 			fb := buildFileInfo(tt)
// 			assert.Nil(t, fb)
// 		}
// 	})

// 	t.Run("valid assets", func(t *testing.T) {
// 		var tests = []struct {
// 			in       string
// 			expected *FileInfo
// 		}{
// 			// Various valid formats of prefix/title/extension
// 			{"0    file 0.avi", &FileInfo{prefix: 0, title: "file 0", ext: "avi", assetType: *types.NewAsset("avi"), isAsset: true}},
// 			{"001 file 1.avi", &FileInfo{prefix: 1, title: "file 1", ext: "avi", assetType: *types.NewAsset("avi"), isAsset: true}},
// 			{"1 - file.avi", &FileInfo{prefix: 1, title: "file", ext: "avi", assetType: *types.NewAsset("avi"), isAsset: true}},
// 			{"1-file.avi", &FileInfo{prefix: 1, title: "file", ext: "avi", assetType: *types.NewAsset("avi"), isAsset: true}},
// 			{"1 --- file.avi", &FileInfo{prefix: 1, title: "file", ext: "avi", assetType: *types.NewAsset("avi"), isAsset: true}},
// 			{"1 file.avi", &FileInfo{prefix: 1, title: "file", ext: "avi", assetType: *types.NewAsset("avi"), isAsset: true}},
// 			// Videos
// 			{"1 - video.avi", &FileInfo{prefix: 1, title: "video", ext: "avi", assetType: *types.NewAsset("avi"), isAsset: true}},
// 			{"1 - video.mkv", &FileInfo{prefix: 1, title: "video", ext: "mkv", assetType: *types.NewAsset("mkv"), isAsset: true}},
// 			{"1 - video.flac", &FileInfo{prefix: 1, title: "video", ext: "flac", assetType: *types.NewAsset("flac"), isAsset: true}},
// 			{"1 - video.mp4", &FileInfo{prefix: 1, title: "video", ext: "mp4", assetType: *types.NewAsset("mp4"), isAsset: true}},
// 			{"1 - video.m4a", &FileInfo{prefix: 1, title: "video", ext: "m4a", assetType: *types.NewAsset("m4a"), isAsset: true}},
// 			{"1 - video.mp3", &FileInfo{prefix: 1, title: "video", ext: "mp3", assetType: *types.NewAsset("mp3"), isAsset: true}},
// 			{"1 - video.ogv", &FileInfo{prefix: 1, title: "video", ext: "ogv", assetType: *types.NewAsset("ogv"), isAsset: true}},
// 			{"1 - video.ogm", &FileInfo{prefix: 1, title: "video", ext: "ogm", assetType: *types.NewAsset("ogm"), isAsset: true}},
// 			{"1 - video.ogg", &FileInfo{prefix: 1, title: "video", ext: "ogg", assetType: *types.NewAsset("ogg"), isAsset: true}},
// 			{"1 - video.oga", &FileInfo{prefix: 1, title: "video", ext: "oga", assetType: *types.NewAsset("oga"), isAsset: true}},
// 			{"1 - video.opus", &FileInfo{prefix: 1, title: "video", ext: "opus", assetType: *types.NewAsset("opus"), isAsset: true}},
// 			{"1 - video.webm", &FileInfo{prefix: 1, title: "video", ext: "webm", assetType: *types.NewAsset("webm"), isAsset: true}},
// 			{"1 - video.wav", &FileInfo{prefix: 1, title: "video", ext: "wav", assetType: *types.NewAsset("wav"), isAsset: true}},
// 			// Document
// 			{"1 - doc.html", &FileInfo{prefix: 1, title: "doc", ext: "html", assetType: *types.NewAsset("html"), isAsset: true}},
// 			{"1 - doc.htm", &FileInfo{prefix: 1, title: "doc", ext: "htm", assetType: *types.NewAsset("htm"), isAsset: true}},
// 			// Images
// 			// {"1 - image.jpg", &FileInfo{prefix: "1", title: "image", Extension: MediaTypeJpg, Category: MediaCategoryImage}},
// 			// {"1 - image.jpeg", &FileInfo{prefix: "1", title: "image", Extension: MediaTypeJpeg, Category: MediaCategoryImage}},
// 			// {"1 - image.png", &FileInfo{prefix: "1", title: "image", Extension: MediaTypePng, Category: MediaCategoryImage}},
// 			// {"1 - image.webp", &FileInfo{prefix: "1", title: "image", Extension: MediaTypeWebp, Category: MediaCategoryImage}},
// 			// {"1 - image.tiff", &FileInfo{prefix: "1", title: "image", Extension: MediaTypeTiff, Category: MediaCategoryImage}},
// 		}

// 		for _, tt := range tests {
// 			fb := buildFileInfo(tt.in)
// 			assert.Equal(t, tt.expected, fb, fmt.Sprintf("error for [%s]", tt.in))
// 		}
// 	})

// 	t.Run("valid attachments", func(t *testing.T) {
// 		var tests = []struct {
// 			in       string
// 			expected *FileInfo
// 		}{
// 			// Various valid formats of prefix/title/extension
// 			{"0    file 0", &FileInfo{prefix: 0, title: "file 0", isAsset: false}},
// 			{"001    file 1", &FileInfo{prefix: 1, title: "file 1", isAsset: false}},
// 			{"1 - file", &FileInfo{prefix: 1, title: "file", isAsset: false}},
// 			{"1-file", &FileInfo{prefix: 1, title: "file", isAsset: false}},
// 			{"1 --- file", &FileInfo{prefix: 1, title: "file", isAsset: false}},
// 			{"1 file.txt", &FileInfo{prefix: 1, title: "file", ext: "txt", isAsset: false}},
// 		}

// 		for _, tt := range tests {
// 			fb := buildFileInfo(tt.in)
// 			assert.Equal(t, tt.expected, fb, fmt.Sprintf("error for [%s]", tt.in))
// 		}
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func Test_IsCard(t *testing.T) {
// 	t.Run("invalid", func(t *testing.T) {
// 		var tests = []string{
// 			"card",
// 			"1234",
// 			"1234.jpg",
// 			"jpg",
// 			"card.test.jpg",
// 			"card.txt",
// 		}

// 		for _, tt := range tests {
// 			assert.False(t, isCard(tt))
// 		}
// 	})

// 	t.Run("valid", func(t *testing.T) {
// 		var tests = []string{
// 			"card.jpg",
// 			"card.jpeg",
// 			"card.png",
// 			"card.webp",
// 			"card.tiff",
// 		}

// 		for _, tt := range tests {
// 			assert.True(t, isCard(tt))
// 		}
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func Test_UpdateAssets(t *testing.T) {
// 	t.Run("error (course lookup)", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		err := scanner.db.DB().Migrator().DropTable(&models.Asset{})
// 		require.Nil(t, err)

// 		err = updateAssets(scanner.db, "1234", []*models.Asset{})
// 		require.EqualError(t, err, "no such table: assets")
// 	})

// 	t.Run("success (nothing added or deleted)", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		// Create and insert 1 course with 10 assets
// 		courses := models.CreateTestCourses(t, scanner.db, 1)
// 		assets := models.CreateTestAssets(t, scanner.db, courses, 10)

// 		// Assert there are no assets
// 		count, err := models.CountAssets(scanner.db, &database.DatabaseParams{Where: database.Where{"course_id": courses[0].ID}})
// 		require.Nil(t, err)
// 		assert.Equal(t, int64(10), count)

// 		err = updateAssets(scanner.db, courses[0].ID, assets)
// 		require.Nil(t, err)

// 		// Assert there are still 10 assets
// 		count, err = models.CountAssets(scanner.db, &database.DatabaseParams{Where: database.Where{"course_id": courses[0].ID}})
// 		require.Nil(t, err)
// 		assert.Equal(t, int64(10), count)
// 	})

// 	t.Run("success (add)", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		// Create and insert 1 course
// 		courses := models.CreateTestCourses(t, scanner.db, 1)

// 		// Create 10 assets for this course (they are not insert into the DB)
// 		assets := models.CreateTestAssets(t, nil, courses, 10)

// 		// Assert there are no assets
// 		count, err := models.CountAssets(scanner.db, &database.DatabaseParams{Where: database.Where{"course_id": courses[0].ID}})
// 		require.Nil(t, err)
// 		assert.Equal(t, int64(0), count)

// 		err = updateAssets(scanner.db, courses[0].ID, assets)
// 		require.Nil(t, err)

// 		count, err = models.CountAssets(scanner.db, &database.DatabaseParams{Where: database.Where{"course_id": courses[0].ID}})
// 		require.Nil(t, err)
// 		assert.Equal(t, int64(10), count)

// 		// Add another 2
// 		assetsNew := models.CreateTestAssets(t, nil, courses, 2)
// 		assets = append(assets, assetsNew...)

// 		err = updateAssets(scanner.db, courses[0].ID, assets)
// 		require.Nil(t, err)

// 		count, err = models.CountAssets(scanner.db, &database.DatabaseParams{Where: database.Where{"course_id": courses[0].ID}})
// 		require.Nil(t, err)
// 		assert.Equal(t, int64(12), count)
// 	})

// 	t.Run("success (delete)", func(t *testing.T) {
// 		scanner, _, teardown := setup(t)
// 		defer teardown(t)

// 		// Create and insert 1 course with 12 asset
// 		courses := models.CreateTestCourses(t, scanner.db, 1)
// 		assets := models.CreateTestAssets(t, scanner.db, courses, 12)

// 		// Assert there are no assets
// 		count, err := models.CountAssets(scanner.db, &database.DatabaseParams{Where: database.Where{"course_id": courses[0].ID}})
// 		require.Nil(t, err)
// 		assert.Equal(t, int64(12), count)

// 		// Remove the first 2 assets
// 		assets = assets[2:]

// 		err = updateAssets(scanner.db, courses[0].ID, assets)
// 		require.Nil(t, err)

// 		count, err = models.CountAssets(scanner.db, &database.DatabaseParams{Where: database.Where{"course_id": courses[0].ID}})
// 		require.Nil(t, err)
// 		assert.Equal(t, int64(10), count)

// 		// Remove another 2
// 		assets = assets[2:]

// 		err = updateAssets(scanner.db, courses[0].ID, assets)
// 		require.Nil(t, err)

// 		count, err = models.CountAssets(scanner.db, &database.DatabaseParams{Where: database.Where{"course_id": courses[0].ID}})
// 		require.Nil(t, err)
// 		assert.Equal(t, int64(8), count)
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// // HELPERS
// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func setup(t *testing.T) (*CourseScanner, *zltest.Tester, func(t *testing.T)) {

// 	// Set test logger
// 	loggerHook := zltest.New(t)
// 	log.Logger = zerolog.New(loggerHook).Level(zerolog.DebugLevel)

// 	appFs := appFs.NewAppFs(afero.NewMemMapFs())

// 	db := database.NewSqliteDB(&database.SqliteDbConfig{
// 		IsDebug: false,
// 		DataDir: "./co_data",
// 		AppFs:   appFs,
// 	})

// 	// Force DB to be in-memory
// 	os.Setenv("OC_InMemDb", "true")

// 	require.Nil(t, db.Bootstrap())
// 	require.Nil(t, models.MigrateModels(db))

// 	courseScanner := NewCourseScanner(&CourseScannerConfig{
// 		Db:    db,
// 		AppFs: appFs,
// 	})

// 	// teardown
// 	return courseScanner, loggerHook, func(t *testing.T) {
// 		os.Unsetenv("OC_InMemDb")
// 	}
// }
