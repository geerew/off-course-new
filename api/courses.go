package api

import (
	"database/sql"
	"os"
	"strings"

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
	db            database.Database
	appFs         *appFs.AppFs
	courseScanner *jobs.CourseScanner
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
type courseResponse struct {
	ID         string         `json:"id"`
	Title      string         `json:"title"`
	Path       string         `json:"path"`
	HasCard    bool           `json:"hasCard"`
	Started    bool           `json:"started"`
	Finished   bool           `json:"finished"`
	ScanStatus string         `json:"scanStatus"`
	CreatedAt  types.DateTime `json:"createdAt"`
	UpdatedAt  types.DateTime `json:"updatedAt"`

	// Association
	Assets []*assetResponse `json:"assets,omitempty"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func bindCoursesApi(router fiber.Router, appFs *appFs.AppFs, db database.Database, courseScanner *jobs.CourseScanner) {
	api := courses{appFs: appFs, db: db, courseScanner: courseScanner}

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

	// Attachments
	subGroup.Get("/:id/assets/:asset/attachments", api.getAssetAttachments)
	subGroup.Get("/:id/assets/:asset/attachments/:attachment", api.getAssetAttachment)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) getCourses(c *fiber.Ctx) error {
	orderBy := c.Query("orderBy", "created_at desc")

	dbParams := &database.DatabaseParams{
		OrderBy:    []string{orderBy},
		Pagination: pagination.New(c),
	}

	// Include relations
	if c.QueryBool("expand", false) {
		dbParams.Relation = []database.Relation{
			{Struct: "Assets", OrderBy: []string{"chapter asc", "prefix asc"}},
			{Struct: "Assets.Attachments", OrderBy: []string{"title asc"}}}
	}

	courses, err := models.GetCourses(c.UserContext(), api.db, dbParams)
	if err != nil {
		log.Err(err).Msg("error looking up courses")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up courses - " + err.Error(),
		})
	}

	pResult, err := dbParams.Pagination.BuildResult(toCourseResponse(courses))
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

	dbParams := &database.DatabaseParams{}

	// Include relations
	if c.QueryBool("expand", false) {
		dbParams.Relation = []database.Relation{
			{Struct: "Assets", OrderBy: []string{"chapter asc", "prefix asc"}},
			{Struct: "Assets.Attachments", OrderBy: []string{"title asc"}}}
	}

	course, err := models.GetCourseById(c.UserContext(), api.db, dbParams, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up course - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(toCourseResponse([]*models.Course{course})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) createCourse(c *fiber.Ctx) error {
	course := new(models.Course)

	if err := c.BodyParser(course); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "error parsing data - " + err.Error(),
		})
	}

	// Ensure we have a title and path
	if course.Title == "" || course.Path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "a title and path are required",
		})
	}

	// Check for invalid path
	if exists, err := afero.DirExists(api.appFs.Fs, course.Path); err != nil || !exists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid course path",
		})
	}

	if err := models.CreateCourse(c.UserContext(), api.db, course); err != nil {
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

	return c.Status(fiber.StatusCreated).JSON(toCourseResponse([]*models.Course{course})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) deleteCourse(c *fiber.Ctx) error {
	id := c.Params("id")

	count, err := models.DeleteCourse(c.UserContext(), api.db, id)
	if err != nil {
		log.Err(err).Msg("error deleting course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error deleting course - " + err.Error(),
		})

	}

	if count == 0 {
		return c.Status(fiber.StatusNotFound).SendString("Not found")
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) getCard(c *fiber.Ctx) error {
	id := c.Params("id")

	course, err := models.GetCourseById(c.UserContext(), api.db, nil, id)

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
	_, err := models.GetCourseById(c.UserContext(), api.db, nil, id)
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
		Pagination: pagination.New(c),
	}

	// Include relations
	if c.QueryBool("expand", false) {
		dbParams.Relation = []database.Relation{
			{Struct: "Attachments", OrderBy: []string{"title asc"}}}
	}

	assets, err := models.GetAssetsByCourseId(c.UserContext(), api.db, dbParams, id)
	if err != nil {
		log.Err(err).Msg("error looking up assets")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up assets - " + err.Error(),
		})
	}

	pResult, err := dbParams.Pagination.BuildResult(toAssetResponse(assets))
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

	// Get the course
	_, err := models.GetCourseById(c.UserContext(), api.db, nil, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up course - " + err.Error(),
		})
	}

	// Include relations
	dbParams := &database.DatabaseParams{}
	if c.QueryBool("expand", false) {
		dbParams.Relation = []database.Relation{
			{Struct: "Attachments", OrderBy: []string{"title asc"}}}
	}

	asset, err := models.GetAssetById(c.UserContext(), api.db, dbParams, assetId)

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

	return c.Status(fiber.StatusOK).JSON(toAssetResponse([]*models.Asset{asset})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) getAssetAttachments(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")

	// Get the course
	_, err := models.GetCourseById(c.UserContext(), api.db, nil, id)
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
	asset, err := models.GetAssetById(c.UserContext(), api.db, nil, assetId)
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
		Pagination: pagination.New(c),
	}

	attachments, err := models.GetAttachmentsByAssetId(c.UserContext(), api.db, dbParams, assetId)
	if err != nil {
		log.Err(err).Msg("error looking up attachments")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up attachments - " + err.Error(),
		})
	}

	pResult, err := dbParams.Pagination.BuildResult(toAttachmentResponse(attachments))
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
	_, err := models.GetCourseById(c.UserContext(), api.db, nil, id)
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
	asset, err := models.GetAssetById(c.UserContext(), api.db, nil, assetId)
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
	attachment, err := models.GetAttachmentById(c.UserContext(), api.db, nil, attachmentId)
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

	return c.Status(fiber.StatusOK).JSON(toAttachmentResponse([]*models.Attachment{attachment})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPER
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func toCourseResponse(courses []*models.Course) []*courseResponse {
	responses := []*courseResponse{}
	for _, course := range courses {
		responses = append(responses, &courseResponse{
			ID:         course.ID,
			Title:      course.Title,
			Path:       course.Path,
			HasCard:    course.CardPath != "",
			Started:    course.Started,
			Finished:   course.Finished,
			ScanStatus: course.ScanStatus,
			CreatedAt:  course.CreatedAt,
			UpdatedAt:  course.UpdatedAt,

			// Association
			Assets: toAssetResponse(course.Assets),
		})
	}

	return responses
}
