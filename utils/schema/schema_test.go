package schema

import (
	"database/sql"
	"sync"
	"testing"

	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type TestBase struct {
	ID        string         `db:"col:id;not null"`
	CreatedAt types.DateTime `db:"col:created_at;not null"`
	UpdatedAt types.DateTime `db:"col:updated_at;not null"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type TestIgnored struct {
	Other string `db:"col:other;not null"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type TestUser struct {
	TestBase `db:"embedded"`
	Name     string        `db:"col:name;not null"`
	Age      sql.NullInt16 `db:"col:age"`

	// Should be ignored
	Ignore int
	TestIgnored
}

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_SchemaParse(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		schema, err := Parse(nil, &sync.Map{})
		require.Nil(t, schema)
		require.ErrorIs(t, err, ErrNilType)
	})

	t.Run("unsupported type", func(t *testing.T) {
		schema, err := Parse(10, &sync.Map{})
		require.Nil(t, schema)
		require.ErrorIs(t, err, ErrUnsupportedType)
	})

	t.Run("nil struct", func(t *testing.T) {
		var user *TestUser
		schema, err := Parse(user, &sync.Map{})
		require.Nil(t, schema)
		require.ErrorIs(t, err, ErrInvalidValue)
	})

	t.Run("struct", func(t *testing.T) {
		schema, err := Parse(TestUser{}, &sync.Map{})
		require.NotNil(t, schema)
		require.NoError(t, err)

		require.Equal(t, 5, len(schema.Fields))
	})

	t.Run("slice", func(t *testing.T) {
		var users []*TestUser
		schema, err := Parse(users, &sync.Map{})
		require.NotNil(t, schema)
		require.NoError(t, err)
	})

	t.Run("cache", func(t *testing.T) {
		cache := &sync.Map{}
		var users []*TestUser
		schema, err := Parse(users, cache)
		require.NotNil(t, schema)
		require.NoError(t, err)

		schema, err = Parse(users, cache)
		require.NotNil(t, schema)
		require.NoError(t, err)
	})
}
