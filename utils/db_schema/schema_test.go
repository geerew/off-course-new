package db_schema

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
	TestBase   `db:"embedded"`
	Name       string        `db:"col:name;not null"`
	Age        sql.NullInt16 `db:"col:age"`
	Profession string        `db:"join:left;table:professions;this:id;that:user_id;col:profession;alias:prof"`

	// Should be ignored
	Ignore int
	TestIgnored
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var tableMapper = map[string]string{
	"TestUser": "test_users",
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_SchemaParse(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		schema, err := Parse(nil, tableMapper, &sync.Map{})
		require.Nil(t, schema)
		require.ErrorIs(t, err, ErrNilType)
	})

	t.Run("unsupported type", func(t *testing.T) {
		schema, err := Parse(10, tableMapper, &sync.Map{})
		require.Nil(t, schema)
		require.ErrorIs(t, err, ErrUnsupportedType)
	})

	t.Run("nil struct", func(t *testing.T) {
		var user *TestUser
		schema, err := Parse(user, tableMapper, &sync.Map{})
		require.Nil(t, schema)
		require.ErrorIs(t, err, ErrInvalidValue)
	})

	t.Run("nil table mapper", func(t *testing.T) {
		schema, err := Parse(TestUser{}, nil, &sync.Map{})
		require.Nil(t, schema)
		require.ErrorIs(t, err, ErrNilTableMapper)
	})

	t.Run("no table mapping", func(t *testing.T) {
		schema, err := Parse(TestUser{}, map[string]string{}, &sync.Map{})
		require.Nil(t, schema)
		require.ErrorContains(t, err, "unable to map db table name for TestUser")
	})

	t.Run("struct", func(t *testing.T) {
		schema, err := Parse(TestUser{}, tableMapper, &sync.Map{})
		require.NotNil(t, schema)
		require.NoError(t, err)

		require.Equal(t, 6, len(schema.Fields))
	})

	t.Run("slice", func(t *testing.T) {
		var users []*TestUser
		schema, err := Parse(users, tableMapper, &sync.Map{})
		require.NotNil(t, schema)
		require.NoError(t, err)
	})

	t.Run("cache", func(t *testing.T) {
		cache := &sync.Map{}
		schema, err := Parse([]*TestUser{}, tableMapper, cache)
		require.NotNil(t, schema)
		require.NoError(t, err)

		schema, err = Parse([]*TestUser{}, tableMapper, cache)
		require.NotNil(t, schema)
		require.NoError(t, err)
	})
}
