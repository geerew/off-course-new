package jobs

import (
	"database/sql"
	"errors"
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
	"github.com/rs/zerolog/log"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseScanner scans a course and finds assets and attachments
type CourseScanner struct {
	appFs     *appFs.AppFs
	jobSignal chan bool
	finished  bool

	// Required DAOs
	courseDao     *daos.CourseDao
	scanDao       *daos.ScanDao
	assetDao      *daos.AssetDao
	attachmentDao *daos.AttachmentDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseScannerConfig is the config for a CourseScanner
type CourseScannerConfig struct {
	Db    database.Database
	AppFs *appFs.AppFs
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewCourseScanner creates a new CourseScanner
func NewCourseScanner(config *CourseScannerConfig) *CourseScanner {
	return &CourseScanner{
		courseDao:     daos.NewCourseDao(config.Db),
		scanDao:       daos.NewScanDao(config.Db),
		assetDao:      daos.NewAssetDao(config.Db),
		attachmentDao: daos.NewAttachmentDao(config.Db),
		appFs:         config.AppFs,
		jobSignal:     make(chan bool, 1),
		finished:      false,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Add inserts a course scan job into the db
func (cs *CourseScanner) Add(courseId string) (*models.Scan, error) {

	// Check if the course exists
	course, err := cs.courseDao.Get(courseId)
	if err != nil {
		return nil, err
	}

	// Do nothing when a scan job is already in progress
	if course.ScanStatus != "" {
		log.Debug().Str("path", course.Path).Msg("scan already in progress")
		return nil, nil
	}

	// Add the job
	scan := &models.Scan{CourseID: courseId, Status: types.NewScanStatus(types.ScanStatusWaiting)}
	if err := cs.scanDao.Create(scan); err != nil {
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
			// Get the next scan
			nextScan, err := cs.scanDao.Next()
			if err != nil {
				log.Error().Err(err).Msg("error looking up next scan job")
				break
			} else if nextScan == nil {
				log.Info().Msg("finished processing all scan jobs")
				break
			}

			log.Info().Str("job", nextScan.ID).Str("path", nextScan.CoursePath).Msg("processing scan job")

			err = processor(cs, nextScan)
			if err != nil {
				log.Error().Str("job", nextScan.ID).Err(err).Msg("error processing scan job")
			} else {
				log.Info().Str("job", nextScan.ID).Str("path", nextScan.CoursePath).Msg("finished processing scan job")
			}

			// Cleanup
			if err := cs.scanDao.Delete(nextScan.ID); err != nil {
				log.Error().Str("job", nextScan.ID).Err(err).Msg("error deleting scan job")
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

	// Get the course for this scan
	course, err := cs.courseDao.Get(scan.CourseID)
	if err != nil {
		log.Debug().Str("course", scan.CourseID).Msg("ignoring scan job as the course no longer exists")
		return err
	}

	// Check the availability of the course. When a course is unavailable, we do not want to scan
	// it. This prevents assets and attachments from being deleted unintentionally
	_, err = cs.appFs.Fs.Stat(course.Path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Debug().Str("path", course.Path).Msg("ignoring scan job as the course path does not exist")

			if course.Available {
				course.Available = false
				err = cs.courseDao.Update(course)
				if err != nil {
					return err
				}
			}

			return errors.New("course unavailable")
		}

		return err
	}

	// If the course is currently marked as unavailable, set it as available
	if !course.Available {
		course.Available = true
		err := cs.courseDao.Update(course)
		if err != nil {
			return err
		}
	}

	// Set the scan status to processing
	scan.Status = types.NewScanStatus(types.ScanStatusProcessing)
	err = cs.scanDao.Update(scan)
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

	for _, file := range files {
		// Get the fileName from the path (ex /path/to/file.txt -> file.txt)
		fileName := filepath.Base(file)

		// Get the fileDir from the path (ex /path/to/file.txt -> /path/to)
		fileDir := filepath.Dir(file)
		isRootDir := fileDir == course.Path

		// Check if this file is a card. Only check when not yet set and the file exists at the
		// course root
		if cardPath == "" && isRootDir {
			if isCard(fileName) {
				cardPath = file
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
			log.Debug().Str("file", file).Msg("ignoring file")
			continue
		}

		if pfn.asset != nil {
			// Check if we have an existing asset for this [chapter][prefix]
			existing, exists := assetsMap[chapter][pfn.prefix]

			newAsset := &models.Asset{
				Title:    pfn.title,
				Prefix:   sql.NullInt16{Int16: int16(pfn.prefix), Valid: true},
				CourseID: course.ID,
				Chapter:  chapter,
				Path:     file,
				Type:     *pfn.asset,
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
					log.Debug().Str("file", file).Str("existing path", existing.Path).Msg("replacing existing asset")

					assetsMap[chapter][pfn.prefix] = newAsset

					attachmentsMap[chapter][pfn.prefix] = append(attachmentsMap[chapter][pfn.prefix], &models.Attachment{
						Title:    existing.Title + filepath.Ext(existing.Path),
						Path:     existing.Path,
						CourseID: course.ID,
					})
				} else {
					// Attachment -> This new asset has a lower priority than the existing asset
					attachmentsMap[chapter][pfn.prefix] = append(attachmentsMap[chapter][pfn.prefix], &models.Attachment{
						Title:    pfn.attachmentTitle,
						Path:     file,
						CourseID: course.ID,
					})
				}
			}
		} else {
			// Attachment
			attachmentsMap[chapter][pfn.prefix] = append(attachmentsMap[chapter][pfn.prefix], &models.Attachment{
				Title:    pfn.attachmentTitle,
				Path:     file,
				CourseID: course.ID,
			})
		}
	}

	course.CardPath = cardPath

	// Convert the assets map to a slice
	assets := make([]*models.Asset, 0, len(files))
	for _, chapterMap := range assetsMap {
		for _, asset := range chapterMap {
			assets = append(assets, asset)
		}
	}

	// Update the assets in DB
	if len(assets) > 0 {
		err = updateAssets(cs.assetDao, course.ID, assets)
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
		err = updateAttachments(cs.attachmentDao, course.ID, attachments)
		if err != nil {
			return err
		}
	}

	// Update the course (card_path, updated_at)
	if err = cs.courseDao.Update(course); err != nil {
		return err
	}

	return nil
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

// updateAssets updates the assets in the DB, by comparing the assets found on disk to the assets
// found in the DB. It will insert new assets and delete assets which no longer exist
func updateAssets(assetDao *daos.AssetDao, courseId string, assets []*models.Asset) error {

	// Get existing assets
	dbParams := &database.DatabaseParams{
		Where: sq.Eq{daos.TableAssets() + ".course_id": courseId},
	}

	existingAssets, err := assetDao.List(dbParams)
	if err != nil {
		return err
	}

	// Compare the assets found on disk to assets found in DB
	toAdd, toDelete := utils.DiffStructs(assets, existingAssets, "Path")

	// Add the missing assets
	for _, asset := range toAdd {
		err := assetDao.Create(asset)
		if err != nil {
			log.Err(err).Str("path", asset.Path).Msg("error creating asset")
			return err
		}
	}

	// Delete the irrelevant assets
	for _, asset := range toDelete {
		err := assetDao.Delete(asset.ID)
		if err != nil {
			log.Err(err).Str("path", asset.Path).Msg("error deleting asset")
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

// updateAttachments updates the attachments in the DB, by comparing the attachments found on disk
// to the attachments found in the DB. It will insert new attachments and delete attachments which
// no longer exist
func updateAttachments(attachmentDao *daos.AttachmentDao, courseId string, attachments []*models.Attachment) error {
	// Get existing attachments
	dbParams := &database.DatabaseParams{
		Where: sq.Eq{daos.TableAttachments() + ".course_id": courseId},
	}

	existingAttachments, err := attachmentDao.List(dbParams)
	if err != nil {
		return err
	}

	// Compare the attachments found on disk to attachments found in DB
	toAdd, toDelete := utils.DiffStructs(attachments, existingAttachments, "Path")

	// Add the missing attachments
	for _, attachment := range toAdd {
		err := attachmentDao.Create(attachment)
		if err != nil {
			log.Err(err).Str("path", attachment.Path).Msg("error creating attachment")
			return err
		}
	}

	// Delete the irrelevant attachments
	for _, attachment := range toDelete {
		err := attachmentDao.Delete(attachment.ID)
		if err != nil {
			log.Err(err).Str("path", attachment.Path).Msg("error deleting attachment")
			return err
		}
	}

	return nil
}
