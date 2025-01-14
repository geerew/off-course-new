package dao

import (
	"testing"
	"time"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)
		user := &models.User{Username: "admin", DisplayName: "Admin", PasswordHash: "password", Role: types.UserRoleAdmin}
		require.NoError(t, dao.CreateUser(ctx, user))
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.CreateUser(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		OriginalUser := &models.User{Username: "Admin", DisplayName: "Michael", Role: types.UserRoleAdmin, PasswordHash: "password"}
		require.NoError(t, dao.CreateUser(ctx, OriginalUser))

		time.Sleep(1 * time.Millisecond)

		newUser := &models.User{
			Base:         OriginalUser.Base,
			Username:     "nimda",
			DisplayName:  "Bob",
			Role:         types.UserRoleUser,
			PasswordHash: "new password",
		}
		require.NoError(t, dao.UpdateUser(ctx, newUser))

		userResult := &models.User{Base: models.Base{ID: OriginalUser.ID}}
		require.NoError(t, dao.GetById(ctx, userResult))
		require.Equal(t, OriginalUser.ID, userResult.ID)                     // No change
		require.Equal(t, OriginalUser.Username, userResult.Username)         // No change
		require.Equal(t, OriginalUser.Role, userResult.Role)                 // No change
		require.True(t, userResult.CreatedAt.Equal(OriginalUser.CreatedAt))  // No change
		require.Equal(t, newUser.DisplayName, userResult.DisplayName)        // Changed
		require.Equal(t, newUser.PasswordHash, userResult.PasswordHash)      // Changed
		require.False(t, userResult.UpdatedAt.Equal(OriginalUser.UpdatedAt)) // Changed
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		user := &models.User{Username: "Admin", DisplayName: "Michael", Role: types.UserRoleAdmin, PasswordHash: "password"}
		require.NoError(t, dao.CreateUser(ctx, user))

		// Empty ID
		user.ID = ""
		require.ErrorIs(t, dao.UpdateUser(ctx, user), utils.ErrInvalidId)

		// Nil Model
		require.ErrorIs(t, dao.UpdateUser(ctx, nil), utils.ErrNilPtr)
	})
}
