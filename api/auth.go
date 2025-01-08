package api

import (
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/auth"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type authAPI struct {
	r *Router
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initFsRoutes initializes the filesystem routes
func (r *Router) initAuthRoutes() {
	authAPI := authAPI{r: r}

	authGroup := r.api.Group("/auth")

	authGroup.Post("/register", authAPI.register)
	authGroup.Post("/bootstrap", authAPI.bootstrap)
	authGroup.Post("/login", authAPI.login)
	authGroup.Post("/logout", authAPI.logout)

	authMeGroup := authGroup.Group("/me")
	authMeGroup.Get("/me", authAPI.me)
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
		DisplayName:  userReq.Username, // Set the display name to the username by default
		PasswordHash: auth.GeneratePassword(userReq.Password),
	}

	// The first user will always be an admin
	bootstrapAdmin, ok := c.Locals("bootstrapAdmin").(bool)
	if ok && bootstrapAdmin {
		user.Role = types.UserRoleAdmin
	} else {
		user.Role = types.UserRoleUser
	}

	err := api.r.dao.CreateUser(c.UserContext(), user)
	if err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return errorResponse(c, fiber.StatusBadRequest, "Username already exists", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error creating user", err)
	}

	session, err := api.r.sessionStore.Get(c)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error getting session", err)
	}

	session.Set("id", user.ID)
	session.Set("role", user.Role.String())
	if err := session.Save(); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error saving session", err)
	}

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

	err := api.r.dao.Get(c.UserContext(), user, options)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Invalid username and/or password", nil)
	}

	if !auth.ComparePassword(user.PasswordHash, userReq.Password) {
		return errorResponse(c, fiber.StatusUnauthorized, "Invalid username and/or password", nil)
	}

	session, err := api.r.sessionStore.Get(c)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error getting session", err)
	}

	session.Set("id", user.ID)
	session.Set("role", user.Role.String())
	if err := session.Save(); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error saving session", err)
	}

	return c.SendStatus(fiber.StatusOK)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api authAPI) logout(c *fiber.Ctx) error {
	session, err := api.r.sessionStore.Get(c)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error getting session", err)
	}

	session.Destroy()
	return c.SendStatus(fiber.StatusNoContent)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api authAPI) me(c *fiber.Ctx) error {
	userId, ok := c.Locals("user.id").(string)
	if !ok {
		return errorResponse(c, fiber.StatusUnauthorized, "Invalid user", nil)
	}

	user := &models.User{Base: models.Base{ID: userId}}
	err := api.r.dao.GetById(c.UserContext(), user)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error getting user information", err)
	}

	return c.Status(fiber.StatusOK).JSON(&UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        user.Role,
	})
}
