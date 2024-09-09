package daos

import (
	"database/sql"
	"testing"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func paramSetup(t *testing.T) (*ParamDao, database.Database) {
	t.Helper()

	dbManager := setup(t)
	paramDao := NewParamDao(dbManager.DataDb)
	return paramDao, dbManager.DataDb
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestParam_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		dao, _ := paramSetup(t)

		p, err := dao.Get("hasAdmin", nil)
		require.Nil(t, err)
		require.False(t, cast.ToBool(p.Value))
		require.False(t, p.CreatedAt.IsZero())
		require.False(t, p.UpdatedAt.IsZero())
	})

	t.Run("not found", func(t *testing.T) {
		dao, _ := paramSetup(t)

		p, err := dao.Get("test", nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		require.Nil(t, p)
	})

	t.Run("empty id", func(t *testing.T) {
		dao, _ := paramSetup(t)

		p, err := dao.Get("", nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		require.Nil(t, p)
	})

	t.Run("db error", func(t *testing.T) {
		dao, db := paramSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
		require.Nil(t, err)

		_, err = dao.Get("1234", nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestParam_Update(t *testing.T) {
	t.Run("hasAdmin", func(t *testing.T) {
		dao, _ := paramSetup(t)

		p, err := dao.Get("hasAdmin", nil)
		require.Nil(t, err)
		require.False(t, cast.ToBool(p.Value))
		require.False(t, p.CreatedAt.IsZero())
		require.False(t, p.UpdatedAt.IsZero())

		// Set to true
		p.Value = "true"
		require.Nil(t, dao.Update(p, nil))

		p2, err := dao.Get("hasAdmin", nil)
		require.Nil(t, err)
		require.True(t, cast.ToBool(p2.Value))
		require.Equal(t, p.CreatedAt, p2.CreatedAt)
		require.NotEqual(t, p.UpdatedAt, p2.UpdatedAt)
	})

	t.Run("empty id", func(t *testing.T) {
		dao, _ := paramSetup(t)

		err := dao.Update(&models.Param{}, nil)
		require.ErrorIs(t, err, ErrEmptyId)
	})

	t.Run("invalid id", func(t *testing.T) {
		dao, _ := paramSetup(t)

		p := &models.Param{}
		p.ID = "1234"
		require.Nil(t, dao.Update(p, nil))
	})

	t.Run("db error", func(t *testing.T) {
		dao, db := paramSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
		require.Nil(t, err)

		_, err = dao.Get("1234", nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table())
	})
}
