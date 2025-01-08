package api

import (
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// loggerMiddleware logs the request and response details
func loggerMiddleware(config *RouterConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		started := time.Now()
		err := c.Next()

		attrs := make([]any, 0, 15)

		attrs = append(attrs, slog.Any("type", types.LogTypeRequest))

		if !started.IsZero() {
			attrs = append(attrs, slog.Float64("execTime", float64(time.Since(started))/float64(time.Millisecond)))
		}

		if err != nil {
			attrs = append(
				attrs,
				slog.String("error", err.Error()),
			)
		}

		method := strings.ToUpper(c.Method())

		attrs = append(
			attrs,
			slog.Int("status", c.Response().StatusCode()),
			slog.String("method", method),
		)

		// Get the port
		host := string(c.Request().Host())
		if strings.Contains(host, ":") {
			attrs = append(attrs, slog.String("port", strings.Split(host, ":")[1]))
		}

		// Determine if the response is an error
		isErrorResponse := err != nil

		var jsonBody map[string]interface{}
		if jsonErr := json.Unmarshal(c.Response().Body(), &jsonBody); jsonErr == nil {
			if val, exists := jsonBody["message"]; exists {
				attrs = append(attrs, slog.String("message", val.(string)))
			}

			if !isErrorResponse {
				if val, exists := jsonBody["error"]; exists {
					isErrorResponse = true
					attrs = append(
						attrs, slog.String("error", val.(string)))
				}
			}
		}

		message := method + " " + c.OriginalURL()

		if isErrorResponse {
			config.Logger.Error(message, attrs...)
		} else {
			config.Logger.Info(message, attrs...)
		}

		return err
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// corsMiddleWare creates a CORS middleware
func corsMiddleWare() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET, POST, PUT, DELETE, HEAD, PATCH",
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// bootstrapMiddleware checks if the app is bootstrapped. If not, it redirects
// to /auth/bootstrap
func bootstrapMiddleware(r *Router) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// If not bootstrapped, for everything through /auth/bootstrap or
		// /api/auth/bootstrap
		if !r.isBootstrapped() {
			if r.isDevUIPath(c) || r.isFavicon(c) {
				return c.Next()
			}

			// API check
			if strings.HasPrefix(c.OriginalURL(), "/api/") {
				if strings.HasPrefix(c.OriginalURL(), "/api/auth/bootstrap") {
					c.Locals("bootstrapping", true)
					return c.Next()
				} else {
					return c.SendStatus(fiber.StatusForbidden)
				}
			}

			// UI check
			if strings.HasPrefix(c.OriginalURL(), "/auth/bootstrap") {
				c.Locals("bootstrapping", true)
				return c.Next()
			} else {
				return c.Redirect("/auth/bootstrap/")
			}
		}

		return c.Next()
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// authMiddleware authenticates the request
func authMiddleware(r *Router) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// When bootstrapping, ignore auth entirely. All invalid requests will be
		// handled by the bootstrap middleware
		if bootstrapping, ok := c.Locals("bootstrapping").(bool); ok && bootstrapping {
			return c.Next()
		}

		if r.isDevUIPath(c) || r.isFavicon(c) {
			return c.Next()
		}

		// API - Always allow logout
		if strings.HasPrefix(c.OriginalURL(), "/api/auth/logout") {
			return c.Next()
		}

		// Get the session
		session, err := r.sessionStore.Get(c)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		if session.Fresh() {
			// API - Only allow login and register
			if strings.HasPrefix(c.OriginalURL(), "/api") {
				if strings.HasPrefix(c.OriginalURL(), "/api/auth/login") ||
					strings.HasPrefix(c.OriginalURL(), "/api/auth/register") {
					return c.Next()
				} else {
					return c.SendStatus(fiber.StatusForbidden)
				}
			}

			// UI - Only allow login and register
			if strings.HasPrefix(c.OriginalURL(), "/auth/login") || strings.HasPrefix(c.OriginalURL(), "/auth/register") {
				return c.Next()
			}

			return c.Redirect("/auth/login/")
		}

		// UI - Redirect auth requests to /
		if strings.HasPrefix(c.OriginalURL(), "/auth/") {
			return c.Redirect("/")
		}

		// API - Return 200 for all auth requests except /me
		if strings.HasPrefix(c.OriginalURL(), "/api/auth/") && !strings.HasPrefix(c.OriginalURL(), "/api/auth/me") {
			return c.SendStatus(fiber.StatusOK)
		}

		// Get the user ID and role from the session and set for downstream handlers
		userId := session.Get("id").(string)
		userRole := session.Get("role").(string)
		if userId == "" || userRole == "" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		c.Locals("user.id", userId)
		c.Locals("user.role", userRole)

		return c.Next()
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// isDevUIPath checks if the request is for a dev UI path when NOT running in
// production
func (r *Router) isDevUIPath(c *fiber.Ctx) bool {
	if !r.config.IsProduction &&
		(strings.HasPrefix(c.OriginalURL(), "/node_modules/") ||
			strings.HasPrefix(c.OriginalURL(), "/.svelte-kit/") ||
			strings.HasPrefix(c.OriginalURL(), "/src/") ||
			strings.HasPrefix(c.OriginalURL(), "/@")) {
		return true
	}

	return false
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// isFavicon checks if the request is for a favicon
func (r *Router) isFavicon(c *fiber.Ctx) bool {
	return strings.HasPrefix(c.OriginalURL(), "/favicon.")
}
