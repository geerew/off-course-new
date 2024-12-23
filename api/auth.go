package api

import (
	"log/slog"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/auth"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type authAPI struct {
	logger    *slog.Logger
	dao       *dao.DAO
	jwtSecret string
	r         *Router
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	cookie_access_token  = "access_token"
	cookie_refresh_token = "refresh_token"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initFsRoutes initializes the filesystem routes
func (r *Router) initAuthRoutes() {
	authAPI := authAPI{
		logger:    r.config.Logger,
		dao:       r.dao,
		jwtSecret: r.config.JwtSecret,
		r:         r,
	}

	authGroup := r.api.Group("/auth")

	authGroup.Post("/register", authAPI.register)
	authGroup.Post("/bootstrap", authAPI.bootstrap)
	authGroup.Post("/login", authAPI.login)
	authGroup.Get("/logout", authAPI.logout)
	authGroup.Get("/me", authAPI.me)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api authAPI) register(c *fiber.Ctx) error {
	userReq := &UserRequest{}

	if err := c.BodyParser(userReq); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	if userReq.Username == "" || userReq.Password == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Username and/or password cannot be empty", nil)
	}

	user := &models.User{
		Username:     userReq.Username,
		PasswordHash: auth.GeneratePassword(userReq.Password),
	}

	// The first user will always be an admin
	bootstrapAdmin, ok := c.Locals("bootstrapAdmin").(bool)
	if ok && bootstrapAdmin {
		user.Role = types.UserRoleAdmin
	} else {
		user.Role = types.UserRoleUser
	}

	err := api.dao.CreateUser(c.UserContext(), user)
	if err != nil {
		if strings.HasPrefix(err.Error(), "constraint failed: UNIQUE constraint failed") {
			return errorResponse(c, fiber.StatusBadRequest, "Username already exists", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error creating user", err)
	}

	token, err := auth.GenerateToken(api.jwtSecret, user)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error generating token", err)
	}

	c.Cookie(&fiber.Cookie{
		Name:     cookie_access_token,
		Value:    token,
		Expires:  time.Now().Add(15 * time.Minute),
		HTTPOnly: true,
		SameSite: "Strict",
	})

	c.Cookie(&fiber.Cookie{
		Name:     cookie_refresh_token,
		Value:    security.PseudorandomString(64),
		Expires:  time.Now().Add(24 * 7 * time.Hour),
		HTTPOnly: true,
		SameSite: "Strict",
	})

	return c.SendStatus(fiber.StatusCreated)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api authAPI) bootstrap(c *fiber.Ctx) error {
	c.Locals("bootstrapAdmin", true)
	err := api.register(c)

	if err == nil {
		api.r.setBootstrapped()
	}

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api authAPI) login(c *fiber.Ctx) error {
	userReq := &UserRequest{}

	if err := c.BodyParser(userReq); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	if userReq.Username == "" || userReq.Password == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Username and/or password cannot be empty", nil)
	}

	user := &models.User{}
	options := &database.Options{
		Where: squirrel.Eq{
			models.USER_TABLE + "." + models.USER_USERNAME: userReq.Username,
		},
	}

	err := api.dao.Get(c.UserContext(), user, options)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Invalid username and/or password", nil)
	}

	if !auth.ComparePassword(user.PasswordHash, userReq.Password) {
		return errorResponse(c, fiber.StatusUnauthorized, "Invalid username and/or password", nil)
	}

	token, err := auth.GenerateToken(api.jwtSecret, user)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error generating token", err)
	}

	c.Cookie(&fiber.Cookie{
		Name:     cookie_access_token,
		Value:    token,
		Expires:  time.Now().Add(15 * time.Minute),
		HTTPOnly: true,
		SameSite: "Strict",
	})

	c.Cookie(&fiber.Cookie{
		Name:     cookie_refresh_token,
		Value:    security.PseudorandomString(64),
		Expires:  time.Now().Add(24 * 7 * time.Hour),
		HTTPOnly: true,
		SameSite: "Strict",
	})

	return c.Status(fiber.StatusOK).JSON(&TokenResponse{Token: token})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api authAPI) logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     cookie_access_token,
		Expires:  time.Now().Add(time.Hour * -1),
		HTTPOnly: true,
		SameSite: "Strict",
	})

	c.Cookie(&fiber.Cookie{
		Name:     cookie_refresh_token,
		Expires:  time.Now().Add(time.Hour * -1),
		HTTPOnly: true,
		SameSite: "Strict",
	})

	return c.SendStatus(fiber.StatusNoContent)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api authAPI) me(c *fiber.Ctx) error {
	userId, ok := c.Locals("user.id").(string)
	if !ok {
		return errorResponse(c, fiber.StatusUnauthorized, "Invalid user", nil)
	}

	user := &models.User{Base: models.Base{ID: userId}}
	err := api.dao.GetById(c.UserContext(), user)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error getting user information", err)
	}

	return c.Status(fiber.StatusOK).JSON(&UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
	})
}
