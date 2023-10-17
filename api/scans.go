package api

import (
	"database/sql"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/jobs"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type scans struct {
	appFs         *appFs.AppFs
	db            database.Database
	courseScanner *jobs.CourseScanner
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type scanResponse struct {
	ID        string           `json:"id"`
	CourseID  string           `json:"courseId"`
	Status    types.ScanStatus `json:"status"`
	CreatedAt types.DateTime   `json:"createdAt"`
	UpdatedAt types.DateTime   `json:"updatedAt"`

	// Association
	Course *models.Course `json:"course,omitempty"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func bindScansApi(router fiber.Router, appFs *appFs.AppFs, db database.Database, courseScanner *jobs.CourseScanner) {
	api := scans{appFs: appFs, db: db, courseScanner: courseScanner}

	subGroup := router.Group("/scans")

	subGroup.Get("", api.getScans)
	subGroup.Get("/:id", api.getScan)
	subGroup.Get("/course/:id", api.getScanByCourseId)
	subGroup.Post("", api.createScan)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *scans) getScans(c *fiber.Ctx) error {
	orderBy := c.Query("orderBy", "created_at desc")

	dbParams := &database.DatabaseParams{
		OrderBy:    []string{orderBy},
		Pagination: pagination.New(c),
	}

	// Include relations
	if c.QueryBool("expand", false) {
		dbParams.Relation = []database.Relation{{Struct: "Course"}}
	}

	scans, err := models.GetScans(c.UserContext(), api.db, dbParams)
	if err != nil {
		log.Err(err).Msg("error looking up scans")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up scans - " + err.Error(),
		})
	}

	pResult, err := dbParams.Pagination.BuildResult(toScanResponse(scans))
	if err != nil {
		log.Err(err).Msg("error building pagination result")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error building pagination result - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *scans) getScan(c *fiber.Ctx) error {
	id := c.Params("id")

	dbParams := &database.DatabaseParams{}

	// Include relations
	if c.QueryBool("expand", false) {
		dbParams.Relation = []database.Relation{{Struct: "Course"}}
	}

	scan, err := models.GetScanById(c.UserContext(), api.db, dbParams, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up scan")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up scan - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(toScanResponse([]*models.Scan{scan})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *scans) getScanByCourseId(c *fiber.Ctx) error {
	id := c.Params("id")

	dbParams := &database.DatabaseParams{}

	// Include relations
	if c.QueryBool("expand", false) {
		dbParams.Relation = []database.Relation{{Struct: "Course"}}
	}

	scan, err := models.GetScanByCourseId(c.UserContext(), api.db, dbParams, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up scan")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up scan - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(toScanResponse([]*models.Scan{scan})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *scans) createScan(c *fiber.Ctx) error {
	scan := &models.Scan{}

	if err := c.BodyParser(scan); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "error parsing data - " + err.Error(),
		})
	}

	if scan.CourseID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "a course ID is required",
		})
	}

	// Start a scan job
	scan, err := api.courseScanner.Add(scan.CourseID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "invalid course ID",
			})
		}

		log.Err(err).Msg("error creating scan job")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error creating scan job - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(toScanResponse([]*models.Scan{scan})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPER
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func toScanResponse(scans []*models.Scan) []*scanResponse {
	responses := []*scanResponse{}
	for _, scan := range scans {
		responses = append(responses, &scanResponse{
			ID:        scan.ID,
			CourseID:  scan.CourseID,
			Status:    scan.Status,
			Course:    scan.Course,
			CreatedAt: scan.CreatedAt,
			UpdatedAt: scan.UpdatedAt,
		})
	}

	return responses
}
