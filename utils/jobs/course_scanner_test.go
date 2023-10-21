package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/migrations"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rzajac/zltest"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Add(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]

		scan, err := scanner.Add(course.ID)
		require.Nil(t, err)
		assert.Equal(t, scan.CourseID, course.ID)
	})

	t.Run("duplicate", func(t *testing.T) {
		scanner, lh, teardown := setup(t)
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
		lh.LastEntry().ExpMsg("scan job already exists")
		lh.LastEntry().ExpLevel(zerolog.DebugLevel)
	})

	t.Run("invalid course", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		scan, err := scanner.Add("bob")
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, scan)

		count, err := models.CountScans(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		require.Equal(t, 0, count)
	})

	t.Run("not blocked", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		courses := models.NewTestCourses(t, scanner.db, 2)

		scan1, err := scanner.Add(courses[0].ID)
		require.Nil(t, err)
		assert.Equal(t, scan1.CourseID, courses[0].ID)

		scan2, err := scanner.Add(courses[1].ID)
		require.Nil(t, err)
		assert.Equal(t, scan2.CourseID, courses[1].ID)
	})

	t.Run("scan lookup error", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		// Drop table
		_, err := scanner.db.DB().NewDropTable().Model(&models.Scan{}).Exec(scanner.ctx)
		require.Nil(t, err)

		scan, err := scanner.Add("")
		require.ErrorContains(t, err, "no such table: scans")
		assert.Nil(t, scan)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Worker(t *testing.T) {
	t.Run("single job", func(t *testing.T) {
		scanner, lh, teardown := setup(t)
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
		count, err := models.CountScans(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		assert.Equal(t, 0, count)

		// Validate the logs
		require.NotNil(t, lh.LastEntry())
		lh.Entries().Get()[lh.Len()-2].ExpMsg("finished processing scan job")
		lh.Entries().Get()[lh.Len()-2].ExpStr("path", course.Path)
		lh.LastEntry().ExpLevel(zerolog.InfoLevel)
		lh.LastEntry().ExpMsg("finished processing all scan jobs")
	})

	t.Run("several jobs", func(t *testing.T) {
		scanner, lh, teardown := setup(t)
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
		time.Sleep(time.Millisecond * 10)

		// Assert the scan job was deleted from the DB
		count, err := models.CountScans(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		assert.Equal(t, 0, count)

		// Validate the logs.
		require.NotNil(t, lh.LastEntry())
		lh.Entries().Get()[lh.Len()-2].ExpMsg("finished processing scan job")
		lh.Entries().Get()[lh.Len()-2].ExpStr("path", courses[2].Path)
		lh.LastEntry().ExpLevel(zerolog.InfoLevel)
		lh.LastEntry().ExpMsg("finished processing all scan jobs")
	})

	t.Run("error processing", func(t *testing.T) {
		scanner, lh, teardown := setup(t)
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
		scanner, lh, teardown := setup(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]

		// Add the job
		scan, err := scanner.Add(course.ID)
		require.Nil(t, err)
		assert.Equal(t, scan.CourseID, course.ID)

		// Drop the DB
		_, err = scanner.db.DB().NewDropTable().Model(&models.Scan{}).Exec(scanner.ctx)
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
		scanner, _, teardown := setup(t)
		defer teardown(t)

		err := CourseProcessor(scanner, nil)
		require.EqualError(t, err, "scan cannot be empty")
	})

	t.Run("error updating scan status", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]

		_, err := scanner.db.DB().NewDropTable().Model(&models.Scan{}).Exec(scanner.ctx)
		require.Nil(t, err)

		err = CourseProcessor(scanner, scan)
		require.ErrorContains(t, err, "no such table: scans")
	})

	t.Run("error getting course", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]

		_, err := scanner.db.DB().NewDropTable().Model(&models.Course{}).Exec(scanner.ctx)
		require.Nil(t, err)

		err = CourseProcessor(scanner, scan)
		require.ErrorContains(t, err, "no such table: courses")
	})

	t.Run("path error", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]

		err := CourseProcessor(scanner, scan)
		require.EqualError(t, err, "unable to open path")
	})

	t.Run("found card", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]
		require.Empty(t, course.CardPath)

		// Create card
		cardPath := fmt.Sprintf("%s/card.jpg", course.Path)
		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		scanner.appFs.Fs.Create(cardPath)

		err := CourseProcessor(scanner, scan)
		require.Nil(t, err)

		// Asset the course card
		foundCourse, err := models.GetCourseById(scanner.ctx, scanner.db, nil, course.ID)
		require.Nil(t, err)
		assert.Equal(t, cardPath, foundCourse.CardPath)
	})

	t.Run("ignore card in chapter", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]
		require.Empty(t, course.CardPath)

		// Create a card within a chapter
		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/chapter 1/card.jpg", course.Path))

		err := CourseProcessor(scanner, scan)
		require.Nil(t, err)

		// Assert the course card
		foundCourse, err := models.GetCourseById(scanner.ctx, scanner.db, nil, course.ID)
		require.Nil(t, err)
		assert.Empty(t, foundCourse.CardPath)
		assert.Empty(t, course.CardPath)
	})

	t.Run("multiple cards", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]
		require.Empty(t, course.CardPath)

		// Create multiple cards. The first one found should be used
		cardPath1 := fmt.Sprintf("%s/card.jpg", course.Path)
		cardPath2 := fmt.Sprintf("%s/card.png", course.Path)
		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		scanner.appFs.Fs.Create(cardPath1)
		scanner.appFs.Fs.Create(cardPath2)

		err := CourseProcessor(scanner, scan)
		require.Nil(t, err)

		// Assert the course card
		foundCourse, err := models.GetCourseById(scanner.ctx, scanner.db, nil, course.ID)
		require.Nil(t, err)
		assert.Equal(t, cardPath1, foundCourse.CardPath)
	})

	t.Run("card error", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]
		require.Empty(t, course.CardPath)

		cardPath := fmt.Sprintf("%s/card.jpg", course.Path)
		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		scanner.appFs.Fs.Create(cardPath)

		// Drop the card path column from the DB
		_, err := scanner.db.DB().NewDropColumn().Model(&models.Course{}).Column("card_path").Exec(scanner.ctx)
		require.Nil(t, err)

		err = CourseProcessor(scanner, scan)
		require.ErrorContains(t, err, "no such column: card_path")
	})

	t.Run("ignore files", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]

		// Create a video file that will be ignored due to no prefix
		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/video 1", course.Path))

		err := CourseProcessor(scanner, scan)
		require.Nil(t, err)

		// Assert the assets
		assets, err := models.GetAssetsByCourseId(scanner.ctx, scanner.db, nil, course.ID)
		require.Nil(t, err)
		require.Len(t, assets, 0)
	})

	t.Run("assets", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]

		// Create the 2 assets (including an ignored asset)
		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 - video.mkv", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/02 - index.html", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 Getting Started/ignored.mp4", course.Path))

		err := CourseProcessor(scanner, scan)
		require.Nil(t, err)

		// Assert there are 2 assets
		dbParams := &database.DatabaseParams{OrderBy: []string{"chapter asc", "prefix asc"}}
		assets, err := models.GetAssetsByCourseId(scanner.ctx, scanner.db, dbParams, course.ID)
		require.Nil(t, err)
		require.Len(t, assets, 2)

		assert.Equal(t, "video", assets[0].Title)
		assert.Equal(t, course.ID, assets[0].CourseID)
		assert.Equal(t, 1, assets[0].Prefix)
		assert.Empty(t, assets[0].Chapter)
		assert.True(t, assets[0].Type.IsVideo())

		assert.Equal(t, "index", assets[1].Title)
		assert.Equal(t, course.ID, assets[1].CourseID)
		assert.Equal(t, 2, assets[1].Prefix)
		assert.Empty(t, assets[1].Chapter)
		assert.True(t, assets[1].Type.IsHTML())

		// ----------------------------
		// Delete asset
		// ----------------------------
		scanner.appFs.Fs.Remove(fmt.Sprintf("%s/01 - video.mkv", course.Path))

		err = CourseProcessor(scanner, scan)
		require.Nil(t, err)

		// Assert there is 1 asset
		assets, err = models.GetAssetsByCourseId(scanner.ctx, scanner.db, dbParams, course.ID)
		require.Nil(t, err)
		require.Len(t, assets, 1)

		assert.Equal(t, "index", assets[0].Title)
		assert.Equal(t, course.ID, assets[0].CourseID)
		assert.Equal(t, 2, assets[0].Prefix)
		assert.Empty(t, assets[0].Chapter)
		assert.True(t, assets[0].Type.IsHTML())

		// ----------------------------
		// Add another asset
		// ----------------------------
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 Getting Started/03 - info.pdf", course.Path))

		err = CourseProcessor(scanner, scan)
		require.Nil(t, err)

		// Assert there are 2 assets
		assets, err = models.GetAssetsByCourseId(scanner.ctx, scanner.db, dbParams, course.ID)
		require.Nil(t, err)
		require.Len(t, assets, 2)

		assert.Equal(t, "index", assets[0].Title)
		assert.Equal(t, course.ID, assets[0].CourseID)
		assert.Equal(t, 2, assets[0].Prefix)
		assert.Empty(t, assets[0].Chapter)
		assert.True(t, assets[0].Type.IsHTML())

		assert.Equal(t, "info", assets[1].Title)
		assert.Equal(t, course.ID, assets[1].CourseID)
		assert.Equal(t, 3, assets[1].Prefix)
		assert.Equal(t, "01 Getting Started", assets[1].Chapter)

		assert.True(t, assets[1].Type.IsPDF())
	})

	t.Run("assets error", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]

		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 - video.mkv", course.Path))

		// Drop the assets table
		_, err := scanner.db.DB().NewDropTable().Model(&models.Asset{}).Exec(scanner.ctx)
		require.Nil(t, err)

		err = CourseProcessor(scanner, scan)
		require.ErrorContains(t, err, "no such table: assets")
	})

	t.Run("attachments", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]

		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)

		assetDbParams := &database.DatabaseParams{Relation: []database.Relation{{Struct: "Attachments"}}}
		attachmentDbParams := &database.DatabaseParams{OrderBy: []string{"title asc"}}

		// Add video asset with 1 attachment
		videoFile := fmt.Sprintf("%s/01 - video.mp4", course.Path)
		attachmentFile := fmt.Sprintf("%s/01 - a.txt", course.Path)
		scanner.appFs.Fs.Create(videoFile)
		scanner.appFs.Fs.Create(attachmentFile)

		err := CourseProcessor(scanner, scan)
		require.Nil(t, err)

		// Assert the asset
		assets, err := models.GetAssetsByCourseId(scanner.ctx, scanner.db, assetDbParams, course.ID)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		assert.Equal(t, "video", assets[0].Title)
		assert.Equal(t, course.ID, assets[0].CourseID)
		assert.Equal(t, 1, assets[0].Prefix)
		assert.Equal(t, videoFile, assets[0].Path)
		assert.True(t, assets[0].Type.IsVideo())
		require.Len(t, assets[0].Attachments, 1)

		// Assert there is 1 attachment
		attachments, err := models.GetAttachmentsByCourseId(scanner.ctx, scanner.db, attachmentDbParams, course.ID)
		require.Nil(t, err)
		require.Len(t, attachments, 1)
		assert.Equal(t, attachmentFile, attachments[0].Path)

		// ----------------------------
		// Add another attachment
		// ----------------------------
		attachmentFile2 := fmt.Sprintf("%s/01 - b.zip", course.Path)
		scanner.appFs.Fs.Create(attachmentFile2)

		err = CourseProcessor(scanner, scan)
		require.Nil(t, err)

		// Assert the asset
		assets, err = models.GetAssetsByCourseId(scanner.ctx, scanner.db, assetDbParams, course.ID)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		assert.Equal(t, "video", assets[0].Title)
		require.Len(t, assets[0].Attachments, 2)

		// Assert there are 2 attachments
		attachments, err = models.GetAttachmentsByCourseId(scanner.ctx, scanner.db, attachmentDbParams, course.ID)
		require.Nil(t, err)
		require.Len(t, attachments, 2)
		assert.Equal(t, attachmentFile, attachments[0].Path)
		assert.Equal(t, attachmentFile2, attachments[1].Path)

		// ----------------------------
		// Delete first attachment
		// ----------------------------
		scanner.appFs.Fs.Remove(attachmentFile)

		err = CourseProcessor(scanner, scan)
		require.Nil(t, err)

		// Assert the asset
		assets, err = models.GetAssetsByCourseId(scanner.ctx, scanner.db, assetDbParams, course.ID)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		assert.Equal(t, "video", assets[0].Title)
		require.Len(t, assets[0].Attachments, 1)

		// Assert there is 1 attachment
		attachments, err = models.GetAttachmentsByCourseId(scanner.ctx, scanner.db, attachmentDbParams, course.ID)
		require.Nil(t, err)
		require.Len(t, attachments, 1)
		assert.Equal(t, attachmentFile2, attachments[0].Path)
	})

	t.Run("attachments error", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]

		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 - video.mkv", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 - info.txt", course.Path))

		// Drop the attachments table
		_, err := scanner.db.DB().NewDropTable().Model(&models.Attachment{}).Exec(scanner.ctx)
		require.Nil(t, err)

		err = CourseProcessor(scanner, scan)
		require.ErrorContains(t, err, "no such table: attachments")
	})

	t.Run("asset priority", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		scan := models.NewTestScans(t, scanner.db, []*models.Course{course})[0]

		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)

		attachmentDbParams := &database.DatabaseParams{OrderBy: []string{"title asc"}}

		// ----------------------------
		// PDF
		// ----------------------------
		pdfFile := fmt.Sprintf("%s/01 - a.pdf", course.Path)
		scanner.appFs.Fs.Create(pdfFile)

		err := CourseProcessor(scanner, scan)
		require.Nil(t, err)

		// Assert the pdf
		assets, err := models.GetAssetsByCourseId(scanner.ctx, scanner.db, nil, course.ID)
		require.Nil(t, err)
		require.Len(t, assets, 1)

		// Assert the HTML file is the asset
		assert.Equal(t, "a", assets[0].Title)
		assert.Equal(t, course.ID, assets[0].CourseID)
		assert.Equal(t, 1, assets[0].Prefix)
		assert.Equal(t, pdfFile, assets[0].Path)
		assert.True(t, assets[0].Type.IsPDF())

		// Assert the attachments
		attachments, err := models.GetAttachmentsByCourseId(scanner.ctx, scanner.db, attachmentDbParams, course.ID)
		require.Nil(t, err)
		assert.Empty(t, attachments)

		// ----------------------------
		// HTML trumps PDF
		// ----------------------------
		htmlFile := fmt.Sprintf("%s/01 - b.html", course.Path)
		scanner.appFs.Fs.Create(htmlFile)

		err = CourseProcessor(scanner, scan)
		require.Nil(t, err)

		// Assert the pdf
		assets, err = models.GetAssetsByCourseId(scanner.ctx, scanner.db, nil, course.ID)
		require.Nil(t, err)
		require.Len(t, assets, 1)

		// Assert the HTML file is the asset
		assert.Equal(t, "b", assets[0].Title)
		assert.Equal(t, course.ID, assets[0].CourseID)
		assert.Equal(t, 1, assets[0].Prefix)
		assert.Equal(t, htmlFile, assets[0].Path)
		assert.True(t, assets[0].Type.IsHTML())

		// Assert the attachments (pdfFile should be the last attachment)
		attachments, err = models.GetAttachmentsByCourseId(scanner.ctx, scanner.db, attachmentDbParams, course.ID)
		require.Nil(t, err)
		require.Len(t, attachments, 1)
		assert.Equal(t, pdfFile, attachments[0].Path)

		// ----------------------------
		// Video trumps PDF
		// ----------------------------
		videoFile := fmt.Sprintf("%s/01 - c.mp4", course.Path)
		scanner.appFs.Fs.Create(videoFile)

		err = CourseProcessor(scanner, scan)
		require.Nil(t, err)

		// Assert the pdf
		assets, err = models.GetAssetsByCourseId(scanner.ctx, scanner.db, nil, course.ID)
		require.Nil(t, err)
		require.Len(t, assets, 1)

		// Assert the video file is the asset
		assert.Equal(t, "c", assets[0].Title)
		assert.Equal(t, course.ID, assets[0].CourseID)
		assert.Equal(t, 1, assets[0].Prefix)
		assert.Equal(t, videoFile, assets[0].Path)
		assert.True(t, assets[0].Type.IsVideo())

		// Assert the attachments (html file should be the last attachment)
		attachments, err = models.GetAttachmentsByCourseId(scanner.ctx, scanner.db, attachmentDbParams, course.ID)
		require.Nil(t, err)
		require.Len(t, attachments, 2)
		assert.Equal(t, htmlFile, attachments[1].Path)

		// ----------------------------
		// Another HTML file (should become attachment)
		// ----------------------------
		htmlFile2 := fmt.Sprintf("%s/01 - d.html", course.Path)
		scanner.appFs.Fs.Create(htmlFile2)

		err = CourseProcessor(scanner, scan)
		require.Nil(t, err)

		// Assert the video file is the asset
		assets, err = models.GetAssetsByCourseId(scanner.ctx, scanner.db, nil, course.ID)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		assert.Equal(t, "c", assets[0].Title)
		assert.Equal(t, videoFile, assets[0].Path)
		assert.True(t, assets[0].Type.IsVideo())

		// Assert the attachments (htmlfile2 should be the last attachment)
		attachments, err = models.GetAttachmentsByCourseId(scanner.ctx, scanner.db, attachmentDbParams, course.ID)
		require.Nil(t, err)
		require.Len(t, attachments, 3)
		assert.Equal(t, htmlFile2, attachments[2].Path)

		// ----------------------------
		// Another PDF file (should become attachment)
		// ----------------------------
		pdfFile2 := fmt.Sprintf("%s/01 - e.pdf", course.Path)
		scanner.appFs.Fs.Create(pdfFile2)

		err = CourseProcessor(scanner, scan)
		require.Nil(t, err)

		// Assert the video file is the asset
		assets, err = models.GetAssetsByCourseId(scanner.ctx, scanner.db, nil, course.ID)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		assert.Equal(t, "c", assets[0].Title)
		assert.Equal(t, videoFile, assets[0].Path)
		assert.True(t, assets[0].Type.IsVideo())

		// Assert the attachments (pdfFile2 should be the last attachment)
		attachments, err = models.GetAttachmentsByCourseId(scanner.ctx, scanner.db, attachmentDbParams, course.ID)
		require.Nil(t, err)
		require.Len(t, attachments, 4)
		assert.Equal(t, pdfFile2, attachments[3].Path)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_BuildFileInfo(t *testing.T) {
	t.Run("invalid assets", func(t *testing.T) {
		var tests = []string{
			// No prefix
			"file",
			"file.file",
			"file.avi",
			" - file.avi",
			"- - file.avi",
			"-1 - file.avi",
			// Invalid prefix
			"a - file.avi",
			// No title
			"1 - .avi",
			"1 .avi",
			"1      .avi",
			"1 --  -- .avi",
		}

		for _, tt := range tests {
			fb := buildFileInfo(tt)
			assert.Nil(t, fb)
		}
	})

	t.Run("valid assets", func(t *testing.T) {
		var tests = []struct {
			in       string
			expected *fileInfo
		}{
			// Various valid formats of prefix/title/extension
			{"0    file 0.avi", &fileInfo{prefix: 0, title: "file 0", ext: "avi", titleWithExt: "file 0.avi", assetType: *types.NewAsset("avi"), isAsset: true}},
			{"001 file 1.avi", &fileInfo{prefix: 1, title: "file 1", ext: "avi", titleWithExt: "file 1.avi", assetType: *types.NewAsset("avi"), isAsset: true}},
			{"1 - file.avi", &fileInfo{prefix: 1, title: "file", ext: "avi", titleWithExt: "file.avi", assetType: *types.NewAsset("avi"), isAsset: true}},
			{"1-file.avi", &fileInfo{prefix: 1, title: "file", ext: "avi", titleWithExt: "file.avi", assetType: *types.NewAsset("avi"), isAsset: true}},
			{"1 --- file.avi", &fileInfo{prefix: 1, title: "file", ext: "avi", titleWithExt: "file.avi", assetType: *types.NewAsset("avi"), isAsset: true}},
			{"1 file.avi", &fileInfo{prefix: 1, title: "file", ext: "avi", titleWithExt: "file.avi", assetType: *types.NewAsset("avi"), isAsset: true}},
			// Videos
			{"1 - video.avi", &fileInfo{prefix: 1, title: "video", ext: "avi", titleWithExt: "video.avi", assetType: *types.NewAsset("avi"), isAsset: true}},
			{"1 - video.mkv", &fileInfo{prefix: 1, title: "video", ext: "mkv", titleWithExt: "video.mkv", assetType: *types.NewAsset("mkv"), isAsset: true}},
			{"1 - video.flac", &fileInfo{prefix: 1, title: "video", ext: "flac", titleWithExt: "video.flac", assetType: *types.NewAsset("flac"), isAsset: true}},
			{"1 - video.mp4", &fileInfo{prefix: 1, title: "video", ext: "mp4", titleWithExt: "video.mp4", assetType: *types.NewAsset("mp4"), isAsset: true}},
			{"1 - video.m4a", &fileInfo{prefix: 1, title: "video", ext: "m4a", titleWithExt: "video.m4a", assetType: *types.NewAsset("m4a"), isAsset: true}},
			{"1 - video.mp3", &fileInfo{prefix: 1, title: "video", ext: "mp3", titleWithExt: "video.mp3", assetType: *types.NewAsset("mp3"), isAsset: true}},
			{"1 - video.ogv", &fileInfo{prefix: 1, title: "video", ext: "ogv", titleWithExt: "video.ogv", assetType: *types.NewAsset("ogv"), isAsset: true}},
			{"1 - video.ogm", &fileInfo{prefix: 1, title: "video", ext: "ogm", titleWithExt: "video.ogm", assetType: *types.NewAsset("ogm"), isAsset: true}},
			{"1 - video.ogg", &fileInfo{prefix: 1, title: "video", ext: "ogg", titleWithExt: "video.ogg", assetType: *types.NewAsset("ogg"), isAsset: true}},
			{"1 - video.oga", &fileInfo{prefix: 1, title: "video", ext: "oga", titleWithExt: "video.oga", assetType: *types.NewAsset("oga"), isAsset: true}},
			{"1 - video.opus", &fileInfo{prefix: 1, title: "video", ext: "opus", titleWithExt: "video.opus", assetType: *types.NewAsset("opus"), isAsset: true}},
			{"1 - video.webm", &fileInfo{prefix: 1, title: "video", ext: "webm", titleWithExt: "video.webm", assetType: *types.NewAsset("webm"), isAsset: true}},
			{"1 - video.wav", &fileInfo{prefix: 1, title: "video", ext: "wav", titleWithExt: "video.wav", assetType: *types.NewAsset("wav"), isAsset: true}},
			// PDF
			// Document
			{"1 - doc.pdf", &fileInfo{prefix: 1, title: "doc", ext: "pdf", titleWithExt: "doc.pdf", assetType: *types.NewAsset("pdf"), isAsset: true}},
			// HTML
			{"1 - index.html", &fileInfo{prefix: 1, title: "index", ext: "html", titleWithExt: "index.html", assetType: *types.NewAsset("html"), isAsset: true}},
			{"1 - index.htm", &fileInfo{prefix: 1, title: "index", ext: "htm", titleWithExt: "index.htm", assetType: *types.NewAsset("htm"), isAsset: true}},
		}

		for _, tt := range tests {
			fb := buildFileInfo(tt.in)
			assert.Equal(t, tt.expected, fb, fmt.Sprintf("error for [%s]", tt.in))
		}
	})

	t.Run("valid attachments", func(t *testing.T) {
		var tests = []struct {
			in       string
			expected *fileInfo
		}{
			// Various valid formats of prefix/title/extension
			{"0    file 0", &fileInfo{prefix: 0, title: "file 0", titleWithExt: "file 0", isAsset: false}},
			{"001    file 1", &fileInfo{prefix: 1, title: "file 1", titleWithExt: "file 1", isAsset: false}},
			{"1 - file", &fileInfo{prefix: 1, title: "file", titleWithExt: "file", isAsset: false}},
			{"1-file", &fileInfo{prefix: 1, title: "file", titleWithExt: "file", isAsset: false}},
			{"1 --- file", &fileInfo{prefix: 1, title: "file", titleWithExt: "file", isAsset: false}},
			{"1 file.txt", &fileInfo{prefix: 1, title: "file", ext: "txt", titleWithExt: "file.txt", isAsset: false}},
		}

		for _, tt := range tests {
			fb := buildFileInfo(tt.in)
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
	t.Run("db error", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		// Drop the table
		_, err := scanner.db.DB().NewDropTable().Model(&models.Asset{}).Exec(scanner.ctx)
		require.Nil(t, err)

		err = updateAssets(scanner.ctx, scanner.db, "1234", []*models.Asset{})
		require.ErrorContains(t, err, "no such table: assets")
	})

	t.Run("nothing added or deleted", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		assets := models.NewTestAssets(t, scanner.db, []*models.Course{course}, 10)

		// Assert there are 10 assets
		count, err := models.CountAssets(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		assert.Equal(t, 10, count)

		err = updateAssets(scanner.ctx, scanner.db, course.ID, assets)
		require.Nil(t, err)

		// Assert there are still 10 assets
		count, err = models.CountAssets(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		assert.Equal(t, 10, count)
	})

	t.Run("add", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		// Create 1 course (inserted into DB) with 10 assets (not inserted into DB)
		course := models.NewTestCourses(t, scanner.db, 1)[0]
		assets := models.NewTestAssets(t, nil, []*models.Course{course}, 10)

		// Assert there are no assets
		count, err := models.CountAssets(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		assert.Equal(t, 0, count)

		err = updateAssets(scanner.ctx, scanner.db, course.ID, assets)
		require.Nil(t, err)

		count, err = models.CountAssets(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		assert.Equal(t, 10, count)

		// Add another 2 assets
		assets = append(assets, models.NewTestAssets(t, nil, []*models.Course{course}, 2)...)

		err = updateAssets(scanner.ctx, scanner.db, course.ID, assets)
		require.Nil(t, err)

		count, err = models.CountAssets(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		assert.Equal(t, 12, count)
	})

	t.Run("delete", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		assets := models.NewTestAssets(t, scanner.db, []*models.Course{course}, 12)

		// Assert there are 12 assets
		count, err := models.CountAssets(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		assert.Equal(t, 12, count)

		// Remove the first 2 assets
		assets = assets[2:]

		err = updateAssets(scanner.ctx, scanner.db, course.ID, assets)
		require.Nil(t, err)

		// Assert there are 10 assets
		count, err = models.CountAssets(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		assert.Equal(t, 10, count)

		// Remove another 2
		assets = assets[2:]

		err = updateAssets(scanner.ctx, scanner.db, course.ID, assets)
		require.Nil(t, err)

		// Assert there are 8 assets
		count, err = models.CountAssets(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		assert.Equal(t, 8, count)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateAttachments(t *testing.T) {
	t.Run("db error)", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		// Drop the table
		_, err := scanner.db.DB().NewDropTable().Model(&models.Attachment{}).Exec(scanner.ctx)
		require.Nil(t, err)

		err = updateAttachments(scanner.ctx, scanner.db, "1234", []*models.Attachment{})
		require.ErrorContains(t, err, "no such table: attachments")
	})

	t.Run("nothing added or delete)", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		course := models.NewTestCourses(t, scanner.db, 1)[0]
		asset := models.NewTestAssets(t, scanner.db, []*models.Course{course}, 1)[0]
		attachments := models.NewTestAttachments(t, scanner.db, []*models.Asset{asset}, 10)

		// Assert there are 10 attachments
		count, err := models.CountAttachments(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		assert.Equal(t, 10, count)

		err = updateAttachments(scanner.ctx, scanner.db, course.ID, attachments)
		require.Nil(t, err)

		// Assert there are still 10 attachments
		count, err = models.CountAttachments(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		assert.Equal(t, 10, count)
	})

	t.Run("add", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		// Create 1 course (inserted into DB) with 1 assets (inserted into DB) with 10 attachments
		// (not inserted into DB)
		course := models.NewTestCourses(t, scanner.db, 1)[0]
		asset := models.NewTestAssets(t, scanner.db, []*models.Course{course}, 1)[0]
		attachments := models.NewTestAttachments(t, nil, []*models.Asset{asset}, 10)

		// Assert there are no assets
		count, err := models.CountAttachments(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		assert.Equal(t, 0, count)

		err = updateAttachments(scanner.ctx, scanner.db, course.ID, attachments)
		require.Nil(t, err)

		count, err = models.CountAttachments(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		assert.Equal(t, 10, count)

		// Add another 2 attachments
		attachments = append(attachments, models.NewTestAttachments(t, nil, []*models.Asset{asset}, 2)...)

		err = updateAttachments(scanner.ctx, scanner.db, course.ID, attachments)
		require.Nil(t, err)

		count, err = models.CountAttachments(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		assert.Equal(t, 12, count)
	})

	t.Run("delete", func(t *testing.T) {
		scanner, _, teardown := setup(t)
		defer teardown(t)

		// Create 1 course (inserted into DB) with 1 assets (inserted into DB) with 10 attachments
		// (not inserted into DB)
		course := models.NewTestCourses(t, scanner.db, 1)[0]
		asset := models.NewTestAssets(t, scanner.db, []*models.Course{course}, 1)[0]
		attachments := models.NewTestAttachments(t, scanner.db, []*models.Asset{asset}, 12)

		// Assert there are 12 attachments
		count, err := models.CountAttachments(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		assert.Equal(t, 12, count)

		// Remove the first 2 attachments
		attachments = attachments[2:]

		err = updateAttachments(scanner.ctx, scanner.db, course.ID, attachments)
		require.Nil(t, err)

		// Assert there are 10 assets
		count, err = models.CountAttachments(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		assert.Equal(t, 10, count)

		// Remove another 2
		attachments = attachments[2:]

		err = updateAttachments(scanner.ctx, scanner.db, course.ID, attachments)
		require.Nil(t, err)

		// Assert there are 8 assets
		count, err = models.CountAttachments(scanner.ctx, scanner.db, nil)
		require.Nil(t, err)
		assert.Equal(t, 8, count)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPERS
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setup(t *testing.T) (*CourseScanner, *zltest.Tester, func(t *testing.T)) {
	// Set test logger
	loggerHook := zltest.New(t)
	log.Logger = zerolog.New(loggerHook).Level(zerolog.DebugLevel)

	appFs := appFs.NewAppFs(afero.NewMemMapFs())

	db := database.NewSqliteDB(&database.SqliteDbConfig{
		IsDebug: false,
		DataDir: "./co_data",
		AppFs:   appFs,
	})

	// Force DB to be in-memory
	os.Setenv("OC_InMemDb", "true")

	require.Nil(t, db.Bootstrap())
	require.Nil(t, migrations.Up(db))

	courseScanner := NewCourseScanner(&CourseScannerConfig{
		Db:    db,
		AppFs: appFs,
		Ctx:   context.Background(),
	})

	// teardown
	return courseScanner, loggerHook, func(t *testing.T) {
		os.Unsetenv("OC_InMemDb")
	}
}
