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
	db            database.Database
	logger        *slog.Logger
	appFs         *appFs.AppFs
	courseScanner *jobs.CourseScanner
	router        *fiber.App
	port          string
	isProduction  bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RouterConfig defines the router config when creating a new
// router
type RouterConfig struct {
	Db            database.Database
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
		db:            config.Db,
		logger:        config.Logger,
		appFs:         config.AppFs,
		courseScanner: config.CourseScanner,
		router:        initRouter(config),
		port:          config.Port,
		isProduction:  config.IsProduction,
	}

}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Serve serves the API and UI
func (router *Router) Serve() error {
	return router.router.Listen(router.port)
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
	courseDao := daos.NewCourseDao(config.Db)
	courseProgressDao := daos.NewCourseProgressDao(config.Db)
	assetDao := daos.NewAssetDao(config.Db)
	assetProgressDao := daos.NewAssetProgressDao(config.Db)
	attachmentDao := daos.NewAttachmentDao(config.Db)
	tagDao := daos.NewTagDao(config.Db)
	courseTagDao := daos.NewCourseTagDao(config.Db)
	scanDao := daos.NewScanDao(config.Db)

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
		tagDao:       daos.NewTagDao(config.Db),
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
