package api

import (
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/auth"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/golang-jwt/jwt/v5"
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
		if r.isDevUIPath(c) || r.isFavicon(c) {
			return c.Next()
		}

		// When bootstrapping, ignore auth entirely. All invalid requests will be
		// handled by the bootstrap middleware
		if bootstrapping, ok := c.Locals("bootstrapping").(bool); ok && bootstrapping {
			return c.Next()
		}

		// Always allow logout
		if strings.HasPrefix(c.OriginalURL(), "/api/auth/logout") {
			return c.Next()
		}

		access_token := c.Cookies(cookie_access_token)
		if access_token == "" {
			// API check
			if strings.HasPrefix(c.OriginalURL(), "/api") {
				if strings.HasPrefix(c.OriginalURL(), "/api/auth/login") ||
					strings.HasPrefix(c.OriginalURL(), "/api/auth/register") {
					return c.Next()
				} else {
					return c.SendStatus(fiber.StatusForbidden)
				}
			}

			// UI check
			if strings.HasPrefix(c.OriginalURL(), "/auth/login") || strings.HasPrefix(c.OriginalURL(), "/auth/register") {
				return c.Next()
			} else {
				return c.Redirect("/auth/login/")
			}
		}

		// Validate the token
		token, err := auth.ParseToken(r.config.JwtSecret, access_token)
		if err != nil {
			utils.Errf("Failed to validate claim: %s\n", err.Error())
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		// If the claim is not valid, redirect to login
		if !token.Valid {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		// UI - If they are authenticated but trying to hit /auth/login or /auth/register,
		// redirect to /
		if c.OriginalURL() == "/auth/login/" || c.OriginalURL() == "/auth/register/" {
			return c.Redirect("/")
		}

		// If they are authenticated but trying to hit /api/auth/..., return 200
		if strings.HasPrefix(c.OriginalURL(), "/api/auth/") && !strings.HasPrefix(c.OriginalURL(), "/api/auth/me") {
			return c.SendStatus(fiber.StatusOK)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		if role, ok := claims["role"].(string); !ok {
			return c.SendStatus(fiber.StatusUnauthorized)
		} else {
			c.Locals("user.role", role)
		}

		if id, ok := claims["sub"].(string); !ok {
			return c.SendStatus(fiber.StatusUnauthorized)
		} else {
			c.Locals("user.id", id)
		}

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
