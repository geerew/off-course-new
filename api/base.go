package api

import (
	"log/slog"

	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/jobs"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Router defines a router
type Router struct {
	router *fiber.App
	config *RouterConfig
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RouterConfig defines the router config when creating a new
// router
type RouterConfig struct {
	DbManager     *database.DatabaseManager
	Logger        *slog.Logger
	AppFs         *appFs.AppFs
	CourseScanner *jobs.CourseScanner
	Port          string
	IsProduction  bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// New creates a new router
func New(config *RouterConfig) *Router {
	return &Router{
		router: initRouter(config),
		config: config,
	}

}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Serve serves the API and UI
func (router *Router) Serve() error {
	return router.router.Listen(router.config.Port)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Private
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initRouter initializes the router (API and UI)
func initRouter(config *RouterConfig) *fiber.App {
	router := fiber.New()

	// Middleware
	router.Use(loggerMiddleware(config))
	router.Use(corsMiddleWare())

	// UI
	bindUi(router, config)

	// API
	api := router.Group("/api")
	initRoutes(config, api)

	return router
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initRoutes initializes the routes
func initRoutes(config *RouterConfig, router fiber.Router) {
	courseDao := daos.NewCourseDao(config.DbManager.DataDb)
	courseProgressDao := daos.NewCourseProgressDao(config.DbManager.DataDb)
	assetDao := daos.NewAssetDao(config.DbManager.DataDb)
	assetProgressDao := daos.NewAssetProgressDao(config.DbManager.DataDb)
	attachmentDao := daos.NewAttachmentDao(config.DbManager.DataDb)
	tagDao := daos.NewTagDao(config.DbManager.DataDb)
	courseTagDao := daos.NewCourseTagDao(config.DbManager.DataDb)
	scanDao := daos.NewScanDao(config.DbManager.DataDb)
	logDao := daos.NewLogDao(config.DbManager.LogsDb)

	// ----------------------
	// Filesystem
	// ----------------------
	FsApi := fs{
		logger:    config.Logger,
		appFs:     config.AppFs,
		courseDao: courseDao,
	}

	filesystem := router.Group("/fileSystem")
	filesystem.Get("", FsApi.fileSystem)
	filesystem.Get("/:path", FsApi.path)

	// ----------------------
	// Courses
	// ----------------------
	coursesApi := courses{
		logger:            config.Logger,
		appFs:             config.AppFs,
		courseScanner:     config.CourseScanner,
		courseDao:         courseDao,
		courseProgressDao: courseProgressDao,
		assetDao:          assetDao,
		attachmentDao:     attachmentDao,
		tagDao:            tagDao,
		courseTagDao:      courseTagDao,
	}

	courses := router.Group("/courses")
	courses.Get("", coursesApi.getCourses)
	courses.Get("/:id", coursesApi.getCourse)
	courses.Post("", coursesApi.createCourse)
	courses.Delete("/:id", coursesApi.deleteCourse)

	courses.Head("/:id/card", coursesApi.getCard)
	courses.Get("/:id/card", coursesApi.getCard)

	courses.Get("/:id/assets", coursesApi.getAssets)
	courses.Get("/:id/assets/:asset", coursesApi.getAsset)

	courses.Get("/:id/assets/:asset/attachments", coursesApi.getAssetAttachments)
	courses.Get("/:id/assets/:asset/attachments/:attachment", coursesApi.getAssetAttachment)

	courses.Get("/:id/tags", coursesApi.getTags)
	courses.Post("/:id/tags", coursesApi.createTag)
	courses.Delete("/:id/tags/:tagId", coursesApi.deleteTag)

	// ----------------------
	// Assets
	// ----------------------
	assetsApi := assets{
		logger:           config.Logger,
		appFs:            config.AppFs,
		assetDao:         assetDao,
		assetProgressDao: assetProgressDao,
		attachmentDao:    attachmentDao,
	}

	assets := router.Group("/assets")
	assets.Get("", assetsApi.getAssets)
	assets.Get("/:id", assetsApi.getAsset)
	assets.Put("/:id", assetsApi.updateAsset)
	assets.Get("/:id/serve", assetsApi.serveAsset)

	// ----------------------
	// Attachments
	// ----------------------
	attachmentsApi := attachments{
		logger:        config.Logger,
		appFs:         config.AppFs,
		attachmentDao: attachmentDao,
	}

	attachments := router.Group("/attachments")
	attachments.Get("", attachmentsApi.getAttachments)
	attachments.Get("/:id", attachmentsApi.getAttachment)
	attachments.Get("/:id/serve", attachmentsApi.serveAttachment)

	// ----------------------
	// Tags
	// ----------------------
	tagsApi := tags{
		logger:       config.Logger,
		tagDao:       tagDao,
		courseTagDao: courseTagDao,
	}

	tags := router.Group("/tags")
	tags.Get("", tagsApi.getTags)
	tags.Get("/:id", tagsApi.getTag)
	tags.Post("", tagsApi.createTag)
	tags.Put("/:id", tagsApi.updateTag)
	tags.Delete("/:id", tagsApi.deleteTag)

	// ----------------------
	// Scans
	// ----------------------
	scansApi := scans{
		logger:        config.Logger,
		appFs:         config.AppFs,
		courseScanner: config.CourseScanner,
		scanDao:       scanDao,
	}

	scans := router.Group("/scans")
	scans.Get("/:courseId", scansApi.getScan)
	scans.Post("", scansApi.createScan)

	// ----------------------
	// Logs
	// ----------------------
	logsApi := logs{
		logger: config.Logger,
		logDao: logDao,
	}

	logs := router.Group("/logs")
	logs.Get("/", logsApi.getLogs)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// errorResponse is a helper method to return an error response
func errorResponse(c *fiber.Ctx, status int, message string, err error) error {
	resp := fiber.Map{
		"message": message,
	}

	if err != nil {
		resp["error"] = err.Error()
	}

	return c.Status(status).JSON(resp)
}
