package schema

import (
	"context"
	"database/sql"
	"reflect"
	"sync"
	"testing"

	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_FieldValueOf(t *testing.T) {
	t.Run("valueOf", func(t *testing.T) {
		schema, err := Parse(TestUser{}, &sync.Map{})
		require.NotNil(t, schema)
		require.NoError(t, err)

		user := TestUser{
			TestBase: TestBase{
				ID:        "1",
				CreatedAt: types.NowDateTime(),
				UpdatedAt: types.NowDateTime(),
			},
			Name: "test_name",
			Age:  sql.NullInt16{Int16: 10, Valid: true},
		}

		reflectValue := reflect.ValueOf(&user)

		var tests = []struct {
			key   string
			value any
		}{
			{"ID", user.ID},
			{"CreatedAt", user.CreatedAt},
			{"UpdatedAt", user.UpdatedAt},
			{"Name", user.Name},
			{"Age", user.Age},
		}

		for _, tt := range tests {
			t.Run(tt.key, func(t *testing.T) {
				fv, _ := schema.FieldsByName[tt.key].ValueOf(context.Background(), reflectValue)
				require.Equal(t, tt.value, fv)
			})
		}
	})
}

func Test_FieldSet(t *testing.T) {
	schema, err := Parse(TestUser{}, &sync.Map{})
	require.NotNil(t, schema)
	require.NoError(t, err)

	user := &TestUser{}
	reflectStruct := reflect.ValueOf(user)

	tests := []struct {
		key   string
		value any
	}{
		{"ID", "5"},
		{"CreatedAt", types.NowDateTime()},
		{"UpdatedAt", types.NowDateTime()},
		{"Name", "test_name_2"},
		{"Age", sql.NullInt16{Int16: 20, Valid: true}},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			err = schema.FieldsByName[tt.key].Set(context.Background(), reflectStruct, tt.value)
			require.Nil(t, err)

			fv, _ := schema.FieldsByName[tt.key].ValueOf(context.Background(), reflectStruct)
			require.Equal(t, tt.value, fv)
		})
	}
}
