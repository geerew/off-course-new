package auth

import (
	"time"

	"github.com/geerew/off-course/models"
	"github.com/golang-jwt/jwt/v5"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GenerateToken generates a token
func GenerateToken(secret string, user *models.User) (string, error) {
	now := time.Now()
	expires := now.Add(15 * time.Minute)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
		"role":     user.Role,
		"iat":      now.Unix(),
		"exp":      expires.Unix(),
		"username": user.Username,
	})

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ParseToken returns the token
func ParseToken(secret, token string) (*jwt.Token, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	return t, nil
}
