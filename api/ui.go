package api

import (
	"net/http"

	"github.com/geerew/off-course/ui"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (r *Router) bindUi() {
	if r.config.IsProduction {
		// Load static assets from binary in production
		r.router.Use(filesystem.New(filesystem.Config{
			Root: http.FS(ui.Assets()),
		}))
	} else {
		// Load static assets from disk in development
		r.router.Use(filesystem.New(filesystem.Config{
			Root: http.Dir("./ui/build"),
		}))
	}
}
