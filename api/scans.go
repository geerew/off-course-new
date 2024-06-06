package api

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/jobs"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type scans struct {
	logger        *slog.Logger
	appFs         *appFs.AppFs
	scanDao       *daos.ScanDao
	courseScanner *jobs.CourseScanner
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type scanResponse struct {
	ID        string           `json:"id"`
	CourseID  string           `json:"courseId"`
	Status    types.ScanStatus `json:"status"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *scans) getScan(c *fiber.Ctx) error {
	courseId := c.Params("courseId")

	scan, err := api.scanDao.Get(courseId, nil)

	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Scan not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up scan", err)
	}

	return c.Status(fiber.StatusOK).JSON(scanResponseHelper([]*models.Scan{scan})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *scans) createScan(c *fiber.Ctx) error {
	scan := &models.Scan{}

	if err := c.BodyParser(scan); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	if scan.CourseID == "" {
		return errorResponse(c, fiber.StatusBadRequest, "A course ID is required", nil)
	}

	scan, err := api.courseScanner.Add(scan.CourseID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid course ID", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error creating scan job", err)
	}

	return c.Status(fiber.StatusCreated).JSON(scanResponseHelper([]*models.Scan{scan})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func scanResponseHelper(scans []*models.Scan) []*scanResponse {
	responses := []*scanResponse{}
	for _, scan := range scans {
		responses = append(responses, &scanResponse{
			ID:        scan.ID,
			CourseID:  scan.CourseID,
			Status:    scan.Status,
			CreatedAt: scan.CreatedAt,
			UpdatedAt: scan.UpdatedAt,
		})
	}

	return responses
}
