package jobs

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
)

var (
	loggerType = slog.Any("type", types.LogTypeCourseScanner)
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

		// Get the scan from the db and return that
		scan, err := cs.scanDao.Get(courseId, nil)
		if err != nil {
			return nil, err
		}

		return scan, nil
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
					loggerType,
					slog.String("error", err.Error()),
					slog.String("path", job.CoursePath),
				)
			}

			// Cleanup
			if err := cs.scanDao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": job.ID}}, nil); err != nil {
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

	for _, fp := range files {
		normalizedFilePath := utils.NormalizeWindowsDrive(fp)
		// Get the fileName from the path (ex /path/to/file.txt -> file.txt)
		fileName := filepath.Base(normalizedFilePath)

		// Get the fileDir from the path (ex /path/to/file.txt -> /path/to)
		fileDir := filepath.Dir(normalizedFilePath)
		isRootDir := fileDir == course.Path

		// Check if this file is a card. Only check when not yet set and the file exists at the
		// course root
		fmt.Println("cardPath", cardPath, "isRootDir", isRootDir, "fileName", fileName)
		if cardPath == "" && isRootDir {
			if isCard(fileName) {
				cardPath = normalizedFilePath
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
				slog.String("file", normalizedFilePath),
			)

			continue
		}

		if pfn.asset != nil {
			// Check if we have an existing asset for this [chapter][prefix]
			existing, exists := assetsMap[chapter][pfn.prefix]

			// Get a (partial) hash of the asset
			hash, err := cs.appFs.PartialHash(normalizedFilePath, 1024*1024)
			if err != nil {
				return err
			}

			newAsset := &models.Asset{
				Title:    pfn.title,
				Prefix:   sql.NullInt16{Int16: int16(pfn.prefix), Valid: true},
				CourseID: course.ID,
				Chapter:  chapter,
				Path:     normalizedFilePath,
				Type:     *pfn.asset,
				Hash:     hash,
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
						slog.String("file", normalizedFilePath),
					)

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
						Path:     normalizedFilePath,
						CourseID: course.ID,
					})
				}
			}
		} else {
			// Attachment
			attachmentsMap[chapter][pfn.prefix] = append(attachmentsMap[chapter][pfn.prefix], &models.Attachment{
				Title:    pfn.attachmentTitle,
				Path:     normalizedFilePath,
				CourseID: course.ID,
			})
		}
	}

	course.CardPath = cardPath

	// Run in a transaction so it all commits, or it rolls back
	err = cs.db.RunInTransaction(func(tx *database.Tx) error {
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

// parseFileName parses a file name and determines if it represents an asset, attachment, or neither
//
// A file is an asset when it matches `<prefix> <title>.<ext>` and <ext> is a valid `types.AssetType`
//
// A file is an attachment when it has a <prefix>, and optionally a <title> and/or <ext>, whereby <ext>
// is not a valid `types.AssetType`
//
// Parameters:
// - fileName: The name of the file to parse
//
// Returns:
//   - *parsedFileName: A struct containing parsed information if the file name matches the expected
//     format, otherwise nil
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

// isCard determines if a given file name represents a card based on its name and extension
//
// Parameters:
// - fileName: The name of the file to check
//
// Returns:
// - bool: True if the file name represents a card, false otherwise
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

// updateAssets updates the assets in the database based on the assets found on disk. It compares
// the existing assets in the database with the assets found on disk, and performs the necessary
// additions, deletions, and updates
//
// Parameters:
// - assetDao: The DAO used to interact with the assets table in the database
// - tx: The database transaction within which all operations should be performed
// - courseId: The ID of the course to which the assets belong
// - assets: A slice of Asset structs representing the assets found on disk
//
// Returns:
// - error: An error if any operation fails, otherwise nil
func updateAssets(assetDao *daos.AssetDao, tx *database.Tx, courseId string, assets []*models.Asset) error {
	// Get existing assets
	dbParams := &database.DatabaseParams{
		Where: squirrel.Eq{assetDao.Table() + ".course_id": courseId},
	}

	existingAssets, err := assetDao.List(dbParams, tx)
	if err != nil {
		return err
	}

	// Compare the assets found on disk to assets found in DB and identify which assets to add and
	// which assets to delete
	toAdd, toDelete, err := utils.DiffSliceOfStructsByKey(assets, existingAssets, "Hash")
	if err != nil {
		return err
	}

	// Add assets
	// TODO: This could be optimized by using a bulk insert
	for _, asset := range toAdd {
		if err := assetDao.Create(asset, tx); err != nil {
			return err
		}
	}

	// Bulk delete assets
	whereClause := squirrel.Or{}
	for _, deleteAsset := range toDelete {
		whereClause = append(whereClause, squirrel.Eq{"id": deleteAsset.ID})
	}

	if err := assetDao.Delete(&database.DatabaseParams{Where: whereClause}, tx); err != nil {
		return err
	}

	// Identify the existing assets whose information has changed
	existingAssetsMap := make(map[string]*models.Asset)
	for _, existingAsset := range existingAssets {
		existingAssetsMap[existingAsset.Hash] = existingAsset
	}

	randomTempSuffix := security.RandomString(10)
	updatedAssets := make([]*models.Asset, 0, len(assets))

	// On the first pass we update the existing assets with details of the new asset. In addition, we
	// set the path to be path+ranomTempSuffix. This is to prevent a `unique path constraint` error if,
	// for example, 2 files are have their titles swapped.
	//
	// On the second pass we update the existing assets and remove the randomTempSuffix from the path
	for _, asset := range assets {
		if existingAsset, exists := existingAssetsMap[asset.Hash]; exists {
			asset.ID = existingAsset.ID

			if !utils.CompareStructs(asset, existingAsset, []string{"CreatedAt", "UpdatedAt"}) {
				asset.Path = asset.Path + randomTempSuffix
				updatedAssets = append(updatedAssets, asset)

				// The assets has been updated to have the existing assets ID, so this will update the
				// existing asset with the details of the new asset
				if err := assetDao.Update(asset, tx); err != nil {
					return err
				}
			}
		}
	}

	for _, asset := range updatedAssets {
		asset.Path = asset.Path[:len(asset.Path)-len(randomTempSuffix)]

		if err := assetDao.Update(asset, tx); err != nil {
			return err
		}
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// updateAttachments updates the attachments in the database based on the attachments found on disk.
// It compares the existing attachments in the database with the attachments found on disk, and performs
// the necessary additions and deletions
//
// Parameters:
// - attachmentDao: The DAO used to interact with the attachments table in the database
// - tx: The database transaction within which all operations should be performed
// - courseId: The ID of the course to which the attachments belong
// - attachments: A slice of Attachment structs representing the attachments found on disk
//
// Returns:
// - error: An error if any operation fails, otherwise nil
func updateAttachments(attachmentDao *daos.AttachmentDao, tx *database.Tx, courseId string, attachments []*models.Attachment) error {
	// Get existing attachments
	dbParams := &database.DatabaseParams{
		Where: squirrel.Eq{attachmentDao.Table() + ".course_id": courseId},
	}

	existingAttachments, err := attachmentDao.List(dbParams, tx)
	if err != nil {
		return err
	}

	// Compare the attachments found on disk to attachments found in DB
	toAdd, toDelete, err := utils.DiffSliceOfStructsByKey(attachments, existingAttachments, "Path")
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
		err := attachmentDao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": attachment.ID}}, tx)
		if err != nil {
			return err
		}
	}

	return nil
}
