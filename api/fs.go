package api

import (
	"path/filepath"

	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type fs struct {
	appFs     *appFs.AppFs
	courseDao *daos.CourseDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func bindFsApi(router fiber.Router, appFs *appFs.AppFs, db database.Database) {
	api := fs{
		appFs:     appFs,
		courseDao: daos.NewCourseDao(db),
	}

	subGroup := router.Group("/fileSystem")

	subGroup.Get("", api.fileSystem)
	subGroup.Get("/:path", api.path)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type fileSystemResponse struct {
	Count       int         `json:"count"`
	Directories []*fileInfo `json:"directories"`
	Files       []*fileInfo `json:"files"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type fileInfo struct {
	Title          string                   `json:"title"`
	Path           string                   `json:"path"`
	Classification types.PathClassification `json:"classification"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// fileSystem queries the underlying system for available drives
//
// Note: On WSL, the drives will consist of / and /mnt* (ignoring /mnt/wsl*)
func (api *fs) fileSystem(c *fiber.Ctx) error {
	drives, err := api.appFs.AvailableDrives()
	if err != nil {
		log.Err(err).Msg("error looking up available drives")
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	directories := make([]*fileInfo, 0)

	for _, d := range drives {
		directories = append(directories, &fileInfo{Title: d, Path: d})
	}

	// Include path classification; ancestor, course, descendant, none
	if classificationResult, err := api.courseDao.ClassifyPaths(drives); err != nil {
		log.Err(err).Msg("error classifying paths")
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	} else {
		for _, dir := range directories {
			dir.Classification = classificationResult[dir.Path]
		}
	}

	return c.Status(fiber.StatusOK).JSON(&fileSystemResponse{
		Count:       len(drives),
		Directories: directories,
		Files:       []*fileInfo{},
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// path queries the path and builds a slice of files and directories
func (api *fs) path(c *fiber.Ctx) error {
	encodedPath := c.Params("path")

	path, err := utils.DecodeString(encodedPath)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	directories := make([]*fileInfo, 0)
	files := make([]*fileInfo, 0)

	// Get a string slice of items in a directory
	items, err := api.appFs.ReadDir(path, true)
	if err != nil {
		log.Err(err).Msg("error reading directory")
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	paths := make([]string, 0)
	for _, directory := range items.Directories {
		path := filepath.Join(path, directory.Name())
		paths = append(paths, path)

		directories = append(directories, &fileInfo{Title: directory.Name(), Path: path})
	}

	// Include path classification; ancestor, course, descendant, none
	if classificationResult, err := api.courseDao.ClassifyPaths(paths); err != nil {
		log.Err(err).Msg("error classifying paths")
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	} else {
		for _, dir := range directories {
			dir.Classification = classificationResult[dir.Path]
		}
	}

	for _, file := range items.Files {
		files = append(files, &fileInfo{Title: file.Name(), Path: filepath.Join(path, file.Name())})
	}

	return c.Status(fiber.StatusOK).JSON(&fileSystemResponse{
		Count:       len(directories) + len(files),
		Directories: directories,
		Files:       files,
	})
}
