package dao

import (
	"testing"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)
		user := &models.User{Username: "admin", PasswordHash: "password", Role: types.UserRoleAdmin}
		require.NoError(t, dao.CreateUser(ctx, user))
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.CreateUser(ctx, nil), utils.ErrNilPtr)
	})
}
