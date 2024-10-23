package schema

import (
	"reflect"
	"testing"

	"github.com/geerew/off-course/utils"
	"github.com/stretchr/testify/require"
)

// -------------------------------------------------------

func Test_ConcreteReflectValue(t *testing.T) {
	t.Run("non-pointer", func(t *testing.T) {
		i := 42
		v := reflect.ValueOf(i)

		result, err := concreteReflectValue(v)
		require.NoError(t, err)
		require.Equal(t, v, result)
		require.Equal(t, reflect.Int, result.Kind())
	})

	t.Run("pointer", func(t *testing.T) {
		i := 42
		v := reflect.ValueOf(&i)

		result, err := concreteReflectValue(v)
		require.NoError(t, err)
		require.Equal(t, reflect.Int, result.Kind())
	})

	t.Run("nested pointer", func(t *testing.T) {
		i := 42
		ptr := &i
		ptrPtr := &ptr
		v := reflect.ValueOf(&ptrPtr)

		result, err := concreteReflectValue(v)
		require.NoError(t, err)
		require.Equal(t, reflect.Int, result.Kind())
	})

	t.Run("nil pointer", func(t *testing.T) {
		var i *int
		v := reflect.ValueOf(i)

		result, err := concreteReflectValue(v)
		require.ErrorIs(t, err, utils.ErrInvalidValue)
		require.Equal(t, reflect.Invalid, result.Kind())
	})

	t.Run("nested pointer nil", func(t *testing.T) {
		var i **int
		v := reflect.ValueOf(i)

		result, err := concreteReflectValue(v)
		require.ErrorIs(t, err, utils.ErrInvalidValue)
		require.Equal(t, reflect.Invalid, result.Kind())
	})
}
