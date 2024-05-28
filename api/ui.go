package api

import (
	"net/http"

	"github.com/geerew/off-course/ui"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func bindUi(router fiber.Router, config *RouterConfig) {
	if config.IsProduction {
		// Load static assets from binary in production
		router.Use(filesystem.New(filesystem.Config{
			Root: http.FS(ui.Assets()),
		}))
	} else {
		// Load static assets from disk in development
		router.Use(filesystem.New(filesystem.Config{
			Root: http.Dir("./ui/build"),
		}))
	}
}
