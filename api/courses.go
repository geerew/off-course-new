package api

// import (
// 	"os"
// 	"strings"
// 	"time"

// 	"github.com/geerew/off-course/database"
// 	"github.com/geerew/off-course/models"
// 	"github.com/geerew/off-course/utils/appFs"
// 	"github.com/geerew/off-course/utils/jobs"
// 	"github.com/geerew/off-course/utils/pagination"
// 	"github.com/gofiber/fiber/v2"
// 	"github.com/gofiber/fiber/v2/middleware/filesystem"
// 	"github.com/rs/zerolog/log"
// 	"github.com/spf13/afero"
// 	"gorm.io/gorm"
// )

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// type courses struct {
// 	db            database.Database
// 	appFs         *appFs.AppFs
// 	courseScanner *jobs.CourseScanner
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// type courseResponse struct {
// 	ID         string    `json:"id"`
// 	Title      string    `json:"title"`
// 	Path       string    `json:"path"`
// 	HasCard    bool      `json:"hasCard"`
// 	Completed  bool      `json:"completed"`
// 	ScanStatus string    `json:"scanStatus"`
// 	CreatedAt  time.Time `json:"createdAt"`
// 	UpdatedAt  time.Time `json:"updatedAt"`

// 	// Association
// 	Assets []*assetResponse `json:"assets,omitempty"`
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func bindCoursesApi(router fiber.Router, appFs *appFs.AppFs, db database.Database, courseScanner *jobs.CourseScanner) {
// 	api := courses{appFs: appFs, db: db, courseScanner: courseScanner}

// 	subGroup := router.Group("/courses")

// 	// Courses
// 	subGroup.Get("", api.getCourses)
// 	subGroup.Get("/:id", api.getCourse)
// 	subGroup.Post("", api.createCourse)
// 	subGroup.Delete("/:id", api.deleteCourse)

// 	// Card
// 	subGroup.Head("/:id/card", api.getCard)
// 	subGroup.Get("/:id/card", api.getCard)

// 	// Assets
// 	subGroup.Get("/:id/assets", api.getAssets)

// 	// Attachments
// 	subGroup.Get("/:id/assets/:asset/attachments", api.getAttachments)
// 	// subGroup.Get("/:id/assets/:asset", api.getAsset)
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func (api *courses) getCourses(c *fiber.Ctx) error {
// 	orderBy := c.Query("orderBy", "created_at desc")
// 	includeAssets := c.QueryBool("includeAssets", false)

// 	dbParams := &database.DatabaseParams{
// 		OrderBy:    orderBy,
// 		Preload:    includeAssets,
// 		Pagination: pagination.New(c),
// 	}

// 	// Always order assets by chapter, then prefix
// 	if includeAssets {
// 		dbParams.PreloadOrderBy = "chapter asc, prefix asc"
// 	}

// 	courses, err := models.GetCourses(api.db, dbParams)
// 	if err != nil {
// 		log.Err(err).Msg("error looking up courses")
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "error looking up courses - " + err.Error(),
// 		})
// 	}

// 	pResult, err := dbParams.Pagination.BuildResult(toCourseResponse(courses))
// 	if err != nil {
// 		log.Err(err).Msg("error building pagination result")
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "error building pagination result - " + err.Error(),
// 		})
// 	}

// 	return c.Status(fiber.StatusOK).JSON(pResult)
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func (api *courses) getCourse(c *fiber.Ctx) error {
// 	id := c.Params("id")
// 	includeAssets := c.QueryBool("includeAssets", false)

// 	dbParams := &database.DatabaseParams{Preload: includeAssets}
// 	if includeAssets {
// 		dbParams.PreloadOrderBy = "chapter asc, prefix asc"
// 	}

// 	course, err := models.GetCourse(api.db, id, dbParams)

// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return c.Status(fiber.StatusNotFound).SendString("Not found")
// 		}

// 		log.Err(err).Msg("error looking up course")
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "error looking up course - " + err.Error(),
// 		})
// 	}

// 	return c.Status(fiber.StatusOK).JSON(toCourseResponse([]*models.Course{course})[0])
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func (api *courses) createCourse(c *fiber.Ctx) error {
// 	course := new(models.Course)

// 	if err := c.BodyParser(course); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"message": "error parsing data - " + err.Error(),
// 		})
// 	}

// 	// Ensure we have a title and path
// 	if course.Title == "" || course.Path == "" {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"message": "a title and path are required",
// 		})
// 	}

// 	// Check for invalid path
// 	if exists, err := afero.DirExists(api.appFs.Fs, course.Path); err != nil || !exists {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"message": "invalid course path",
// 		})
// 	}

// 	if err := models.CreateCourse(api.db, course); err != nil {
// 		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
// 			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 				"message": "a course with this path already exists - " + err.Error(),
// 			})
// 		}

// 		log.Err(err).Msg("error creating course")
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "error creating course - " + err.Error(),
// 		})
// 	}

// 	// Start a scan job
// 	if scan, err := api.courseScanner.Add(course.ID); err != nil {
// 		log.Err(err).Msg("error creating scan job")
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "error creating scan job - " + err.Error(),
// 		})
// 	} else {
// 		course.ScanStatus = scan.Status.String()
// 	}

// 	return c.Status(fiber.StatusCreated).JSON(toCourseResponse([]*models.Course{course})[0])
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func (api *courses) deleteCourse(c *fiber.Ctx) error {
// 	id := c.Params("id")

// 	err := models.DeleteCourse(api.db, id)

// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return c.Status(fiber.StatusNotFound).SendString("Not found")
// 		}

// 		log.Err(err).Msg("error deleting course")
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "error deleting course - " + err.Error(),
// 		})

// 	}

// 	return c.Status(fiber.StatusNoContent).Send(nil)
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func (api *courses) getCard(c *fiber.Ctx) error {
// 	id := c.Params("id")

// 	course, err := models.GetCourse(api.db, id, nil)

// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return c.Status(fiber.StatusNotFound).SendString("Course not found")
// 		}

// 		log.Err(err).Msg("error looking up course")
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "error looking up course - " + err.Error(),
// 		})
// 	}

// 	if course.CardPath == "" {
// 		return c.Status(fiber.StatusNotFound).SendString("Course has no card")
// 	}

// 	_, err = api.appFs.Fs.Stat(course.CardPath)
// 	if os.IsNotExist(err) {
// 		log.Err(err).Str("card", course.CardPath).Msg("card not found on disk")
// 		return c.Status(fiber.StatusNotFound).SendString("Course card not found")
// 	}

// 	// The fiber function sendFile(...) does not support using a custom FS. Therefore, use
// 	// SendFile() from the filesystem middleware.
// 	return filesystem.SendFile(c, afero.NewHttpFs(api.appFs.Fs), course.CardPath)
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func (api *courses) getAssets(c *fiber.Ctx) error {
// 	id := c.Params("id")

// 	// Validate the course exists
// 	course, err := models.GetCourse(api.db, id, nil)

// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return c.Status(fiber.StatusNotFound).SendString("Not found")
// 		}

// 		log.Err(err).Msg("error looking up course")
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "error looking up course - " + err.Error(),
// 		})
// 	}

// 	// Get the order by param
// 	orderBy := c.Query("orderBy", "created_at desc")

// 	dbParams := &database.DatabaseParams{
// 		OrderBy:    orderBy,
// 		Where:      database.Where{"course_id": course.ID},
// 		Preload:    true,
// 		Pagination: pagination.New(c),
// 	}

// 	assets, err := models.GetAssets(api.db, dbParams)
// 	if err != nil {
// 		log.Err(err).Msg("error looking up assets")
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "error looking up assets - " + err.Error(),
// 		})
// 	}

// 	pResult, err := dbParams.Pagination.BuildResult(toAssetResponse(assets))
// 	if err != nil {
// 		log.Err(err).Msg("error building pagination result")
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "error building pagination result - " + err.Error(),
// 		})
// 	}

// 	return c.Status(fiber.StatusOK).JSON(pResult)
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func (api *courses) getAttachments(c *fiber.Ctx) error {
// 	id := c.Params("id")
// 	assetId := c.Params("asset")

// 	// Validate the course exists
// 	course, err := models.GetCourse(api.db, id, nil)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return c.Status(fiber.StatusNotFound).SendString("Not found")
// 		}

// 		log.Err(err).Msg("error looking up course")
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "error looking up course - " + err.Error(),
// 		})
// 	}

// 	// Validate the asset exists
// 	asset, err := models.GetAsset(api.db, assetId, nil)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return c.Status(fiber.StatusNotFound).SendString("Not found")
// 		}

// 		log.Err(err).Msg("error looking up asset")
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "error looking up asset - " + err.Error(),
// 		})
// 	}

// 	if asset.CourseID != course.ID {
// 		return c.Status(fiber.StatusBadRequest).SendString("Asset does not belong to course")
// 	}

// 	attachments, err := models.GetAttachments(api.db, &database.DatabaseParams{Where: database.Where{"asset_id": asset.ID}})
// 	if err != nil {
// 		log.Err(err).Msg("error looking up attachments")
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "error looking up attachments - " + err.Error(),
// 		})
// 	}

// 	return c.Status(fiber.StatusOK).JSON(toAttachmentResponse(attachments))
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// // HELPER
// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func toCourseResponse(courses []*models.Course) []*courseResponse {
// 	responses := []*courseResponse{}
// 	for _, course := range courses {
// 		responses = append(responses, &courseResponse{
// 			ID:         course.ID,
// 			Title:      course.Title,
// 			Path:       course.Path,
// 			HasCard:    course.CardPath != "",
// 			Completed:  course.Completed,
// 			ScanStatus: course.ScanStatus,
// 			CreatedAt:  course.CreatedAt,
// 			UpdatedAt:  course.UpdatedAt,

// 			// Association
// 			Assets: toAssetResponse(course.Assets),
// 		})
// 	}

// 	return responses
// }
