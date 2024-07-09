package api

import (
	"log/slog"
	"path/filepath"

	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type fs struct {
	logger    *slog.Logger
	appFs     *appFs.AppFs
	courseDao *daos.CourseDao
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up available drives - " + err.Error(),
		})
	}

	directories := make([]*fileInfo, 0)

	normalizedPaths := make([]string, len(drives))
	for _, d := range drives {
		normalizedPath := utils.NormalizeWindowsDrive(d)
		directories = append(directories, &fileInfo{Title: d, Path: normalizedPath})
		normalizedPaths = append(normalizedPaths, normalizedPath)
	}

	// Include path classification; ancestor, course, descendant, none
	if classificationResult, err := api.courseDao.ClassifyPaths(normalizedPaths); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error classifying paths - " + err.Error(),
		})
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
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "error reading directory - " + err.Error(),
		})
	}

	paths := make([]string, 0)
	for _, directory := range items.Directories {
		path := utils.NormalizeWindowsDrive(filepath.Join(path, directory.Name()))
		paths = append(paths, path)

		directories = append(directories, &fileInfo{Title: directory.Name(), Path: path})
	}

	// Include path classification; ancestor, course, descendant, none
	if classificationResult, err := api.courseDao.ClassifyPaths(paths); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error classifying paths - " + err.Error(),
		})
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
