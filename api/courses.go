package api

import (
	"database/sql"
	"os"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/jobs"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type courses struct {
	appFs         *appFs.AppFs
	courseScanner *jobs.CourseScanner
	courseDao     *daos.CourseDao
	assetDao      *daos.AssetDao
	attachmentDao *daos.AttachmentDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
type courseResponse struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Path      string         `json:"path"`
	HasCard   bool           `json:"hasCard"`
	Available bool           `json:"available"`
	CreatedAt types.DateTime `json:"createdAt"`
	UpdatedAt types.DateTime `json:"updatedAt"`

	// Scan status
	ScanStatus string `json:"scanStatus"`

	// Progress
	Started           bool           `json:"started"`
	StartedAt         types.DateTime `json:"startedAt"`
	Percent           int            `json:"percent"`
	CompletedAt       types.DateTime `json:"completedAt"`
	ProgressUpdatedAt types.DateTime `json:"progressUpdatedAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func bindCoursesApi(router fiber.Router, appFs *appFs.AppFs, db database.Database, courseScanner *jobs.CourseScanner) {
	api := courses{
		appFs:         appFs,
		courseScanner: courseScanner,
		courseDao:     daos.NewCourseDao(db),
		assetDao:      daos.NewAssetDao(db),
		attachmentDao: daos.NewAttachmentDao(db),
	}

	subGroup := router.Group("/courses")

	// Courses
	subGroup.Get("", api.getCourses)
	subGroup.Get("/:id", api.getCourse)
	subGroup.Post("", api.createCourse)
	subGroup.Delete("/:id", api.deleteCourse)

	// Card
	subGroup.Head("/:id/card", api.getCard)
	subGroup.Get("/:id/card", api.getCard)

	// Assets
	subGroup.Get("/:id/assets", api.getAssets)
	subGroup.Get("/:id/assets/:asset", api.getAsset)

	// // Attachments
	subGroup.Get("/:id/assets/:asset/attachments", api.getAssetAttachments)
	subGroup.Get("/:id/assets/:asset/attachments/:attachment", api.getAssetAttachment)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) getCourses(c *fiber.Ctx) error {
	started := c.Query("started", "undefined")
	completed := c.Query("completed", "undefined")

	dbParams := &database.DatabaseParams{
		OrderBy:    []string{c.Query("orderBy", []string{"created_at desc"}...)},
		Pagination: pagination.NewFromApi(c),
	}

	// Filter on started (if defined)
	if started != "undefined" {
		if started == "true" {
			dbParams.Where = squirrel.And{
				squirrel.Eq{daos.TableCoursesProgress() + ".started": true},
				squirrel.NotEq{daos.TableCoursesProgress() + ".percent": 100},
			}
		} else {
			dbParams.Where = squirrel.NotEq{daos.TableCoursesProgress() + ".started": true}
		}
	}

	// Filter on completed (if defined)
	if completed != "undefined" {
		if completed == "true" {
			dbParams.Where = squirrel.Eq{daos.TableCoursesProgress() + ".percent": 100}

		} else {
			dbParams.Where = squirrel.Lt{daos.TableCoursesProgress() + ".percent": 100}
		}
	}

	courses, err := api.courseDao.List(dbParams, nil)

	if err != nil {
		log.Err(err).Msg("error looking up courses")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up courses - " + err.Error(),
		})
	}

	pResult, err := dbParams.Pagination.BuildResult(courseResponseHelper(courses))
	if err != nil {
		log.Err(err).Msg("error building pagination result")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error building pagination result - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) getCourse(c *fiber.Ctx) error {
	id := c.Params("id")

	course, err := api.courseDao.Get(id, nil)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up course - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(courseResponseHelper([]*models.Course{course})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) createCourse(c *fiber.Ctx) error {
	course := new(models.Course)

	if err := c.BodyParser(course); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "error parsing data - " + err.Error(),
		})
	}

	// Ensure there is a title and path
	if course.Title == "" || course.Path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "a title and path are required",
		})
	}

	// Validate the path
	if exists, err := afero.DirExists(api.appFs.Fs, course.Path); err != nil || !exists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid course path",
		})
	}

	// Set the course to available
	course.Available = true

	if err := api.courseDao.Create(course); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "a course with this path already exists - " + err.Error(),
			})
		}

		log.Err(err).Msg("error creating course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error creating course - " + err.Error(),
		})
	}

	// Start a scan job
	if scan, err := api.courseScanner.Add(course.ID); err != nil {
		log.Err(err).Msg("error creating scan job")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error creating scan job - " + err.Error(),
		})
	} else {
		course.ScanStatus = scan.Status.String()
	}

	return c.Status(fiber.StatusCreated).JSON(courseResponseHelper([]*models.Course{course})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) deleteCourse(c *fiber.Ctx) error {
	id := c.Params("id")

	err := api.courseDao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": id}}, nil)
	if err != nil {
		log.Err(err).Msg("error deleting course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error deleting course - " + err.Error(),
		})

	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) getCard(c *fiber.Ctx) error {
	id := c.Params("id")

	course, err := api.courseDao.Get(id, nil)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Course not found")
		}

		log.Err(err).Msg("error looking up course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up course - " + err.Error(),
		})
	}

	if course.CardPath == "" {
		return c.Status(fiber.StatusNotFound).SendString("Course has no card")
	}

	_, err = api.appFs.Fs.Stat(course.CardPath)
	if os.IsNotExist(err) {
		log.Err(err).Str("card", course.CardPath).Msg("card not found on disk")
		return c.Status(fiber.StatusNotFound).SendString("Course card not found")
	}

	// The fiber function sendFile(...) does not support using a custom FS. Therefore, use
	// SendFile() from the filesystem middleware.
	return filesystem.SendFile(c, afero.NewHttpFs(api.appFs.Fs), course.CardPath)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) getAssets(c *fiber.Ctx) error {
	id := c.Params("id")

	// Get the course
	_, err := api.courseDao.Get(id, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up course - " + err.Error(),
		})
	}

	dbParams := &database.DatabaseParams{
		OrderBy:    []string{c.Query("orderBy", []string{"chapter asc", "prefix asc"}...)},
		Where:      squirrel.Eq{daos.TableAssets() + ".course_id": id},
		Pagination: pagination.NewFromApi(c),
	}

	assets, err := api.assetDao.List(dbParams, nil)
	if err != nil {
		log.Err(err).Msg("error looking up assets")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up assets - " + err.Error(),
		})
	}

	pResult, err := dbParams.Pagination.BuildResult(assetResponseHelper(assets))
	if err != nil {
		log.Err(err).Msg("error building pagination result")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error building pagination result - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) getAsset(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")

	_, err := api.courseDao.Get(id, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up course - " + err.Error(),
		})
	}

	asset, err := api.assetDao.Get(assetId, nil, nil)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up asset")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up asset - " + err.Error(),
		})
	}

	if asset.CourseID != id {
		return c.Status(fiber.StatusBadRequest).SendString("Asset does not belong to course")
	}

	return c.Status(fiber.StatusOK).JSON(assetResponseHelper([]*models.Asset{asset})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) getAssetAttachments(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")

	// Get the course
	_, err := api.courseDao.Get(id, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up course - " + err.Error(),
		})
	}

	// Get the asset
	asset, err := api.assetDao.Get(assetId, nil, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up asset")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up asset - " + err.Error(),
		})
	}

	if asset.CourseID != id {
		return c.Status(fiber.StatusBadRequest).SendString("Asset does not belong to course")
	}

	dbParams := &database.DatabaseParams{
		OrderBy:    []string{c.Query("orderBy", []string{"title asc"}...)},
		Where:      squirrel.Eq{daos.TableAttachments() + ".asset_id": assetId},
		Pagination: pagination.NewFromApi(c),
	}

	attachments, err := api.attachmentDao.List(dbParams, nil)
	if err != nil {
		log.Err(err).Msg("error looking up attachments")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up attachments - " + err.Error(),
		})
	}

	pResult, err := dbParams.Pagination.BuildResult(attachmentResponseHelper(attachments))
	if err != nil {
		log.Err(err).Msg("error building pagination result")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error building pagination result - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) getAssetAttachment(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")
	attachmentId := c.Params("attachment")

	// Get the course
	_, err := api.courseDao.Get(id, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up course - " + err.Error(),
		})
	}

	// Get the asset
	asset, err := api.assetDao.Get(assetId, nil, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up asset")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up asset - " + err.Error(),
		})
	}

	if asset.CourseID != id {
		return c.Status(fiber.StatusBadRequest).SendString("Asset does not belong to course")
	}

	// Get the attachment
	attachment, err := api.attachmentDao.Get(attachmentId, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up attachment")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up attachment - " + err.Error(),
		})
	}

	if attachment.AssetID != assetId {
		return c.Status(fiber.StatusBadRequest).SendString("Attachment does not belong to asset")
	}

	return c.Status(fiber.StatusOK).JSON(attachmentResponseHelper([]*models.Attachment{attachment})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPER
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func courseResponseHelper(courses []*models.Course) []*courseResponse {
	responses := []*courseResponse{}
	for _, course := range courses {
		responses = append(responses, &courseResponse{
			ID:        course.ID,
			Title:     course.Title,
			Path:      course.Path,
			HasCard:   course.CardPath != "",
			Available: course.Available,
			CreatedAt: course.CreatedAt,
			UpdatedAt: course.UpdatedAt,

			// Scan status
			ScanStatus: course.ScanStatus,

			// Progress
			Started:           course.Started,
			StartedAt:         course.StartedAt,
			Percent:           course.Percent,
			CompletedAt:       course.CompletedAt,
			ProgressUpdatedAt: course.ProgressUpdatedAt,
		})
	}

	return responses
}
