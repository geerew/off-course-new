package coursescan

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
)

var (
	loggerType = slog.Any("type", types.LogTypeCourseScan)
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseScanProcessorFn is a function that processes a course scan job
type CourseScanProcessorFn func(context.Context, *CourseScan, *models.Scan) error

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseScan scans a course and finds assets and attachments
type CourseScan struct {
	appFs     *appFs.AppFs
	db        database.Database
	dao       *dao.DAO
	logger    *slog.Logger
	jobSignal chan bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseScanConfig is the config for a CourseScan
type CourseScanConfig struct {
	Db     database.Database
	AppFs  *appFs.AppFs
	Logger *slog.Logger
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewCourseScan creates a new CourseScan
func NewCourseScan(config *CourseScanConfig) *CourseScan {
	return &CourseScan{
		appFs:     config.AppFs,
		db:        config.Db,
		dao:       dao.NewDAO(config.Db),
		logger:    config.Logger,
		jobSignal: make(chan bool, 1),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Add inserts a course scan job into the db
func (s *CourseScan) Add(ctx context.Context, courseId string) (*models.Scan, error) {
	// Check if the course exists
	course := &models.Course{Base: models.Base{ID: courseId}}
	err := s.dao.GetById(ctx, course)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrInvalidId
		}

		return nil, err
	}

	// Do nothing when a scan job is already in progress
	if course.ScanStatus.IsWaiting() || course.ScanStatus.IsProcessing() {
		s.logger.Debug(
			"Scan already in progress",
			loggerType,
			slog.String("path", course.Path),
		)

		// Get the scan from the db and return that
		scan := &models.Scan{}
		err := s.dao.Get(ctx, scan, &database.Options{Where: squirrel.Eq{scan.Table() + ".course_id": courseId}})
		if err != nil {
			return nil, err
		}

		return scan, nil
	}

	// Add the job
	scan := &models.Scan{CourseID: courseId, Status: types.NewScanStatusWaiting()}
	if err := s.dao.CreateScan(ctx, scan); err != nil {
		return nil, err
	}

	// Signal the worker to process the job
	select {
	case s.jobSignal <- true:
	default:
	}

	s.logger.Info(
		"Added scan job",
		loggerType,
		slog.String("path", course.Path),
	)

	return scan, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Worker processes jobs out of the DB sequentially
func (s *CourseScan) Worker(ctx context.Context, processorFn CourseScanProcessorFn, processingDone chan bool) {
	s.logger.Debug("Started course scanner worker", loggerType)

	for {
		<-s.jobSignal

		// Keep process jobs from the scans table until there are no more jobs
		for {
			nextScan := &models.Scan{}
			err := s.dao.NextWaitingScan(ctx, nextScan)
			if err != nil {
				// Nothing more to process
				if err == sql.ErrNoRows {
					s.logger.Debug("Finished processing all scan jobs", loggerType)
					break
				}

				// Error
				s.logger.Error(
					"Failed to look up the next scan job",
					loggerType,
					slog.String("error", err.Error()),
				)

				break
			}

			s.logger.Info(
				"Processing scan job",
				loggerType,
				slog.String("job", nextScan.ID),
				slog.String("path", nextScan.CoursePath),
			)

			err = processorFn(ctx, s, nextScan)
			if err != nil {
				s.logger.Error(
					"Failed to process scan job",
					loggerType,
					slog.String("error", err.Error()),
					slog.String("path", nextScan.CoursePath),
				)
			}

			// Cleanup
			if err := s.dao.Delete(ctx, nextScan, nil); err != nil {
				s.logger.Error(
					"Failed to delete scan job",
					loggerType,
					slog.String("error", err.Error()),
					slog.String("job", nextScan.ID),
				)

				break
			}
		}

		// Signal that processing is done
		if processingDone != nil {
			processingDone <- true
		}

		// Clear any pending signal that were sent while processing
		select {
		case <-s.jobSignal:
		default:
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type assetMap map[string]map[int]*models.Asset
type attachmentMap map[string]map[int][]*models.Attachment

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Processor scans a course to identify assets and attachments
func Processor(ctx context.Context, s *CourseScan, scan *models.Scan) error {
	if scan == nil {
		return ErrNilScan
	}

	// Set the scan status to processing
	scan.Status.SetProcessing()
	err := s.dao.UpdateScan(ctx, scan)
	if err != nil {
		return err
	}

	// Get the course for this scan
	course := &models.Course{}
	err = s.dao.Get(ctx, course, &database.Options{Where: squirrel.Eq{course.Table() + ".id": scan.CourseID}})
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Debug(
				"Ignoring scan job as the course no longer exists",
				loggerType,
				slog.String("path", scan.CoursePath),
			)

			return nil
		}

		return err
	}

	// Check availability and skip when unavailable. Also marks the course as unavailable
	_, err = s.appFs.Fs.Stat(course.Path)
	if err != nil {
		if os.IsNotExist(err) {
			s.logger.Debug(
				"Skipping as the course is unavailable",
				loggerType,
				slog.String("path", scan.CoursePath),
			)

			if course.Available {
				course.Available = false
				err = s.dao.UpdateCourse(ctx, course)
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
		err := s.dao.UpdateCourse(ctx, course)
		if err != nil {
			return err
		}

		s.logger.Debug(
			"Setting unavailable course as available",
			loggerType,
			slog.String("path", scan.CoursePath),
		)
	}

	cardPath := ""

	// Get all files down to a depth of 2
	files, err := s.appFs.ReadDirFlat(course.Path, 2)
	if err != nil {
		return err
	}

	// Maps to hold assets and attachments by [chapter][prefix]
	assetsMap := assetMap{}
	attachmentsMap := attachmentMap{}

	for _, fp := range files {
		normalizedPath := utils.NormalizeWindowsDrive(fp)
		filename := filepath.Base(normalizedPath)
		fileDir := filepath.Dir(normalizedPath)
		isInRoot := fileDir == utils.NormalizeWindowsDrive(course.Path)

		// Check if this file is  the course card
		if isInRoot && isCard(filename) {
			if cardPath != "" {
				s.logger.Debug(
					"Found another course card. Ignoring",
					loggerType,
					slog.String("file", filename),
					slog.String("path", scan.CoursePath),
				)
			} else {
				cardPath = normalizedPath
			}

			continue
		}

		// Set the chapter. This will be empty when the file is in the root directory
		chapter := ""
		if !isInRoot {
			chapter = filepath.Base(fileDir)
		}

		if _, exists := assetsMap[chapter]; !exists {
			assetsMap[chapter] = make(map[int]*models.Asset)
		}

		if _, exists := attachmentsMap[chapter]; !exists {
			attachmentsMap[chapter] = make(map[int][]*models.Attachment)
		}

		pfn := parseFilename(filename)

		// Ignore files that are neither assets nor attachments
		if pfn == nil {
			s.logger.Debug(
				"Incompatible file name. Ignoring",
				loggerType,
				slog.String("path", scan.CoursePath),
				slog.String("file", normalizedPath),
			)

			continue
		}

		// Add attachment
		if pfn.asset == nil {
			attachmentsMap[chapter][pfn.prefix] = append(
				attachmentsMap[chapter][pfn.prefix],
				&models.Attachment{
					Title: pfn.title,
					Path:  normalizedPath,
				},
			)

			continue
		}

		newAsset := &models.Asset{
			Title:    pfn.title,
			Prefix:   sql.NullInt16{Int16: int16(pfn.prefix), Valid: true},
			CourseID: course.ID,
			Chapter:  chapter,
			Path:     normalizedPath,
			Type:     *pfn.asset,
		}

		existing, exists := assetsMap[chapter][pfn.prefix]

		if !exists {
			// Add asset
			hash, err := s.appFs.PartialHash(normalizedPath, 1024*1024)
			if err != nil {
				return err
			}
			newAsset.Hash = hash

			assetsMap[chapter][pfn.prefix] = newAsset
		} else {
			// Check if this new asset has a higher priority than the existing asset. The priority
			// is video > html > pdf
			if newAsset.Type.IsVideo() && !existing.Type.IsVideo() ||
				newAsset.Type.IsHTML() && existing.Type.IsPDF() {

				// Demote the existing asset to an attachment and add the new asset
				s.logger.Debug(
					"Found a higher priority asset. Replacing",
					loggerType,
					slog.String("path", scan.CoursePath),
					slog.String("file", normalizedPath),
				)

				hash, err := s.appFs.PartialHash(normalizedPath, 1024*1024)
				if err != nil {
					return err
				}
				newAsset.Hash = hash

				assetsMap[chapter][pfn.prefix] = newAsset

				attachmentsMap[chapter][pfn.prefix] = append(
					attachmentsMap[chapter][pfn.prefix],
					&models.Attachment{
						Title: existing.Title + filepath.Ext(existing.Path),
						Path:  existing.Path,
					},
				)
			} else {
				// Add the new asset as an attachment
				attachmentsMap[chapter][pfn.prefix] = append(
					attachmentsMap[chapter][pfn.prefix],
					&models.Attachment{
						Title: pfn.title,
						Path:  normalizedPath,
					},
				)
			}
		}
	}

	course.CardPath = cardPath

	return s.db.RunInTransaction(ctx, func(txCtx context.Context) error {
		// Convert the assets map to a slice
		assets := make([]*models.Asset, 0, len(files))
		for _, chapterMap := range assetsMap {
			for _, asset := range chapterMap {
				assets = append(assets, asset)
			}
		}

		// Update the assets in DB
		if len(assets) > 0 {
			err = updateAssets(txCtx, s.dao, course.ID, assets)
			if err != nil {
				return err
			}
		}

		ids, err := s.dao.ListPluck(txCtx, &models.Asset{}, &database.Options{Where: squirrel.Eq{models.ASSET_TABLE + ".course_id": course.ID}}, models.BASE_ID)
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

		// Update the attachments in DB
		if len(attachments) > 0 {
			err = updateAttachments(txCtx, s.dao, ids, attachments)
			if err != nil {
				return err
			}
		}

		err = s.dao.UpdateCourse(txCtx, course)
		if err != nil {
			return err
		}

		return nil
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// PRIVATE
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// parsedFilename that holds information following a filename being parsed
type parsedFilename struct {
	prefix int
	title  string
	asset  *types.Asset
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
var filenameRegex = regexp.MustCompile(`^\s*(?P<Prefix>[0-9]+)((?:\s+-+\s+|\s+-+|\s+|-+\s*)(?P<Title>[^.][^.]*)?)?(?:\.(?P<Ext>\w+))?$`)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// parseFilename parses a file name and determines if it represents an asset, attachment, or neither
//
// A file is an asset when it matches `<prefix> <title>.<ext>` and <ext> is a valid `types.AssetType`
//
// A file is an attachment when it has a <prefix>, and optionally a <title> and/or <ext>, whereby <ext>
// is not a valid `types.AssetType`
//
// When a file is neither an asset nor an attachment, nil is returned
func parseFilename(filename string) *parsedFilename {
	pfn := &parsedFilename{}

	matches := filenameRegex.FindStringSubmatch(filename)
	if len(matches) == 0 {
		return nil
	}

	prefix, err := strconv.Atoi(matches[filenameRegex.SubexpIndex("Prefix")])
	if err != nil {
		return nil
	}

	pfn.prefix = prefix
	pfn.title = matches[filenameRegex.SubexpIndex("Title")]

	// When title is empty, consider this an attachment
	if pfn.title == "" {
		pfn.title = filename
		return pfn
	}

	// Where there is no extension, consider this an attachment
	ext := matches[filenameRegex.SubexpIndex("Ext")]
	if ext == "" {
		return pfn
	}

	pfn.asset = types.NewAsset(ext)

	// When the extension is not supported, consider this an attachment
	if pfn.asset == nil {
		pfn.title = pfn.title + "." + ext
	}

	return pfn
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// isCard determines if a given file name represents a card based on its name and extension
func isCard(filename string) bool {
	// Get the extension. If there is no extension, return false
	ext := filepath.Ext(filename)
	if ext == "" {
		return false
	}

	fileWithoutExt := filename[:len(filename)-len(ext)]
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
func updateAssets(ctx context.Context, dao *dao.DAO, courseId string, assets []*models.Asset) error {
	existingAssets := []*models.Asset{}
	err := dao.List(ctx, &existingAssets, &database.Options{Where: squirrel.Eq{models.ASSET_TABLE + ".course_id": courseId}})
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
		if err := dao.CreateAsset(ctx, asset); err != nil {
			return err
		}
	}

	// Delete assets
	for _, deleteAsset := range toDelete {
		err := dao.Delete(ctx, deleteAsset, nil)
		if err != nil {
			return err
		}
	}

	// Identify the existing assets whose information has changed
	existingAssetsMap := make(map[string]*models.Asset)
	for _, existingAsset := range existingAssets {
		existingAssetsMap[existingAsset.Hash] = existingAsset
	}

	randomTempSuffix := security.RandomString(10)
	updatedAssets := make([]*models.Asset, 0, len(assets))

	// On the first pass we update the existing assets with details of the new asset. In addition, we
	// set the path to be path+randomTempSuffix. This is to prevent a `unique path constraint` error if,
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
				if err := dao.UpdateAsset(ctx, asset); err != nil {
					return err
				}
			}
		}
	}

	for _, asset := range updatedAssets {
		asset.Path = asset.Path[:len(asset.Path)-len(randomTempSuffix)]

		if err := dao.UpdateAsset(ctx, asset); err != nil {
			return err
		}
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// updateAttachments updates the attachments in the database based on the attachments found on disk.
// It compares the existing attachments in the database with the attachments found on disk, and performs
// the necessary additions and deletions
func updateAttachments(ctx context.Context, dao *dao.DAO, assetIDs []string, attachments []*models.Attachment) error {
	existingAttachments := []*models.Attachment{}
	err := dao.List(ctx, &existingAttachments, &database.Options{Where: squirrel.Eq{models.ATTACHMENT_TABLE + ".asset_id": assetIDs}})
	if err != nil {
		return err
	}

	// Compare the attachments found on disk to attachments found in DB
	toAdd, toDelete, err := utils.DiffSliceOfStructsByKey(attachments, existingAttachments, "Path")
	if err != nil {
		return err
	}

	// Add attachments
	for _, attachment := range toAdd {
		if err := dao.CreateAttachment(ctx, attachment); err != nil {
			return err
		}
	}

	// Delete attachments
	for _, attachment := range toDelete {
		err := dao.Delete(ctx, attachment, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
