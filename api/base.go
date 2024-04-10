package api

import (
	"time"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/jobs"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/rs/zerolog/log"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Router defines a router
type Router struct {
	db            database.Database
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

	// Logger middleware
	router.Use(func(c *fiber.Ctx) error {
		// Start timer
		startTime := time.Now()
		err := c.Next()
		stopTime := time.Now()

		// Log the request and response details using zerolog
		log.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("duration", stopTime.Sub(startTime)).
			Msg("Request completed")

		return err
	})

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET, POST, PUT, DELETE, HEAD, PATCH",
	}))

	// UI
	bindUi(router, config.IsProduction)

	// API
	api := router.Group("/api")
	bindFsApi(api, config.AppFs)
	bindCoursesApi(api, config.AppFs, config.Db, config.CourseScanner)
	bindAssetsApi(api, config.AppFs, config.Db)
	bindAttachmentsApi(api, config.AppFs, config.Db)
	bindScansApi(api, config.AppFs, config.Db, config.CourseScanner)
	bindTagsApi(api, config.Db)

	return router
}
