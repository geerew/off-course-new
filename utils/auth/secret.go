package auth

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/security"
)

// GetJwtSecret gets the JWT secret from the database, which is used to sign JWTs. If the
// secret does not exist, it is first created
func GetJwtSecret(db database.Database) (string, error) {
	jwtSecret := &models.Param{}
	options := &database.Options{
		Where: squirrel.Eq{models.PARAM_TABLE + "." + models.PARAM_KEY: "jwt_secret"},
	}

	dao := dao.NewDAO(db)

	err := dao.Get(context.Background(), jwtSecret, options)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}

	if err == sql.ErrNoRows {
		jwtSecret.Key = "jwt_secret"
		jwtSecret.Value = security.PseudorandomString(256)

		err = dao.CreateParam(context.Background(), jwtSecret)
		if err != nil {
			return "", err
		}
	}

	return jwtSecret.Value, nil
}
