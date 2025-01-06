package api

import (
	"context"
	"log/slog"
	"net"
	"time"

	"sync/atomic"

	"github.com/Masterminds/squirrel"
	"github.com/fatih/color"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/coursescan"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Router defines a router
type Router struct {
	Router       *fiber.App
	api          fiber.Router
	config       *RouterConfig
	dao          *dao.DAO
	logDao       *dao.DAO
	bootstrapped int32
	sessionStore *session.Store
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RouterConfig defines the configuration for the router
type RouterConfig struct {
	DbManager    *database.DatabaseManager
	Logger       *slog.Logger
	AppFs        *appFs.AppFs
	CourseScan   *coursescan.CourseScan
	HttpAddr     string
	IsProduction bool
	SkipAuth     bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// New creates a new router
func NewRouter(config *RouterConfig) *Router {
	r := &Router{
		config: config,
		dao:    dao.NewDAO(config.DbManager.DataDb),
		logDao: dao.NewDAO(config.DbManager.LogsDb),
	}

	r.sessionStore = session.New(session.Config{
		KeyLookup:      "cookie:session",
		Expiration:     7 * (24 * time.Hour),
		CookieHTTPOnly: true,
	})

	r.initRouter()

	return r
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Serve serves the API and UI
func (r *Router) Serve() error {
	r.initBootstrap()

	ln, err := net.Listen("tcp", r.config.HttpAddr)
	if err != nil {
		return err
	}

	utils.Infof(
		"%s %s",
		"Server started at", color.CyanString("http://%s\n", r.config.HttpAddr),
	)

	return r.Router.Listener(ln)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Private
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initRouter initializes the router (API and UI)
func (r *Router) initRouter() {
	r.Router = fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// Middleware
	r.Router.Use(loggerMiddleware(r.config))
	r.Router.Use(corsMiddleWare())
	if !r.config.SkipAuth {
		r.Router.Use(bootstrapMiddleware(r))
		r.Router.Use(authMiddleware(r))
	}

	// UI
	r.bindUi()

	// API
	r.api = r.Router.Group("/api")
	r.initAuthRoutes()
	r.initFsRoutes()
	r.initCourseRoutes()
	r.initScanRoutes()
	r.initTagRoutes()
	r.initLogRoutes()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initBootstrap determines if the application is bootstrapped by checking if there is
// an admin user
func (r *Router) initBootstrap() {
	options := &database.Options{
		Where: squirrel.Eq{models.USER_TABLE + "." + models.USER_ROLE: types.UserRoleAdmin},
	}
	count, err := r.dao.Count(context.Background(), &models.User{}, options)
	if err != nil {
		utils.Errf("Failed to count users: %s\n", err.Error())
	}

	if count != 0 {
		atomic.StoreInt32(&r.bootstrapped, 1)
	} else {
		atomic.StoreInt32(&r.bootstrapped, 0)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// setBootstrapped sets the application as bootstrapped
func (r *Router) setBootstrapped() {
	atomic.StoreInt32(&r.bootstrapped, 1)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// isBootstrapped checks if the application is bootstrapped
func (r *Router) isBootstrapped() bool {
	return atomic.LoadInt32(&r.bootstrapped) == 1
}
