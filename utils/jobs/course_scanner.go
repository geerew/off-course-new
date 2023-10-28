package jobs

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/types"
	"github.com/rs/zerolog/log"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseScanner scans a course and finds assets and attachments
type CourseScanner struct {
	db        database.Database
	appFs     *appFs.AppFs
	ctx       context.Context
	jobSignal chan bool
	finished  bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseScannerConfig is the config for a CourseScanner
type CourseScannerConfig struct {
	Db    database.Database
	AppFs *appFs.AppFs
	Ctx   context.Context
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewCourseScanner creates a new CourseScanner
func NewCourseScanner(config *CourseScannerConfig) *CourseScanner {
	return &CourseScanner{
		db:        config.Db,
		appFs:     config.AppFs,
		ctx:       config.Ctx,
		jobSignal: make(chan bool, 1),
		finished:  false,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Add inserts a course scan job into the db
func (cs *CourseScanner) Add(courseId string) (*models.Scan, error) {
	// Ensure a job does not already exists for this course
	dbParams := &database.DatabaseParams{Relation: []database.Relation{{Struct: "Course"}}}

	scan, err := models.GetScanByCourseId(cs.ctx, cs.db, dbParams, courseId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	} else if scan != nil {
		log.Debug().Str("path", scan.Course.Path).Msg("scan job already exists")
		return nil, nil
	}

	// Get the course
	course, err := models.GetCourseById(cs.ctx, cs.db, nil, courseId)
	if err != nil {
		return nil, err
	}

	// Add the job
	scan = &models.Scan{CourseID: courseId}
	if err := models.CreateScan(cs.ctx, cs.db, scan); err != nil {
		return nil, err
	}

	// Signal the worker to process the job
	select {
	case cs.jobSignal <- true:
	default:
	}

	log.Info().Str("path", course.Path).Msg("added scan job")

	return scan, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Worker processes jobs out of the DB sequentially
func (cs *CourseScanner) Worker(processor func(*CourseScanner, *models.Scan) error) {
	log.Info().Msg("Started course scanner worker")

	for {
		<-cs.jobSignal
		for {
			// Get the next scan job
			scanJob, err := models.NextScan(cs.ctx, cs.db)
			if err != nil {
				log.Error().Err(err).Msg("error looking up next scan job")
				break
			} else if scanJob == nil {
				log.Info().Msg("finished processing all scan jobs")
				break
			}

			log.Info().Str("job", scanJob.ID).Str("path", scanJob.Course.Path).Msg("processing scan job")

			err = processor(cs, scanJob)
			if err != nil {
				log.Error().Str("job", scanJob.ID).Err(err).Msg("error processing scan job")
			} else {
				log.Info().Str("job", scanJob.ID).Str("path", scanJob.Course.Path).Msg("finished processing scan job")
			}

			// Cleanup
			if _, err := models.DeleteScan(cs.ctx, cs.db, scanJob.ID); err != nil {
				log.Error().Str("job", scanJob.ID).Err(err).Msg("error deleting scan job")
				break
			}
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type assetMap map[string]map[int]*models.Asset
type attachmentMap map[string]map[int][]*models.Attachment

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseProcessor scans a course and finds assets and attachments
func CourseProcessor(cs *CourseScanner, scan *models.Scan) error {
	if scan == nil {
		return errors.New("scan cannot be empty")
	}

	// Set the scan status to processing
	if err := models.UpdateScanStatus(cs.ctx, cs.db, scan, types.ScanStatusProcessing); err != nil {
		return err
	}

	// Check the course still exists
	course, err := models.GetCourseById(cs.ctx, cs.db, nil, scan.CourseID)
	if err != nil {
		return err
	} else if course == nil {
		// This should never happen due to cascading deletes
		log.Debug().Str("course", scan.CourseID).Msg("ignoring scan job as the course no longer exists")
		return nil
	}

	cardPath := ""

	// Get all files down to a depth of 2 (immediate files and and files within 'chapters')
	files, err := cs.appFs.ReadDirFlat(course.Path, 2)
	if err != nil {
		return err
	}

	// A map to hold encountered assets and attachments by chapter and prefix
	assetsMap := assetMap{}
	attachmentsMap := attachmentMap{}

	for _, file := range files {
		// Get the fileName from the path (ex /path/to/file.txt -> file.txt)
		fileName := filepath.Base(file)

		// Get the fileDir from the path (ex /path/to/file.txt -> /path/to)
		fileDir := filepath.Dir(file)
		isAtRoot := fileDir == course.Path

		// Check if this file is a card. Only check when not yet set and the file is at the `root`
		// of this course
		if cardPath == "" && isAtRoot {
			if isCard(fileName) {
				cardPath = file
				continue
			}
		}

		// Get the chapter for this file. This will be empty if the file is at the the `root` of
		// this course
		chapter := ""
		if !isAtRoot {
			chapter = filepath.Base(fileDir)
		}

		// Create a new map entry for this chapter
		if _, exists := assetsMap[chapter]; !exists {
			assetsMap[chapter] = make(map[int]*models.Asset)
		}

		// Add chapter to attachments map
		if _, exists := attachmentsMap[chapter]; !exists {
			attachmentsMap[chapter] = make(map[int][]*models.Attachment)
		}

		// Determine if this file is an asset or attachment
		fileInfo := buildFileInfo(fileName)

		if fileInfo == nil {
			log.Debug().Str("file", file).Msg("ignoring file")
			continue
		}

		if fileInfo.isAsset {
			// File is an asset. Check if we have already encountered an asset with the same
			// prefix for this chapter
			existingAsset, exists := assetsMap[chapter][fileInfo.prefix]

			newAsset := &models.Asset{
				Title:    fileInfo.title,
				Prefix:   fileInfo.prefix,
				CourseID: course.ID,
				Chapter:  chapter,
				Path:     file,
				Type:     fileInfo.assetType,
			}

			if !exists {
				// Found a new asset for this chapter
				assetsMap[chapter][fileInfo.prefix] = newAsset
			} else {
				// Found an existing asset for this chapter with the same prefix
				if newAsset.Type.IsVideo() && !existingAsset.Type.IsVideo() ||
					newAsset.Type.IsHTML() && existingAsset.Type.IsPDF() {
					// This new asset has a higher priority than the existing asset. Update the
					// asset map with the new asset and set the existing asset as an attachment
					//
					//  - Video assets have a higher priority than pdf/html assets
					//  - Html assets have a higher priority than pdf assets
					log.Debug().Str("file", file).Str("existing file", existingAsset.Path).Msg("replacing existing asset")

					assetsMap[chapter][fileInfo.prefix] = newAsset

					// Add the existing asset as an attachment
					attachmentsMap[chapter][fileInfo.prefix] = append(attachmentsMap[chapter][fileInfo.prefix], &models.Attachment{
						Title:    existingAsset.Title + filepath.Ext(existingAsset.Path),
						Path:     existingAsset.Path,
						CourseID: course.ID,
					})
				} else {
					// This new asset has a lower priority than the existing asset so add it as an attachment
					attachmentsMap[chapter][fileInfo.prefix] = append(attachmentsMap[chapter][fileInfo.prefix], &models.Attachment{
						Title:    fileInfo.fullTitle,
						Path:     file,
						CourseID: course.ID,
					})
				}
			}
		} else {
			// File is an attachment
			attachmentsMap[chapter][fileInfo.prefix] = append(attachmentsMap[chapter][fileInfo.prefix], &models.Attachment{
				Title:    fileInfo.fullTitle,
				Path:     file,
				CourseID: course.ID,
			})
		}
	}

	// Update the card path for this course
	if course.CardPath != cardPath {
		if err := models.UpdateCourseCardPath(cs.ctx, cs.db, course, cardPath); err != nil {
			return err
		}
	}

	// Convert the assets map to a slice
	assets := make([]*models.Asset, 0, len(files))
	for _, chapterMap := range assetsMap {
		for _, asset := range chapterMap {
			assets = append(assets, asset)
		}
	}

	// Update the assets in DB. This will insert new assets and delete assets which no longer
	// exist. For assets inserts, the ID will not be populated
	err = updateAssets(cs.ctx, cs.db, course.ID, assets)
	if err != nil {
		return err
	}

	// Convert the attachments map to a slice
	attachments := []*models.Attachment{}
	for chapter, attachmentMap := range attachmentsMap {
		for prefix, potentialAttachments := range attachmentMap {
			// Only add attachments when there is an assert
			if asset, exists := assetsMap[chapter][prefix]; exists {
				for _, attachment := range potentialAttachments {
					attachment.AssetID = asset.ID
					attachments = append(attachments, attachment)
				}
			}
		}
	}

	// Update the attachments in DB. This will insert new attachments and delete attachments which
	// no longer exist
	err = updateAttachments(cs.ctx, cs.db, course.ID, attachments)
	if err != nil {
		return err
	}

	// Update the course updated_at now that the scan is complete
	err = models.UpdateCourse(cs.ctx, cs.db, course)
	if err != nil {
		return err
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// PRIVATE
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type fileInfo struct {
	prefix    int
	title     string
	ext       string
	fullTitle string
	assetType types.Asset
	isAsset   bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// buildFileInfo is a rudimentary way of determining if a file is an asset, an attachment, or
// neither based upon the file name
//
// An asset/attachment which match one of the following formats:
//
//   - `<prefix> - <title>.<ext>`
//   - `<prefix> <title>.<ext>`
//
// For example, `01 - My Title.avi` or `2 My Title.pdf`
//
// The <ext> is optional and when not present, the file is considered an attachment
//
// If the <ext> is present, it must match one of the supported asset types (video, html, pdf) to be
// considered an asset. If it does not match, the file is considered an attachment
func buildFileInfo(fileName string) *fileInfo {
	fileInfo := &fileInfo{}

	// Build the regex to extract the prefix and title from the fileName
	re := regexp.MustCompile(`^\s*(?P<Prefix>[0-9]+)[\s-]*(?P<Title>.*?)(?:\.\w+)?$`)

	// Parse. When there is no match, return nil so that this file is ignored
	matches := re.FindStringSubmatch(fileName)
	if len(matches) == 0 {
		return nil
	}

	// The prefix
	prefix := matches[re.SubexpIndex("Prefix")]

	// ensures that the prefix is a number. For example, this will turn 001 into 1
	fileInfo.prefix, _ = strconv.Atoi(prefix)
	fileInfo.title = matches[re.SubexpIndex("Title")]

	// When there is no title, this is not an asset
	if fileInfo.title == "" {
		fileInfo.isAsset = false
		fileInfo.fullTitle = fileName
		return fileInfo
	}

	// Get the extension from the fileName without the leading dot (ex file.txt -> txt)
	ext := filepath.Ext(fileName)
	if ext == "" {
		// No extension. This is not an asset
		fileInfo.isAsset = false
		fileInfo.fullTitle = fileInfo.title
	} else {
		fileInfo.ext = ext[1:]
		fileInfo.fullTitle = fileInfo.title + "." + fileInfo.ext
	}

	// Set whether this is an asset or attachment
	assetType := types.NewAsset(fileInfo.ext)
	if assetType == nil {
		fileInfo.isAsset = false
	} else {
		fileInfo.isAsset = true
		fileInfo.assetType = *assetType
	}

	return fileInfo
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// isCard returns true if the fileName is a card and the extension is supported. For example
// `card.jpg` or `card.png`
func isCard(fileName string) bool {
	// Get the extension. If there is no extension, return false
	ext := filepath.Ext(fileName)
	if ext == "" {
		return false
	}

	fileWithoutExt := fileName[:len(fileName)-len(ext)]
	if fileWithoutExt != "card" {
		return false
	}

	// Check if the extension is supported
	switch ext[1:] {
	case
		"jpg",
		"jpeg",
		"png",
		"webp",
		"tiff":
		return true
	}

	return false
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func updateAssets(ctx context.Context, db database.Database, courseId string, assets []*models.Asset) error {
	// Get existing assets for this course
	existingAssets, err := models.GetAssetsByCourseId(ctx, db, nil, courseId)
	if err != nil {
		return err
	}

	// Compare the assets found with what is current in the DB. This will determine what needs to
	// be added and deleted
	toAdd, toDelete := utils.StructDiffer(assets, existingAssets, "Path")

	for _, asset := range toAdd {
		err := models.CreateAsset(ctx, db, asset)
		if err != nil {
			log.Err(err).Str("path", asset.Path).Msg("error creating asset")
			return err
		}
	}

	for _, asset := range toDelete {
		_, err := models.DeleteAsset(ctx, db, asset.ID)
		if err != nil {
			log.Err(err).Str("path", asset.Path).Msg("error deleting asset")
			return err
		}
	}

	// Set the asset id for those not added. This is required by potential attachments
	for _, asset := range assets {
		if asset.ID == "" {
			for _, existingAsset := range existingAssets {
				if asset.Path == existingAsset.Path {
					asset.ID = existingAsset.ID
				}
			}
		}
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func updateAttachments(ctx context.Context, db database.Database, courseId string, attachments []*models.Attachment) error {
	// Get existing attachments for this course
	existingAttachments, err := models.GetAttachmentsByCourseId(ctx, db, nil, courseId)
	if err != nil {
		return err
	}

	// Compare the attachments found with what is current in the DB. This will determine what needs to
	// be added and deleted
	toAdd, toDelete := utils.StructDiffer(attachments, existingAttachments, "Path")

	for _, attachment := range toAdd {
		err := models.CreateAttachment(ctx, db, attachment)
		if err != nil {
			log.Err(err).Str("path", attachment.Path).Msg("error creating attachment")
			return err
		}
	}

	for _, attachment := range toDelete {
		_, err := models.DeleteAttachment(ctx, db, attachment.ID)
		if err != nil {
			log.Err(err).Str("path", attachment.Path).Msg("error deleting attachment")
			return err
		}
	}

	return nil
}
