package api

import (
	"net/http"
	"strings"

	"github.com/geerew/off-course/ui"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (r *Router) bindUi() {
	if r.config.IsProduction {
		r.Router.Use(filesystem.New(filesystem.Config{
			Root: http.FS(ui.Assets()),
		}))
	} else {
		r.Router.Use(func(c *fiber.Ctx) error {
			if strings.HasPrefix(c.OriginalURL(), "/api") {
				return c.Next()
			}

			uri := "http://localhost:5173" + c.OriginalURL()

			var err error
			err = proxy.Do(c, uri)
			if err != nil {
				// Sometimes svelte closes the  connection before returning the first response
				// byte. This just attempts the proxy again
				err = proxy.Do(c, uri)
			}
			return err
		})
	}
}
