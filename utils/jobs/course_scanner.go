package jobs

import (
	"context"
	"database/sql"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/rs/zerolog/log"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseScanner scans a course and finds assets and attachments
type CourseScanner struct {
	db        database.Database
	appFs     *appFs.AppFs
	jobSignal chan bool
	ctx       context.Context
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
		jobSignal: make(chan bool, 1),
		ctx:       Ctx,
		finished:  false,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Add inserts a course scan job into the db
func (cs *CourseScanner) Add(id string) (*models.Scan, error) {
	// Ensure a job does not already exists for this course
	dbParams := &database.DatabaseParams{
		Where:    []database.Where{{Column: "course_id", Value: id}},
		Relation: []database.Relation{{Struct: "Course"}},
	}
	scan, err := models.GetScan(cs.db, dbParams, ctx)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	} else if scan != nil {
		log.Debug().Str("path", scan.Course.Path).Msg("scan job already exists")
		return nil, nil
	}

	// Get the course
	dbParams = &database.DatabaseParams{Where: []database.Where{{Column: "course.id", Value: id}}}
	course, err := models.GetCourse(cs.db, dbParams, ctx)
	if err != nil {
		return nil, err
	}

	// Add the job
	scan = &models.Scan{CourseID: course.ID}
	if err := models.CreateScan(cs.db, scan, ctx); err != nil {
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
func (cs *CourseScanner) Worker(processor func(*models.Scan, database.Database, *appFs.AppFs) error) {
	log.Info().Msg("Started course scanner worker")

	for {
		<-cs.jobSignal
		for {
			// Get the next scan job
			scanJob, err := models.NextScan(cs.db)
			if err != nil {
				log.Error().Err(err).Msg("error looking up next scan job")
				break
			} else if scanJob == nil {
				log.Info().Msg("finished processing all scan jobs")
				break
			}

			log.Info().Str("job", scanJob.ID).Str("path", scanJob.Course.Path).Msg("processing scan job")

			err = processor(scanJob, cs.db, cs.appFs)
			if err != nil {
				log.Error().Str("job", scanJob.ID).Err(err).Msg("error processing scan job")

				// Cleanup
				if err := models.DeleteScan(cs.db, scanJob.ID); err != nil {
					log.Error().Str("job", scanJob.ID).Err(err).Msg("error deleting scan job")
				}

				break
			}

			log.Info().Str("job", scanJob.ID).Str("path", scanJob.Course.Path).Msg("finished processing scan job")

			// Cleanup
			if err := models.DeleteScan(cs.db, scanJob.ID); err != nil {
				log.Error().Str("job", scanJob.ID).Err(err).Msg("error deleting scan job")
				break
			}
		}
	}
}

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// // CourseProcessor scans a course and finds assets and attachments
// func CourseProcessor(scan *models.Scan, db database.Database, appFs *appFs.AppFs) error {
// 	if scan == nil {
// 		return errors.New("scan job cannot be empty")
// 	}

// 	// Set the scan status to processing
// 	if err := models.UpdateScanStatus(db, scan, types.ScanStatusProcessing); err != nil {
// 		return err
// 	}

// 	// Get the course
// 	course, err := models.GetCourse(db, scan.CourseID, nil)
// 	if err != nil {
// 		return err
// 	} else if course == nil {
// 		// The Course no longer exists. It was probably deleted.
// 		return nil
// 	}

// 	// Scan the course for assets/attachments and insert to DB
// 	if err := scanCourse(course, db, appFs); err != nil {
// 		return err
// 	}

// 	return nil
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// // PRIVATE
// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// type AssetMap map[string]map[int]*models.Asset

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func scanCourse(course *models.Course, db database.Database, appFs *appFs.AppFs) error {
// 	if course == nil {
// 		return errors.New("course cannot be empty")
// 	}

// 	cardPath := ""

// 	// Get all files with a depth of 2. This will include immediate files and and files within
// 	// chapters
// 	files, err := appFs.ReadDirFlat(course.Path, 2)
// 	if err != nil {
// 		return err
// 	}

// 	// This map stores encountered assets by chapter and prefix
// 	assetsMap := AssetMap{}

// 	for _, file := range files {
// 		// Get the filename from the path (ex /path/to/file.txt -> file.txt)
// 		filename := filepath.Base(file)

// 		// First check if this is a card but only if a card has not yet been set and the current
// 		// file is at the root of the course
// 		if cardPath == "" && filepath.Dir(file) == course.Path {
// 			if isCard(filename) {
// 				cardPath = file
// 				continue
// 			}
// 		}

// 		// Get the chapter. This will remain empty when the file is in the root of the course
// 		chapter := ""
// 		if filepath.Dir(file) != course.Path {
// 			chapter = filepath.Base(filepath.Dir(file))
// 		}

// 		if _, exists := assetsMap[chapter]; !exists {
// 			assetsMap[chapter] = make(map[int]*models.Asset)
// 		}

// 		fileInfo := buildFileInfo(filename)

// 		// Ignore this file when it is not an asset or attachment
// 		if fileInfo == nil {
// 			log.Debug().Str("file", file).Msg("ignoring file")
// 			continue
// 		}

// 		if fileInfo.isAsset {
// 			// Asset
// 			existingAsset, exists := assetsMap[chapter][fileInfo.prefix]

// 			asset := &models.Asset{
// 				Title:    fileInfo.title,
// 				Prefix:   fileInfo.prefix,
// 				CourseID: course.ID,
// 				Chapter:  chapter,
// 				Path:     file,
// 				Type:     fileInfo.assetType,
// 			}

// 			if !exists {
// 				// New asset
// 				assetsMap[chapter][fileInfo.prefix] = asset
// 			} else if fileInfo.assetType.IsVideo() && !existingAsset.Type.IsVideo() ||
// 				fileInfo.assetType.IsHTML() && !existingAsset.Type.IsHTML() {
// 				// This new assert is trumping another asset. For example, this new asset is a
// 				// video file whereas the existing is a pdf file

// 				// TODO: Set existing asset as an attachment

// 				// Replace the existing asset with the new asset
// 				assetsMap[chapter][fileInfo.prefix] = asset
// 				log.Debug().Str("file", file).Str("existing file", existingAsset.Path).Msg("replacing existing asset")
// 			} else {
// 				// Duplicate
// 				log.Debug().Str("file", file).Msg("ignoring duplicate")
// 				continue
// 			}
// 		} else {
// 			// Attachment
// 		}

// 		// if assetType != nil {
// 		// 	if assetType.IsVideo() || assetType.IsHTML() {
// 		// 		// This is an asset
// 		// 		// Have we already encountered this asset?
// 		// 		// existingAsset, exists := assetsMap[chapter][assetType.pr]
// 		// 	}
// 		// } else {
// 		// 	// Is this an attachment
// 		// }

// 	}

// 	// Convert the assets map to a slice
// 	assets := make([]*models.Asset, 0, len(files))
// 	for _, chapterMap := range assetsMap {
// 		for _, asset := range chapterMap {
// 			assets = append(assets, asset)
// 		}
// 	}

// 	// Update the card path for this course
// 	if course.CardPath != cardPath {
// 		if err := models.UpdateCourseCardPath(db, course, cardPath); err != nil {
// 			return err
// 		}
// 	}

// 	err = updateAssets(db, course.ID, assets)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// type FileInfo struct {
// 	prefix    int
// 	title     string
// 	ext       string
// 	assetType types.Asset
// 	isAsset   bool
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// // buildFileInfo builds a `FileInfo` struct based upon the file name, which is in the format:
// // `<prefix> - <title>.<ext>` or `<prefix> <title>.<ext>`. The <ext> is optional.
// //
// // The main goal is to determine if this file is an asset, an attachment of an asset, or neither
// func buildFileInfo(filename string) *FileInfo {
// 	fileInfo := &FileInfo{}

// 	// Build the regex to extract the prefix and title from the filename
// 	re := regexp.MustCompile(`^\s*(?P<Prefix>[0-9]+)[\s-]*(?P<Title>.*?)(?:\.\w+)?$`)

// 	// Parse. When there is no match, return nil so that this file is ignored
// 	matches := re.FindStringSubmatch(filename)
// 	if len(matches) == 0 {
// 		return nil
// 	}

// 	// Convert the prefix to a number. We don't need to check for errors here because the regex
// 	// ensures that the prefix is a number. For example, this will turn 001 into 1
// 	fileInfo.prefix, _ = strconv.Atoi(matches[re.SubexpIndex("Prefix")])
// 	fileInfo.title = matches[re.SubexpIndex("Title")]

// 	// When the title is empty, return nil so that this file is ignored
// 	if fileInfo.title == "" {
// 		return nil
// 	}

// 	// Get the extension from the filename without the leading dot (ex file.txt -> txt)
// 	ext := filepath.Ext(filename)
// 	if ext == "" {
// 		// No extension. This is not an asset
// 		fileInfo.isAsset = false
// 		return fileInfo
// 	} else {
// 		fileInfo.ext = ext[1:]
// 	}

// 	// Set whether this is an asset or attachment
// 	assetType := types.NewAsset(fileInfo.ext)
// 	if assetType == nil {
// 		fileInfo.isAsset = false
// 	} else {
// 		fileInfo.isAsset = true
// 		fileInfo.assetType = *assetType
// 	}

// 	return fileInfo
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func isCard(filename string) bool {
// 	// Get the extension. If there is no extension, return false
// 	ext := filepath.Ext(filename)
// 	if ext == "" {
// 		return false
// 	}

// 	fileWithoutExt := filename[:len(filename)-len(ext)]
// 	if fileWithoutExt != "card" {
// 		return false
// 	}

// 	// Check if the extension is supported
// 	switch ext[1:] {
// 	case
// 		"jpg",
// 		"jpeg",
// 		"png",
// 		"webp",
// 		"tiff":
// 		return true
// 	}

// 	return false
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func updateAssets(db database.Database, courseId string, assets []*models.Asset) error {
// 	// Get existing assets for this course
// 	existingAssets, err := models.GetAssets(db, &database.DatabaseParams{Where: database.Where{"course_id": courseId}})
// 	if err != nil {
// 		return err
// 	}

// 	// Compare the assets found with what is current in the DB. This will determine what needs to
// 	// be added and deleted
// 	toAdd, toDelete := utils.StructDiffer(assets, existingAssets, "Path")

// 	for _, asset := range toAdd {
// 		err := models.CreateAsset(db, asset)
// 		if err != nil {
// 			log.Err(err).Str("path", asset.Path).Msg("error creating asset")
// 			return err
// 		}
// 	}

// 	for _, asset := range toDelete {
// 		err := models.DeleteAsset(db, asset.ID)
// 		if err != nil {
// 			log.Err(err).Str("path", asset.Path).Msg("error deleting asset")
// 			return err
// 		}
// 	}

// 	return nil
// }
