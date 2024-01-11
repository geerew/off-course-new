package jobs

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Add(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]

		scan, err := scanner.Add(course.ID)
		require.Nil(t, err)
		assert.Equal(t, scan.CourseID, course.ID)
	})

	t.Run("duplicate", func(t *testing.T) {
		scanner, lh, teardown := setupCourseScanner(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]

		scan, err := scanner.Add(course.ID)
		assert.Nil(t, err)
		assert.Equal(t, scan.CourseID, course.ID)

		// Add the same course again
		scan, err = scanner.Add(course.ID)
		require.Nil(t, err)
		require.NotNil(t, lh.LastEntry())
		require.Nil(t, scan)
		lh.LastEntry().ExpMsg("scan already in progress")
		lh.LastEntry().ExpLevel(zerolog.DebugLevel)
	})

	t.Run("invalid course", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		scan, err := scanner.Add("test")
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, scan)
	})

	t.Run("not blocked", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		courses := models.NewTestCourses(t, scanner.db, 2)

		scan1, err := scanner.Add(courses[0].ID)
		require.Nil(t, err)
		assert.Equal(t, scan1.CourseID, courses[0].ID)

		scan2, err := scanner.Add(courses[1].ID)
		require.Nil(t, err)
		assert.Equal(t, scan2.CourseID, courses[1].ID)
	})

	t.Run("db error", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]

		_, err := scanner.db.Exec("DROP TABLE IF EXISTS " + models.TableScans())
		require.Nil(t, err)

		scan, err := scanner.Add(course.ID)
		require.ErrorContains(t, err, fmt.Sprintf("no such table: %s", models.TableScans()))
		assert.Nil(t, scan)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Worker(t *testing.T) {
	t.Run("single job", func(t *testing.T) {
		scanner, lh, teardown := setupCourseScanner(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]

		// Start the worker
		go scanner.Worker(func(*CourseScanner, *models.Scan) error {
			return nil
		})

		// Add the job
		scan, err := scanner.Add(course.ID)
		require.Nil(t, err)
		assert.Equal(t, scan.CourseID, course.ID)

		// Give time for the worker to finish
		time.Sleep(time.Millisecond * 3)

		// Assert the scan job was deleted from the DB
		s, err := models.GetScan(scanner.db, course.ID)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, s)

		// Validate the logs
		require.NotNil(t, lh.LastEntry())
		lh.Entries().Get()[lh.Len()-2].ExpMsg("finished processing scan job")
		lh.Entries().Get()[lh.Len()-2].ExpStr("path", course.Path)
		lh.LastEntry().ExpLevel(zerolog.InfoLevel)
		lh.LastEntry().ExpMsg("finished processing all scan jobs")
	})

	t.Run("several jobs", func(t *testing.T) {
		scanner, lh, teardown := setupCourseScanner(t)
		defer teardown(t)

		courses := models.NewTestCourses(t, scanner.db, 3)

		for _, course := range courses {
			_, err := scanner.Add(course.ID)
			require.Nil(t, err)
		}

		// Start the worker
		go scanner.Worker(func(*CourseScanner, *models.Scan) error {
			return nil
		})

		// Give time for the worker to finish
		time.Sleep(time.Millisecond * 5)

		// Assert the scan job was deleted from the DB
		for _, course := range courses {
			s, err := models.GetScan(scanner.db, course.ID)
			require.ErrorIs(t, err, sql.ErrNoRows)
			assert.Nil(t, s)
		}

		// Validate the logs.
		require.NotNil(t, lh.LastEntry())
		lh.Entries().Get()[lh.Len()-2].ExpMsg("finished processing scan job")
		lh.Entries().Get()[lh.Len()-2].ExpStr("path", courses[2].Path)
		lh.LastEntry().ExpLevel(zerolog.InfoLevel)
		lh.LastEntry().ExpMsg("finished processing all scan jobs")
	})

	t.Run("error processing", func(t *testing.T) {
		scanner, lh, teardown := setupCourseScanner(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]

		// Start the worker
		go scanner.Worker(func(*CourseScanner, *models.Scan) error {
			return errors.New("processing error")
		})

		scan, err := scanner.Add(course.ID)
		require.Nil(t, err)
		assert.Equal(t, scan.CourseID, course.ID)

		// Give time for the worker to finish
		time.Sleep(time.Millisecond * 3)

		// Validate the logs
		require.NotNil(t, lh.LastEntry())
		lh.Entries().Get()[lh.Len()-2].ExpMsg("error processing scan job")
		lh.Entries().Get()[lh.Len()-2].ExpErr(errors.New("processing error"))
		lh.Entries().Get()[lh.Len()-2].ExpLevel(zerolog.ErrorLevel)
		lh.LastEntry().ExpLevel(zerolog.InfoLevel)
		lh.LastEntry().ExpMsg("finished processing all scan jobs")
	})

	t.Run("scan error", func(t *testing.T) {
		scanner, lh, teardown := setupCourseScanner(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]

		// Add the job
		scan, err := scanner.Add(course.ID)
		require.Nil(t, err)
		assert.Equal(t, scan.CourseID, course.ID)

		// Drop the DB
		_, err = scanner.db.Exec("DROP TABLE IF EXISTS " + models.TableScans())
		require.Nil(t, err)

		// Start the worker
		go scanner.Worker(func(*CourseScanner, *models.Scan) error {
			return nil
		})

		// Give time for the worker to finish
		time.Sleep(time.Millisecond * 3)

		// Validate the logs
		require.NotNil(t, lh.LastEntry())
		lh.LastEntry().ExpMsg("error looking up next scan job")
		lh.LastEntry().ExpLevel(zerolog.ErrorLevel)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CourseProcessor(t *testing.T) {
	t.Run("scan nil", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		err := CourseProcessor(scanner, nil)
		require.EqualError(t, err, "scan cannot be empty")
	})

	t.Run("error getting course", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]

		// Drop the table
		_, err := scanner.db.Exec("DROP TABLE IF EXISTS " + models.TableCourses())
		require.Nil(t, err)

		err = CourseProcessor(scanner, scan)
		require.ErrorContains(t, err, fmt.Sprintf("no such table: %s", models.TableCourses()))
	})

	t.Run("path error", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]

		err := CourseProcessor(scanner, scan)
		require.EqualError(t, err, "unable to open path")
	})

	t.Run("card", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		// ----------------------------
		// Found card
		// ----------------------------
		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]
		require.Empty(t, course.CardPath)

		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/card.jpg", course.Path))

		err := CourseProcessor(scanner, scan)
		require.Nil(t, err)

		c, err := models.GetCourse(scanner.db, course.ID)
		require.Nil(t, err)
		assert.Equal(t, fmt.Sprintf("%s/card.jpg", course.Path), c.CardPath)

		// ----------------------------
		// Ignore card in chapter
		// ----------------------------
		course = models.NewTestCourses(t, scanner.db, 1)[0]
		scan = models.NewTestScans(t, scanner.db, []*models.Course{course})[0]
		require.Empty(t, course.CardPath)

		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/chapter 1/card.jpg", course.Path))

		err = CourseProcessor(scanner, scan)
		require.Nil(t, err)

		c, err = models.GetCourse(scanner.db, course.ID)
		require.Nil(t, err)
		assert.Empty(t, c.CardPath)
		assert.Empty(t, course.CardPath)

		// ----------------------------
		// Multiple cards types
		// ----------------------------
		course = models.NewTestCourses(t, scanner.db, 1)[0]
		scan = models.NewTestScans(t, scanner.db, []*models.Course{course})[0]
		require.Empty(t, course.CardPath)

		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/card.jpg", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/card.png", course.Path))

		err = CourseProcessor(scanner, scan)
		require.Nil(t, err)

		c, err = models.GetCourse(scanner.db, course.ID)
		require.Nil(t, err)
		assert.Equal(t, fmt.Sprintf("%s/card.jpg", course.Path), c.CardPath)
	})

	t.Run("card error", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]
		require.Empty(t, course.CardPath)

		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/card.jpg", course.Path))

		// Rename the card_path column
		_, err := scanner.db.Exec(fmt.Sprintf("ALTER TABLE %s RENAME COLUMN card_path TO ignore_card_path", models.TableCourses()))
		require.Nil(t, err)

		err = CourseProcessor(scanner, scan)
		require.ErrorContains(t, err, "no such column: card_path")
	})

	t.Run("ignore files", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]

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

		err := CourseProcessor(scanner, scan)
		require.Nil(t, err)

		assets, err := models.GetAssets(scanner.db, &database.DatabaseParams{Where: sq.Eq{models.TableAssets() + ".course_id": course.ID}})
		require.Nil(t, err)
		require.Len(t, assets, 0)
	})

	t.Run("assets", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]

		dbParams := &database.DatabaseParams{
			OrderBy: []string{"chapter asc", "prefix asc"},
			Where:   sq.Eq{models.TableAssets() + ".course_id": course.ID},
		}

		// ----------------------------
		// Add 2 assets
		// ----------------------------
		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 file 1.mkv", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/02 file 2.html", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/should ignore", course.Path))

		err := CourseProcessor(scanner, scan)
		require.Nil(t, err)

		assets, err := models.GetAssets(scanner.db, dbParams)
		require.Nil(t, err)
		require.Len(t, assets, 2)

		assert.Equal(t, "file 1", assets[0].Title)
		assert.Equal(t, course.ID, assets[0].CourseID)
		assert.Equal(t, 1, int(assets[0].Prefix.Int16))
		assert.Empty(t, assets[0].Chapter)
		assert.True(t, assets[0].Type.IsVideo())

		assert.Equal(t, "file 2", assets[1].Title)
		assert.Equal(t, course.ID, assets[1].CourseID)
		assert.Equal(t, 2, int(assets[1].Prefix.Int16))
		assert.Empty(t, assets[1].Chapter)
		assert.True(t, assets[1].Type.IsHTML())

		// ----------------------------
		// Delete asset
		// ----------------------------
		scanner.appFs.Fs.Remove(fmt.Sprintf("%s/01 file 1.mkv", course.Path))

		err = CourseProcessor(scanner, scan)
		require.Nil(t, err)

		assets, err = models.GetAssets(scanner.db, dbParams)
		require.Nil(t, err)
		require.Len(t, assets, 1)

		assert.Equal(t, "file 2", assets[0].Title)
		assert.Equal(t, course.ID, assets[0].CourseID)
		assert.Equal(t, 2, int(assets[0].Prefix.Int16))
		assert.Empty(t, assets[0].Chapter)
		assert.True(t, assets[0].Type.IsHTML())

		// ----------------------------
		// Add chapter asset
		// ----------------------------
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 Chapter 1/03 file 3.pdf", course.Path))

		err = CourseProcessor(scanner, scan)
		require.Nil(t, err)

		assets, err = models.GetAssets(scanner.db, dbParams)
		require.Nil(t, err)
		require.Len(t, assets, 2)

		assert.Equal(t, "file 2", assets[0].Title)
		assert.Equal(t, course.ID, assets[0].CourseID)
		assert.Equal(t, 2, int(assets[0].Prefix.Int16))
		assert.Empty(t, assets[0].Chapter)
		assert.True(t, assets[0].Type.IsHTML())

		assert.Equal(t, "file 3", assets[1].Title)
		assert.Equal(t, course.ID, assets[1].CourseID)
		assert.Equal(t, 3, int(assets[1].Prefix.Int16))
		assert.Equal(t, "01 Chapter 1", assets[1].Chapter)
		assert.True(t, assets[1].Type.IsPDF())
	})

	t.Run("assets error", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]

		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 video.mkv", course.Path))

		_, err := scanner.db.Exec("DROP TABLE IF EXISTS " + models.TableAssets())
		require.Nil(t, err)

		err = CourseProcessor(scanner, scan)
		require.ErrorContains(t, err, "no such table: "+models.TableAssets())
	})

	t.Run("attachments", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]

		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)

		assetDbParams := &database.DatabaseParams{Where: sq.Eq{models.TableAssets() + ".course_id": course.ID}}
		attachmentDbParams := &database.DatabaseParams{
			OrderBy: []string{"created_at asc"},
			Where:   sq.Eq{models.TableAttachments() + ".course_id": course.ID},
		}

		// ----------------------------
		// Add 1 asset with 1 attachment
		// ----------------------------
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 video.mp4", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 info.txt", course.Path))

		err := CourseProcessor(scanner, scan)
		require.Nil(t, err)

		assets, err := models.GetAssets(scanner.db, assetDbParams)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		assert.Equal(t, "video", assets[0].Title)
		assert.Equal(t, course.ID, assets[0].CourseID)
		assert.Equal(t, 1, int(assets[0].Prefix.Int16))
		assert.Equal(t, fmt.Sprintf("%s/01 video.mp4", course.Path), assets[0].Path)
		assert.True(t, assets[0].Type.IsVideo())

		attachments, err := models.GetAttachments(scanner.db, attachmentDbParams)
		require.Nil(t, err)
		require.Len(t, attachments, 1)
		assert.Equal(t, "info.txt", attachments[0].Title)
		assert.Equal(t, fmt.Sprintf("%s/01 info.txt", course.Path), attachments[0].Path)

		// ----------------------------
		// Add another attachment
		// ----------------------------
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 code.zip", course.Path))

		err = CourseProcessor(scanner, scan)
		require.Nil(t, err)

		// Assert the asset
		assets, err = models.GetAssets(scanner.db, assetDbParams)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		require.Len(t, assets[0].Attachments, 2)

		attachments, err = models.GetAttachments(scanner.db, attachmentDbParams)
		require.Nil(t, err)
		require.Len(t, attachments, 2)
		assert.Equal(t, "info.txt", attachments[0].Title)
		assert.Equal(t, fmt.Sprintf("%s/01 info.txt", course.Path), attachments[0].Path)
		assert.Equal(t, "code.zip", attachments[1].Title)
		assert.Equal(t, fmt.Sprintf("%s/01 code.zip", course.Path), attachments[1].Path)

		// ----------------------------
		// Delete first attachment
		// ----------------------------
		scanner.appFs.Fs.Remove(fmt.Sprintf("%s/01 info.txt", course.Path))

		err = CourseProcessor(scanner, scan)
		require.Nil(t, err)

		// Assert the asset
		assets, err = models.GetAssets(scanner.db, assetDbParams)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		assert.Equal(t, "video", assets[0].Title)
		require.Len(t, assets[0].Attachments, 1)

		attachments, err = models.GetAttachments(scanner.db, attachmentDbParams)
		require.Nil(t, err)
		require.Len(t, attachments, 1)
		assert.Equal(t, "code.zip", attachments[0].Title)
		assert.Equal(t, fmt.Sprintf("%s/01 code.zip", course.Path), attachments[0].Path)
	})

	t.Run("attachments error", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]

		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 video.mkv", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 info", course.Path))

		// Drop the attachments table
		_, err := scanner.db.Exec("DROP TABLE IF EXISTS " + models.TableAttachments())
		require.Nil(t, err)

		err = CourseProcessor(scanner, scan)
		require.ErrorContains(t, err, "no such table: "+models.TableAttachments())
	})

	t.Run("asset priority", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		// ----------------------------
		// Priority is VIDEO -> HTML -> PDF
		// ----------------------------

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]

		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)

		assetDbParams := &database.DatabaseParams{Where: sq.Eq{models.TableAssets() + ".course_id": course.ID}}
		attachmentDbParams := &database.DatabaseParams{
			OrderBy: []string{"created_at asc"},
			Where:   sq.Eq{models.TableAttachments() + ".course_id": course.ID},
		}

		// ----------------------------
		// Add PDF (asset)
		// ----------------------------
		pdfFile := fmt.Sprintf("%s/01 doc 1.pdf", course.Path)
		scanner.appFs.Fs.Create(pdfFile)

		err := CourseProcessor(scanner, scan)
		require.Nil(t, err)

		assets, err := models.GetAssets(scanner.db, assetDbParams)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		assert.Equal(t, pdfFile, assets[0].Path)
		assert.True(t, assets[0].Type.IsPDF())
		require.Empty(t, assets[0].Attachments)

		// ----------------------------
		// Add HTML (asset)
		// ----------------------------
		htmlFile := fmt.Sprintf("%s/01 example.html", course.Path)
		scanner.appFs.Fs.Create(htmlFile)

		err = CourseProcessor(scanner, scan)
		require.Nil(t, err)

		assets, err = models.GetAssets(scanner.db, assetDbParams)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		assert.Equal(t, htmlFile, assets[0].Path)
		assert.True(t, assets[0].Type.IsHTML())
		require.Len(t, assets[0].Attachments, 1)

		attachments, err := models.GetAttachments(scanner.db, attachmentDbParams)
		require.Nil(t, err)
		require.Len(t, attachments, 1)
		assert.Equal(t, pdfFile, attachments[0].Path)

		// ----------------------------
		// Add VIDEO (asset)
		// ----------------------------
		videoFile := fmt.Sprintf("%s/01 video.mp4", course.Path)
		scanner.appFs.Fs.Create(videoFile)

		err = CourseProcessor(scanner, scan)
		require.Nil(t, err)

		assets, err = models.GetAssets(scanner.db, assetDbParams)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		assert.Equal(t, videoFile, assets[0].Path)
		assert.True(t, assets[0].Type.IsVideo())
		require.Len(t, assets[0].Attachments, 2)

		attachments, err = models.GetAttachments(scanner.db, attachmentDbParams)
		require.Nil(t, err)
		require.Len(t, attachments, 2)
		assert.Equal(t, pdfFile, attachments[0].Path)
		assert.Equal(t, htmlFile, attachments[1].Path)

		// ----------------------------
		// Add PDF file (attachment)
		// ----------------------------
		pdfFile2 := fmt.Sprintf("%s/01 - e.pdf", course.Path)
		scanner.appFs.Fs.Create(pdfFile2)

		err = CourseProcessor(scanner, scan)
		require.Nil(t, err)

		assets, err = models.GetAssets(scanner.db, assetDbParams)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		assert.Equal(t, videoFile, assets[0].Path)
		assert.True(t, assets[0].Type.IsVideo())
		require.Len(t, assets[0].Attachments, 3)

		attachments, err = models.GetAttachments(scanner.db, attachmentDbParams)
		require.Nil(t, err)
		require.Len(t, attachments, 3)
		assert.Equal(t, pdfFile, attachments[0].Path)
		assert.Equal(t, htmlFile, attachments[1].Path)
		assert.Equal(t, pdfFile2, attachments[2].Path)
	})

	t.Run("course updated", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		origCourse := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{origCourse})[0]

		scanner.appFs.Fs.Mkdir(origCourse.Path, os.ModePerm)

		err := CourseProcessor(scanner, scan)
		require.Nil(t, err)

		updatedCourse, err := models.GetCourse(scanner.db, origCourse.ID)
		require.Nil(t, err)
		assert.NotEqual(t, origCourse.UpdatedAt, updatedCourse.UpdatedAt)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_BuildFileInfo(t *testing.T) {
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
			assert.Nil(t, fb)
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
			assert.Equal(t, tt.expected, fb, fmt.Sprintf("error for [%s]", tt.in))
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
			assert.Equal(t, tt.expected, fb, fmt.Sprintf("error for [%s]", tt.in))
		}
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_IsCard(t *testing.T) {
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
			assert.False(t, isCard(tt))
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
			assert.True(t, isCard(tt))
		}
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateAssets(t *testing.T) {
	t.Run("nothing added or deleted", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		courses := models.NewTestCourses(t, scanner.db, 1)
		assets := models.NewTestAssets(t, scanner.db, courses, 10)

		err := updateAssets(scanner.db, courses[0].ID, assets)
		require.Nil(t, err)

		dbParams := &database.DatabaseParams{Where: sq.Eq{models.TableAssets() + ".course_id": courses[0].ID}}
		count, err := models.CountAssets(scanner.db, dbParams)
		require.Nil(t, err)
		require.Equal(t, 10, count)
	})

	t.Run("add", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		courses := models.NewTestCourses(t, scanner.db, 1)
		assets := models.NewTestAssets(t, nil, courses, 10)

		dbParams := &database.DatabaseParams{Where: sq.Eq{models.TableAssets() + ".course_id": courses[0].ID}}

		// ----------------------------
		// Add 10 assets
		// ----------------------------
		err := updateAssets(scanner.db, courses[0].ID, assets)
		require.Nil(t, err)

		count, err := models.CountAssets(scanner.db, dbParams)
		require.Nil(t, err)
		require.Equal(t, 10, count)

		// ----------------------------
		// Add another 2 assets
		// ----------------------------
		additionalAssets := models.NewTestAssets(t, nil, courses, 2)
		additionalAssets[0].ID = ""
		additionalAssets[1].ID = ""
		assets = append(assets, additionalAssets...)

		err = updateAssets(scanner.db, courses[0].ID, assets)
		require.Nil(t, err)

		count, err = models.CountAssets(scanner.db, dbParams)
		require.Nil(t, err)
		require.Equal(t, 12, count)

		// Ensure all assets have an ID
		for _, a := range assets {
			assert.NotEmpty(t, a.ID)
		}
	})

	t.Run("delete", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		courses := models.NewTestCourses(t, scanner.db, 1)
		assets := models.NewTestAssets(t, scanner.db, courses, 12)

		dbParams := &database.DatabaseParams{Where: sq.Eq{models.TableAssets() + ".course_id": courses[0].ID}}

		// ----------------------------
		// Remove 2 assets
		// ----------------------------
		assets = assets[2:]

		err := updateAssets(scanner.db, courses[0].ID, assets)
		require.Nil(t, err)

		count, err := models.CountAssets(scanner.db, dbParams)
		require.Nil(t, err)
		require.Equal(t, 10, count)

		// ----------------------------
		// Remove another 2 assets
		// ----------------------------
		assets = assets[2:]

		err = updateAssets(scanner.db, courses[0].ID, assets)
		require.Nil(t, err)

		count, err = models.CountAssets(scanner.db, dbParams)
		require.Nil(t, err)
		require.Equal(t, 8, count)
	})

	t.Run("db error", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		// Drop the table
		_, err := scanner.db.Exec("DROP TABLE IF EXISTS " + models.TableAssets())
		require.Nil(t, err)

		err = updateAssets(scanner.db, "1234", []*models.Asset{})
		require.ErrorContains(t, err, fmt.Sprintf("no such table: %s", models.TableAssets()))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateAttachments(t *testing.T) {
	t.Run("nothing added or delete)", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		courses := models.NewTestCourses(t, scanner.db, 1)
		assets := models.NewTestAssets(t, scanner.db, courses, 1)
		attachments := models.NewTestAttachments(t, scanner.db, assets, 10)

		err := updateAttachments(scanner.db, courses[0].ID, attachments)
		require.Nil(t, err)

		count, err := models.CountAttachments(scanner.db, &database.DatabaseParams{Where: sq.Eq{models.TableAttachments() + ".course_id": courses[0].ID}})
		require.Nil(t, err)
		require.Equal(t, 10, count)
	})

	t.Run("add", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		courses := models.NewTestCourses(t, scanner.db, 1)
		assets := models.NewTestAssets(t, scanner.db, courses, 1)
		attachments := models.NewTestAttachments(t, nil, assets, 10)

		dbParams := &database.DatabaseParams{Where: sq.Eq{models.TableAttachments() + ".course_id": courses[0].ID}}

		// ----------------------------
		// Add 10 attachments
		// ----------------------------
		err := updateAttachments(scanner.db, courses[0].ID, attachments)
		require.Nil(t, err)

		count, err := models.CountAttachments(scanner.db, dbParams)
		require.Nil(t, err)
		require.Equal(t, 10, count)

		// ----------------------------
		// Add another 2 attachments
		// ----------------------------
		additionalAttachments := models.NewTestAttachments(t, nil, assets, 2)
		attachments = append(attachments, additionalAttachments...)

		err = updateAttachments(scanner.db, courses[0].ID, attachments)
		require.Nil(t, err)

		count, err = models.CountAttachments(scanner.db, dbParams)
		require.Nil(t, err)
		require.Equal(t, 12, count)
	})

	t.Run("delete", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		courses := models.NewTestCourses(t, scanner.db, 1)
		assets := models.NewTestAssets(t, scanner.db, courses, 1)
		attachments := models.NewTestAttachments(t, scanner.db, assets, 12)

		dbParams := &database.DatabaseParams{Where: sq.Eq{models.TableAttachments() + ".course_id": courses[0].ID}}

		// ----------------------------
		// Remove 2 attachments
		// ----------------------------
		attachments = attachments[2:]

		err := updateAttachments(scanner.db, courses[0].ID, attachments)
		require.Nil(t, err)

		count, err := models.CountAttachments(scanner.db, dbParams)
		require.Nil(t, err)
		require.Equal(t, 10, count)

		// ----------------------------
		// Remove another 2 attachments
		// ----------------------------
		attachments = attachments[2:]

		err = updateAttachments(scanner.db, courses[0].ID, attachments)
		require.Nil(t, err)

		count, err = models.CountAttachments(scanner.db, dbParams)
		require.Nil(t, err)
		require.Equal(t, 8, count)

	})

	t.Run("db error", func(t *testing.T) {
		scanner, _, teardown := setupCourseScanner(t)
		defer teardown(t)

		// Drop the table
		_, err := scanner.db.Exec("DROP TABLE IF EXISTS " + models.TableAttachments())
		require.Nil(t, err)

		err = updateAttachments(scanner.db, "1234", []*models.Attachment{})
		require.ErrorContains(t, err, fmt.Sprintf("no such table: %s", models.TableAttachments()))
	})
}
