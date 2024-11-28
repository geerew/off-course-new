package api

import (
	"log/slog"
	"net/url"
	"strings"

	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/coursescan"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Router defines a router
type Router struct {
	router *fiber.App
	api    fiber.Router
	config *RouterConfig
	dao    *dao.DAO
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RouterConfig defines the configuration for the router
type RouterConfig struct {
	DbManager    *database.DatabaseManager
	Logger       *slog.Logger
	AppFs        *appFs.AppFs
	CourseScan   *coursescan.CourseScan
	Port         string
	IsProduction bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// New creates a new router
func NewRouter(config *RouterConfig) *Router {
	r := &Router{
		config: config,
		dao:    dao.NewDAO(config.DbManager.DataDb),
	}

	r.initRouter()

	return r

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
func (r *Router) initRouter() {
	r.router = fiber.New()

	// Middleware
	r.router.Use(loggerMiddleware(r.config))
	r.router.Use(corsMiddleWare())

	// UI
	r.bindUi()

	// API
	r.api = r.router.Group("/api")
	r.initFsRoutes()
	r.initCoursesRoutes()
}

// // ----------------------
// // Tags
// // ----------------------
// tagsApi := tags{
// 	logger:       config.Logger,
// 	tagDao:       tagDao,
// 	courseTagDao: courseTagDao,
// }

// tags := router.Group("/tags")
// tags.Get("", tagsApi.getTags)
// tags.Get("/:id", tagsApi.getTag)
// tags.Post("", tagsApi.createTag)
// tags.Put("/:id", tagsApi.updateTag)
// tags.Delete("/:id", tagsApi.deleteTag)

// // ----------------------
// // Scans
// // ----------------------
// scansApi := scans{
// 	logger:        config.Logger,
// 	appFs:         config.AppFs,
// 	courseScan: config.Scanner,
// 	scanDao:       scanDao,
// }

// scans := router.Group("/scans")
// scans.Get("/:courseId", scansApi.getScan)
// scans.Post("", scansApi.createScan)

// // ----------------------
// // Logs
// // ----------------------
// logsApi := logs{
// 	logger: config.Logger,
// 	logDao: logDao,
// }

// logs := router.Group("/logs")
// logs.Get("/", logsApi.getLogs)
// logs.Get("/types", logsApi.getLogTypes)

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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func filter(s string) ([]string, error) {
	unescaped, err := url.QueryUnescape(s)
	if err != nil {
		return nil, err
	}

	return utils.Map(strings.Split(unescaped, ","), strings.TrimSpace), nil
}
