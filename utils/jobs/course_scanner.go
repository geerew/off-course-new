package jobs

import (
	"crypto/md5"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/types"
)

var (
	loggerType = slog.String("type", "course_scanner")
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseScannerProcessorFn is a function that processes a course scan job
type CourseScannerProcessorFn func(*CourseScanner, *models.Scan) error

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseScanner scans a course and finds assets and attachments
type CourseScanner struct {
	appFs  *appFs.AppFs
	db     database.Database
	logger *slog.Logger

	jobSignal chan bool

	// Required DAOs
	courseDao     *daos.CourseDao
	scanDao       *daos.ScanDao
	assetDao      *daos.AssetDao
	attachmentDao *daos.AttachmentDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseScannerConfig is the config for a CourseScanner
type CourseScannerConfig struct {
	Db     database.Database
	AppFs  *appFs.AppFs
	Logger *slog.Logger
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewCourseScanner creates a new CourseScanner
func NewCourseScanner(config *CourseScannerConfig) *CourseScanner {
	return &CourseScanner{
		appFs:         config.AppFs,
		db:            config.Db,
		logger:        config.Logger,
		jobSignal:     make(chan bool, 1),
		courseDao:     daos.NewCourseDao(config.Db),
		scanDao:       daos.NewScanDao(config.Db),
		assetDao:      daos.NewAssetDao(config.Db),
		attachmentDao: daos.NewAttachmentDao(config.Db),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Add inserts a course scan job into the db
func (cs *CourseScanner) Add(courseId string) (*models.Scan, error) {
	// Check if the course exists
	course, err := cs.courseDao.Get(courseId, nil, nil)
	if err != nil {
		return nil, err
	}

	// Do nothing when a scan job is already in progress
	if course.ScanStatus != "" {
		cs.logger.Debug(
			"Scan already in progress",
			loggerType,
			slog.String("path", course.Path),
		)
		return nil, nil
	}

	// Add the job
	scan := &models.Scan{CourseID: courseId, Status: types.NewScanStatus(types.ScanStatusWaiting)}
	if err := cs.scanDao.Create(scan, nil); err != nil {
		return nil, err
	}

	// Signal the worker to process the job
	select {
	case cs.jobSignal <- true:
	default:
	}

	cs.logger.Info(
		"Added scan job",
		loggerType,
		slog.String("path", course.Path),
	)

	return scan, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Worker processes jobs out of the DB sequentially
func (cs *CourseScanner) Worker(processor CourseScannerProcessorFn, processingDone chan bool) {
	cs.logger.Debug("Started course scanner worker", loggerType)

	for {
		<-cs.jobSignal
		for {
			// Get the next scan
			job, err := cs.scanDao.Next(nil)
			if err != nil {
				cs.logger.Error(
					"Failed to look up next scan job",
					loggerType,
					slog.String("error", err.Error()),
				)

				if processingDone != nil {
					processingDone <- true
				}

				break
			}

			if job == nil {
				if processingDone != nil {
					processingDone <- true
				}
				cs.logger.Debug("Finished processing all jobs", loggerType)
				break
			}

			cs.logger.Info(
				"Processing scan job",
				loggerType,
				slog.String("job", job.ID),
				slog.String("path", job.CoursePath),
			)

			// Process the job
			err = processor(cs, job)
			if err != nil {
				cs.logger.Error(
					"Failed to process scan job",
					slog.String("error", err.Error()),
				)
			}

			// Cleanup
			if err := cs.scanDao.Delete(&database.DatabaseParams{Where: sq.Eq{"id": job.ID}}, nil); err != nil {
				cs.logger.Error(
					"Failed to delete scan job",
					loggerType,
					slog.String("error", err.Error()),
					slog.String("job", job.ID),
				)

				if processingDone != nil {
					processingDone <- true
				}

				break
			}

		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// assetMap and attachmentMap are maps to hold encountered assets and attachments
type assetMap map[string]map[int]*models.Asset
type attachmentMap map[string]map[int][]*models.Attachment

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseProcessor scans a course and finds assets and attachments
func CourseProcessor(cs *CourseScanner, scan *models.Scan) error {
	if scan == nil {
		return errors.New("scan cannot be empty")
	}

	// Get the course for this scan
	course, err := cs.courseDao.Get(scan.CourseID, nil, nil)
	if err != nil {
		// If the course does not exist, we can ignore this job
		if err == sql.ErrNoRows {
			cs.logger.Debug(
				"Ignoring scan job as the course no longer exists",
				loggerType,
				slog.String("path", scan.CoursePath),
			)

			return nil
		}

		return err
	}

	// Check the availability of the course. When a course is unavailable, we do not want to scan
	// it. This prevents assets and attachments from being deleted unintentionally
	_, err = cs.appFs.Fs.Stat(course.Path)
	if err != nil {
		if os.IsNotExist(err) {
			cs.logger.Debug(
				"Ignoring scan job as the course path does not exist",
				loggerType,
				slog.String("path", scan.CoursePath),
			)

			if course.Available {
				course.Available = false
				err = cs.courseDao.Update(course, nil)
				if err != nil {
					return err
				}
			}

			return nil
		}

		return err
	}

	// If the course is currently marked as unavailable, set it as available
	if !course.Available {
		course.Available = true
		err := cs.courseDao.Update(course, nil)
		if err != nil {
			return err
		}
	}

	// Set the scan status to processing
	scan.Status = types.NewScanStatus(types.ScanStatusProcessing)
	err = cs.scanDao.Update(scan, nil)
	if err != nil {
		return err
	}

	cardPath := ""

	// Get all files down to a depth of 2 (immediate files and and files within 'chapters')
	files, err := cs.appFs.ReadDirFlat(course.Path, 2)
	if err != nil {
		return err
	}

	// Maps to hold encountered assets and attachments by [chapter][prefix]
	assetsMap := assetMap{}
	attachmentsMap := attachmentMap{}

	for _, filePath := range files {
		// Get the fileName from the path (ex /path/to/file.txt -> file.txt)
		fileName := filepath.Base(filePath)

		// Get the fileDir from the path (ex /path/to/file.txt -> /path/to)
		fileDir := filepath.Dir(filePath)
		isRootDir := fileDir == course.Path

		// Check if this file is a card. Only check when not yet set and the file exists at the
		// course root
		if cardPath == "" && isRootDir {
			if isCard(fileName) {
				cardPath = filePath
				continue
			}
		}

		// Get the chapter for this file. This will be empty when the file exists at the course
		// root
		chapter := ""
		if !isRootDir {
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

		// Parse the file name to see if it is an asset, attachment or neither
		pfn := parseFileName(fileName)

		if pfn == nil {
			cs.logger.Debug(
				"Ignoring file during scan job",
				loggerType,
				slog.String("path", scan.CoursePath),
				slog.String("file", filePath),
			)

			continue
		}

		// Generate an MD5 hash for the file
		file, err := cs.appFs.Fs.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		hash := md5.New()
		if _, err := io.Copy(hash, file); err != nil {
			return err
		}

		md5 := string(hash.Sum(nil))

		if pfn.asset != nil {
			// Check if we have an existing asset for this [chapter][prefix]
			existing, exists := assetsMap[chapter][pfn.prefix]

			newAsset := &models.Asset{
				Title:    pfn.title,
				Prefix:   sql.NullInt16{Int16: int16(pfn.prefix), Valid: true},
				CourseID: course.ID,
				Chapter:  chapter,
				Path:     filePath,
				Type:     *pfn.asset,
				Md5:      md5,
			}

			if !exists {
				// New asset
				assetsMap[chapter][pfn.prefix] = newAsset
			} else {
				// Found an existing asset. Check if this new asset has a higher priority than the
				// existing asset. The priority is video > html > pdf
				if newAsset.Type.IsVideo() && !existing.Type.IsVideo() ||
					newAsset.Type.IsHTML() && existing.Type.IsPDF() {
					// Asset -> Replace the existing asset with the new asset and set the existing
					// asset as an attachment

					cs.logger.Debug(
						"Replacing existing asset with new asset",
						loggerType,
						slog.String("path", scan.CoursePath),
						slog.String("file", filePath),
					)

					assetsMap[chapter][pfn.prefix] = newAsset

					attachmentsMap[chapter][pfn.prefix] = append(attachmentsMap[chapter][pfn.prefix], &models.Attachment{
						Title:    existing.Title + filepath.Ext(existing.Path),
						Path:     existing.Path,
						CourseID: course.ID,
						Md5:      existing.Md5,
					})
				} else {
					// Attachment -> This new asset has a lower priority than the existing asset
					attachmentsMap[chapter][pfn.prefix] = append(attachmentsMap[chapter][pfn.prefix], &models.Attachment{
						Title:    pfn.attachmentTitle,
						Path:     filePath,
						CourseID: course.ID,
						Md5:      md5,
					})
				}
			}
		} else {
			// Attachment
			attachmentsMap[chapter][pfn.prefix] = append(attachmentsMap[chapter][pfn.prefix], &models.Attachment{
				Title:    pfn.attachmentTitle,
				Path:     filePath,
				CourseID: course.ID,
				Md5:      md5,
			})
		}
	}

	course.CardPath = cardPath

	// Run in a transaction so it all commits, or it rolls back
	err = cs.db.RunInTransaction(func(tx *sql.Tx) error {
		// Convert the assets map to a slice
		assets := make([]*models.Asset, 0, len(files))
		for _, chapterMap := range assetsMap {
			for _, asset := range chapterMap {
				assets = append(assets, asset)
			}
		}

		// Update the assets in DB
		if len(assets) > 0 {
			err = updateAssets(cs.assetDao, tx, course.ID, assets)
			if err != nil {
				return err
			}
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

		// Update the attachments in DB
		if len(attachments) > 0 {
			err = updateAttachments(cs.attachmentDao, tx, course.ID, attachments)
			if err != nil {
				return err
			}
		}

		// Update the course (card_path, updated_at)
		if err = cs.courseDao.Update(course, tx); err != nil {
			return err
		}

		return nil
	})

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// PRIVATE
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// parsedFileName that holds information following a filename being parsed
type parsedFileName struct {
	prefix          int
	title           string
	ext             string
	attachmentTitle string
	asset           *types.Asset
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// A regex for parsing a file name into a prefix, title, and extension
//
// Valid patterns:
//
//	 `<prefix>`
//	 `<prefix>.<ext>`
//	 `<prefix> <title>`
//	 `<prefix>-<title>`
//	 `<prefix> - <title>`
//	 `<prefix> <title>.<ext>`
//	 `<prefix>-<title>.<ext>`
//	 `<prefix> - <title>.<ext>`
//
//	- <prefix> is required and must be a number
//	- A dash (-) is optional
//	- <title> is optional and can be any non-empty string
//	- <ext> is optional
var fileNameRegex = regexp.MustCompile(`^\s*(?P<Prefix>[0-9]+)((?:\s+-+\s+|\s+-+|\s+|-+\s*)(?P<Title>[^.][^.]*)?)?(?:\.(?P<Ext>\w+))?$`)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// parseFileName is a rudimentary function for determining if a file is an asset, attachment, or
// neither based upon a regex pattern match.
//
// A file is an asset when it has a <prefix>, <title>, <ext> and the <ext> is of type video, html,
// or pdf
//
// A file is an attachment when it has a <prefix>, and optionally a <title> and/or <ext>, with the
// <ext> not being of type video, html, or pdf
//
// Nil will be returned A file that does not match the regex is ignored
func parseFileName(fileName string) *parsedFileName {
	pfn := &parsedFileName{}

	// Match the file name against the regex and ignore if no match
	matches := fileNameRegex.FindStringSubmatch(fileName)
	if len(matches) == 0 {
		return nil
	}

	// Convert the prefix to an int and ignore if missing or not a number
	prefix, err := strconv.Atoi(matches[fileNameRegex.SubexpIndex("Prefix")])
	if err != nil {
		return nil
	}

	pfn.prefix = prefix
	pfn.title = matches[fileNameRegex.SubexpIndex("Title")]

	// Check the title. When empty, this is an attachment
	if pfn.title == "" {
		pfn.attachmentTitle = fileName
		return pfn
	}

	// Check the extension. When empty, this is an attachment
	pfn.ext = matches[fileNameRegex.SubexpIndex("Ext")]
	if pfn.ext == "" {
		pfn.attachmentTitle = pfn.title
		return pfn
	}

	// Set the attachment title, in the event that this is an attachment
	pfn.attachmentTitle = pfn.title + "." + pfn.ext

	// Check if this is a valid asset
	pfn.asset = types.NewAsset(pfn.ext)

	return pfn
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// isCard returns true if the fileName is `card` and the extension is supported
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

// updateAssets updates the assets in the DB (in a transaction), by comparing the assets found on disk
// to the assets found in the DB. It will insert new assets and delete assets which no longer exist
func updateAssets(assetDao *daos.AssetDao, tx *sql.Tx, courseId string, assets []*models.Asset) error {
	// Get existing assets
	dbParams := &database.DatabaseParams{
		Where: sq.Eq{assetDao.Table() + ".course_id": courseId},
	}

	existingAssets, err := assetDao.List(dbParams, tx)
	if err != nil {
		return err
	}

	// Compare the assets found on disk to assets found in DB
	toAdd, toDelete, err := utils.DiffStructs(assets, existingAssets, "Path")
	if err != nil {
		return err
	}

	// Add the missing assets
	for _, asset := range toAdd {
		if err := assetDao.Create(asset, tx); err != nil {
			return err
		}
	}

	// Delete the irrelevant assets
	for _, asset := range toDelete {
		err := assetDao.Delete(&database.DatabaseParams{Where: sq.Eq{"id": asset.ID}}, tx)
		if err != nil {
			return err
		}
	}

	// Assets that already exist in the DB will have an empty ID. We need to map the existing ID
	// to these assets, so that we can update the attachments with the correct asset ID later on.
	existingAssetsMap := make(map[string]*models.Asset)
	for _, asset := range existingAssets {
		existingAssetsMap[asset.Path] = asset
	}

	for _, asset := range assets {
		if asset.ID == "" {
			if existingAsset, exists := existingAssetsMap[asset.Path]; exists {
				asset.ID = existingAsset.ID
			}
		}
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// updateAttachments updates the attachments in the DB (in a transaction), by comparing the attachments
// found on disk to the attachments found in the DB. It will insert new attachments and delete attachments
// which no longer exist
func updateAttachments(attachmentDao *daos.AttachmentDao, tx *sql.Tx, courseId string, attachments []*models.Attachment) error {
	// Get existing attachments
	dbParams := &database.DatabaseParams{
		Where: sq.Eq{attachmentDao.Table() + ".course_id": courseId},
	}

	existingAttachments, err := attachmentDao.List(dbParams, tx)
	if err != nil {
		return err
	}

	// Compare the attachments found on disk to attachments found in DB
	toAdd, toDelete, err := utils.DiffStructs(attachments, existingAttachments, "Path")
	if err != nil {
		return err
	}

	// Add the missing attachments
	for _, attachment := range toAdd {
		if err := attachmentDao.Create(attachment, tx); err != nil {
			return err
		}
	}

	// Delete the irrelevant attachments
	for _, attachment := range toDelete {
		err := attachmentDao.Delete(&database.DatabaseParams{Where: sq.Eq{"id": attachment.ID}}, tx)
		if err != nil {
			return err
		}
	}

	return nil
}
