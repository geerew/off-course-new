package dao

import (
	"testing"
	"time"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateParam(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)
		param := &models.Param{Key: "param 1", Value: "value 1"}
		require.NoError(t, dao.CreateParam(ctx, param))
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.CreateParam(ctx, nil), utils.ErrNilPtr)
	})

	t.Run("duplicate", func(t *testing.T) {
		dao, ctx := setup(t)

		param := &models.Param{Key: "param 1", Value: "value 1"}
		require.NoError(t, dao.CreateParam(ctx, param))

		require.ErrorContains(t, dao.CreateParam(ctx, param), "UNIQUE constraint failed: "+models.PARAM_TABLE+".key")
	})
}

func Test_GetParamByKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)
		param := &models.Param{Key: "param 1", Value: "value 1"}
		require.NoError(t, dao.CreateParam(ctx, param))

		paramResult := &models.Param{Key: param.Key}
		require.NoError(t, dao.GetParamByKey(ctx, paramResult))
		require.Equal(t, param.ID, paramResult.ID)
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		// Nil model
		require.ErrorIs(t, dao.GetParamByKey(ctx, nil), utils.ErrNilPtr)

		// Invalid key
		param := &models.Param{Base: models.Base{ID: "1234"}}
		require.ErrorIs(t, dao.GetParamByKey(ctx, param), utils.ErrInvalidKey)
	})

	t.Run("duplicate", func(t *testing.T) {
		dao, ctx := setup(t)

		param := &models.Param{Key: "param 1", Value: "value 1"}
		require.NoError(t, dao.CreateParam(ctx, param))

		require.ErrorContains(t, dao.CreateParam(ctx, param), "UNIQUE constraint failed: "+models.PARAM_TABLE+".key")
	})
}

func Test_UpdateParam(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		originalParam := &models.Param{Key: "param 1", Value: "value 1"}
		require.NoError(t, dao.CreateParam(ctx, originalParam))

		time.Sleep(1 * time.Millisecond)

		newParam := &models.Param{
			Base:  originalParam.Base,
			Key:   "new key",   // Immutable
			Value: "new value", // Mutable
		}
		require.NoError(t, dao.UpdateParam(ctx, newParam))

		paramResult := &models.Param{Base: models.Base{ID: originalParam.ID}}
		require.NoError(t, dao.GetById(ctx, paramResult))
		require.Equal(t, originalParam.ID, paramResult.ID)                     // No change
		require.Equal(t, originalParam.Key, paramResult.Key)                   // No change
		require.True(t, paramResult.CreatedAt.Equal(originalParam.CreatedAt))  // No change
		require.Equal(t, newParam.Value, paramResult.Value)                    // Changed
		require.False(t, paramResult.UpdatedAt.Equal(originalParam.UpdatedAt)) // Changed
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		param := &models.Param{Key: "param 1", Value: "value 1"}
		require.NoError(t, dao.CreateParam(ctx, param))

		// Empty ID
		param.ID = ""
		require.ErrorIs(t, dao.UpdateParam(ctx, param), utils.ErrInvalidId)

		// Nil Model
		require.ErrorIs(t, dao.UpdateParam(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestParam_Get(t *testing.T) {
// 	t.Run("found", func(t *testing.T) {
// 		dao, _ := paramSetup(t)

// 		p, err := dao.Get("hasAdmin", nil)
// 		require.Nil(t, err)
// 		require.False(t, cast.ToBool(p.Value))
// 		require.False(t, p.CreatedAt.IsZero())
// 		require.False(t, p.UpdatedAt.IsZero())
// 	})

// 	t.Run("not found", func(t *testing.T) {
// 		dao, _ := paramSetup(t)

// 		p, err := dao.Get("test", nil)
// 		require.ErrorIs(t, err, sql.ErrNoRows)
// 		require.Nil(t, p)
// 	})

// 	t.Run("empty id", func(t *testing.T) {
// 		dao, _ := paramSetup(t)

// 		p, err := dao.Get("", nil)
// 		require.ErrorIs(t, err, sql.ErrNoRows)
// 		require.Nil(t, p)
// 	})

// 	t.Run("db error", func(t *testing.T) {
// 		dao, db := paramSetup(t)

// 		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
// 		require.Nil(t, err)

// 		_, err = dao.Get("1234", nil)
// 		require.ErrorContains(t, err, "no such table: "+dao.Table())
// 	})
// }

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestParam_Update(t *testing.T) {
// 	t.Run("hasAdmin", func(t *testing.T) {
// 		dao, _ := paramSetup(t)

// 		p, err := dao.Get("hasAdmin", nil)
// 		require.Nil(t, err)
// 		require.False(t, cast.ToBool(p.Value))
// 		require.False(t, p.CreatedAt.IsZero())
// 		require.False(t, p.UpdatedAt.IsZero())

// 		// Set to true
// 		p.Value = "true"
// 		require.Nil(t, dao.Update(p, nil))

// 		p2, err := dao.Get("hasAdmin", nil)
// 		require.Nil(t, err)
// 		require.True(t, cast.ToBool(p2.Value))
// 		require.Equal(t, p.CreatedAt, p2.CreatedAt)
// 		require.NotEqual(t, p.UpdatedAt, p2.UpdatedAt)
// 	})

// 	t.Run("empty id", func(t *testing.T) {
// 		dao, _ := paramSetup(t)

// 		err := dao.Update(&models.Param{}, nil)
// 		require.ErrorIs(t, err, ErrEmptyId)
// 	})

// 	t.Run("invalid id", func(t *testing.T) {
// 		dao, _ := paramSetup(t)

// 		p := &models.Param{}
// 		p.ID = "1234"
// 		require.Nil(t, dao.Update(p, nil))
// 	})

// 	t.Run("db error", func(t *testing.T) {
// 		dao, db := paramSetup(t)

// 		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
// 		require.Nil(t, err)

// 		_, err = dao.Get("1234", nil)
// 		require.ErrorContains(t, err, "no such table: "+dao.Table())
// 	})
// }
