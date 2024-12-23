package auth

import (
	"testing"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

func Test_Token(t *testing.T) {
	t.Run("GenerateToken", func(t *testing.T) {
		u := &models.User{
			Username:     "test",
			PasswordHash: GeneratePassword("abcd1234"),
			Role:         types.UserRoleAdmin,
		}

		secret := "abcdefghijklmnopqrstuvwxyz0123456789"

		token, err := GenerateToken(secret, u)
		require.NoError(t, err)

		// Valid
		resp, err := ParseToken(secret, token)
		require.NoError(t, err)
		require.True(t, resp.Valid)

		// Error - malformed
		resp, err = ParseToken(secret, "test")
		require.EqualError(t, err, "token is malformed: token contains an invalid number of segments")
		require.Nil(t, resp)

		// Error - invalid signature
		resp, err = ParseToken(secret, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYWRtaW4iLCJzdWIiOiJ0ZXN0In0.g72Zj_f5DMsSoPpWdl06du3caajcGGfV6jyOeyDLzbw")
		require.EqualError(t, err, "token signature is invalid: signature is invalid")
		require.Nil(t, resp)
	})
}
