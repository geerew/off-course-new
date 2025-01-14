package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/auth"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAuth_Register(t *testing.T) {
	t.Run("201 (created)", func(t *testing.T) {
		router, ctx := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"username": "test", "password": "abcd1234" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		user := &models.User{}
		options := &database.Options{Where: squirrel.Eq{models.USER_TABLE + "." + models.USER_USERNAME: "test"}}
		require.NoError(t, router.dao.Get(ctx, user, options))
		require.NotEqual(t, "password", user.PasswordHash)
		require.Equal(t, types.UserRoleUser, user.Role)
	})

	t.Run("400 (bind error)", func(t *testing.T) {
		router, _ := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Error parsing data")
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		router, _ := setup(t)

		// Missing both
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username and/or password cannot be empty")

		// Missing password
		req = httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"username": "test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username and/or password cannot be empty")

		// Missing username
		req = httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"password": "test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username and/or password cannot be empty")

		// Both empty
		req = httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"username": "", "password": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username and/or password cannot be empty")
	})

	t.Run("400 (existing user)", func(t *testing.T) {
		router, _ := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"username": "test", "password": "abcd1234" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		// Same case
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username already exists")

		// Different case
		req = httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"username": "TEST", "password": "abcd1234" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username already exists")
	})
}

func TestAuth_Bootstrap(t *testing.T) {
	t.Run("201 (created)", func(t *testing.T) {
		router, ctx := setup(t)

		// Create user
		req := httptest.NewRequest(http.MethodPost, "/api/auth/bootstrap", strings.NewReader(`{"username": "test", "password": "abcd1234" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		user := &models.User{}
		options := &database.Options{Where: squirrel.Eq{models.USER_TABLE + "." + models.USER_USERNAME: "test"}}
		require.NoError(t, router.dao.Get(ctx, user, options))
		require.NotEqual(t, "password", user.PasswordHash)
		require.Equal(t, types.UserRoleAdmin, user.Role)
		require.True(t, router.isBootstrapped())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAuth_Login(t *testing.T) {
	t.Run("200 (success)", func(t *testing.T) {
		router, ctx := setup(t)

		user := &models.User{
			Username:     "test",
			DisplayName:  "Test",
			PasswordHash: auth.GeneratePassword("abcd1234"),
			Role:         types.UserRoleAdmin,
		}
		require.NoError(t, router.dao.CreateUser(ctx, user))

		u := &models.User{Base: models.Base{ID: user.ID}}
		require.NoError(t, router.dao.GetById(ctx, u))

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username": "test", "password": "abcd1234" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
	})

	t.Run("400 (bind error)", func(t *testing.T) {
		router, _ := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Error parsing data")
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		router, _ := setup(t)

		// Missing both
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username and/or password cannot be empty")

		// Missing password
		req = httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username": "test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username and/or password cannot be empty")

		// Missing username
		req = httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"password": "test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username and/or password cannot be empty")

		// Both empty
		req = httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username": "", "password": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username and/or password cannot be empty")
	})

	t.Run("401 (invalid user)", func(t *testing.T) {
		router, ctx := setup(t)

		user := &models.User{
			Username:     "test",
			DisplayName:  "Test",
			PasswordHash: auth.GeneratePassword("abcd1234"),
			Role:         types.UserRoleAdmin,
		}
		require.NoError(t, router.dao.CreateUser(ctx, user))

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username": "invalid", "password": "abcd1234" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, status)
		require.Contains(t, string(body), "Invalid username and/or password")
	})

	t.Run("401 (invalid password)", func(t *testing.T) {
		router, ctx := setup(t)

		user := &models.User{
			Username:     "test",
			DisplayName:  "Test",
			PasswordHash: auth.GeneratePassword("abcd1234"),
			Role:         types.UserRoleAdmin,
		}
		require.NoError(t, router.dao.CreateUser(ctx, user))

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username": "test", "password": "invalid" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, status)
		require.Contains(t, string(body), "Invalid username and/or password")
	})
}
