package api

import (
	"log/slog"
	"path/filepath"

	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type fsAPI struct {
	logger *slog.Logger
	appFs  *appFs.AppFs
	dao    *dao.DAO
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initFsRoutes initializes the filesystem routes
func (r *Router) initFsRoutes() {
	fsAPI := fsAPI{
		logger: r.config.Logger,
		appFs:  r.config.AppFs,
		dao:    r.dao,
	}

	fsGroup := r.api.Group("/fileSystem")

	fsGroup.Get("", fsAPI.fileSystem)
	fsGroup.Get("/:path", fsAPI.path)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// fileSystem queries the underlying system for available drives
//
// Note: On WSL, the drives will consist of / and /mnt* (ignoring /mnt/wsl*)
func (api fsAPI) fileSystem(c *fiber.Ctx) error {
	drives, err := api.appFs.AvailableDrives()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up available drives - " + err.Error(),
		})
	}

	directories := make([]*fileInfoResponse, 0)

	normalizedPaths := make([]string, len(drives))
	for _, d := range drives {
		normalizedPath := utils.NormalizeWindowsDrive(d)
		directories = append(directories, &fileInfoResponse{Title: d, Path: normalizedPath})
		normalizedPaths = append(normalizedPaths, normalizedPath)
	}

	// Include path classification; ancestor, course, descendant, none
	if classificationResult, err := api.dao.ClassifyCoursePaths(c.UserContext(), normalizedPaths); err != nil {
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
		Files:       []*fileInfoResponse{},
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// path queries the path and builds a slice of files and directories
func (api fsAPI) path(c *fiber.Ctx) error {
	encodedPath := c.Params("path")

	path, err := utils.DecodeString(encodedPath)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	directories := make([]*fileInfoResponse, 0)
	files := make([]*fileInfoResponse, 0)

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

		directories = append(directories, &fileInfoResponse{Title: directory.Name(), Path: path})
	}

	// Include path classification; ancestor, course, descendant, none
	if classificationResult, err := api.dao.ClassifyCoursePaths(c.UserContext(), paths); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error classifying paths - " + err.Error(),
		})
	} else {
		for _, dir := range directories {
			dir.Classification = classificationResult[dir.Path]
		}
	}

	for _, file := range items.Files {
		files = append(files, &fileInfoResponse{Title: file.Name(), Path: filepath.Join(path, file.Name())})
	}

	return c.Status(fiber.StatusOK).JSON(&fileSystemResponse{
		Count:       len(directories) + len(files),
		Directories: directories,
		Files:       files,
	})
}
