package api

import (
	"path/filepath"

	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type fs struct {
	appFs *appFs.AppFs
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func bindFsApi(router fiber.Router, appFs *appFs.AppFs) {
	api := fs{appFs: appFs}

	subGroup := router.Group("/fileSystem")

	subGroup.Get("", api.fileSystem)
	subGroup.Get("/:path", api.path)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type fileSystemResponse struct {
	Count       int        `json:"count"`
	Directories []fileInfo `json:"directories"`
	Files       []fileInfo `json:"files"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type fileInfo struct {
	Title string `json:"title"`
	Path  string `json:"path"`
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

	directories := make([]fileInfo, 0)

	for _, d := range drives {
		directories = append(directories, fileInfo{Title: d, Path: d})
	}

	return c.Status(fiber.StatusOK).JSON(&fileSystemResponse{
		Count:       len(drives),
		Directories: directories,
		Files:       []fileInfo{},
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

	directories := make([]fileInfo, 0)
	files := make([]fileInfo, 0)

	// Get a string slice of items in a directory
	items, err := api.appFs.ReadDir(path, true)
	if err != nil {
		log.Err(err).Msg("error reading directory")
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	for _, directory := range items.Directories {
		directories = append(directories, fileInfo{Title: directory.Name(), Path: filepath.Join(path, directory.Name())})
	}

	for _, file := range items.Files {
		files = append(files, fileInfo{Title: file.Name(), Path: filepath.Join(path, file.Name())})
	}

	return c.Status(fiber.StatusOK).JSON(&fileSystemResponse{
		Count:       len(directories) + len(files),
		Directories: directories,
		Files:       files,
	})
}
