package schema

import (
	"database/sql"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils"
	_ "github.com/mattn/go-sqlite3" // Import the SQLite3 driver
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type Base struct {
	ID int
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (b *Base) Define(c *ModelConfig) {
	c.Field("ID").Column("id")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type User struct {
	Base
	Name   string
	Age    int
	Number sql.NullInt16
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (u *User) Table() string {
	return "users"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (u *User) Define(c *ModelConfig) {
	c.Field("Base").Embedded()
	c.Field("Name").Column("name").NotNull()
	c.Field("Age").NotNull()
	c.Field("Number").Column("number")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setup(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY, 
			name TEXT NOT NULL,
			age TEXT NOT NULL,
			number INTEGER
		);
	`)
	require.NoError(t, err)

	return db
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Parse(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		sch, err := Parse(&User{})
		require.NoError(t, err)
		require.NotNil(t, sch)
	})

	t.Run("slice", func(t *testing.T) {
		var users []*User
		schema, err := Parse(users)
		require.NotNil(t, schema)
		require.NoError(t, err)
	})

	t.Run("nil", func(t *testing.T) {
		schema, err := Parse(nil)
		require.Nil(t, schema)
		require.ErrorIs(t, err, utils.ErrNilPtr)
	})

	t.Run("nil struct", func(t *testing.T) {
		var user *User
		schema, err := Parse(user)
		require.Nil(t, schema)
		require.ErrorIs(t, err, utils.ErrInvalidValue)
	})

	t.Run("not a modeler", func(t *testing.T) {
		schema, err := Parse(&struct{}{})
		require.Nil(t, schema)
		require.ErrorIs(t, err, utils.ErrNotModeler)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Scan(t *testing.T) {
	t.Run("struct success", func(t *testing.T) {
		db := setup(t)

		_, err := db.Exec(`INSERT INTO users (id, name, age) VALUES (?, ?, ?)`, 1, "John", 30)
		require.NoError(t, err)

		sch, err := Parse(&User{})
		require.NoError(t, err)
		require.NotNil(t, sch)

		rows, err := db.Query(`SELECT * FROM users WHERE id = ?`, 1)
		require.NoError(t, err)
		defer rows.Close()

		u := &User{}
		err = sch.Scan(rows, u)
		require.NoError(t, err)

		require.Equal(t, 1, u.ID)
		require.Equal(t, "John", u.Name)
		require.Equal(t, 30, u.Age)
		require.False(t, u.Number.Valid)
	})

	t.Run("slice success", func(t *testing.T) {
		db := setup(t)

		_, err := db.Exec(`INSERT INTO users (id, name, age) VALUES (?, ?, ?)`, 1, "John", 30)
		require.NoError(t, err)

		_, err = db.Exec(`INSERT INTO users (id, name, age) VALUES (?, ?, ?)`, 2, "Jane", 25)
		require.NoError(t, err)

		sch, err := Parse(&User{})
		require.NoError(t, err)
		require.NotNil(t, sch)

		rows, err := db.Query(`SELECT * FROM users`)
		require.NoError(t, err)
		defer rows.Close()

		u := []User{}
		err = sch.Scan(rows, &u)
		require.NoError(t, err)
		require.Len(t, u, 2)
	})

	t.Run("no rows", func(t *testing.T) {
		db := setup(t)

		sch, err := Parse(&User{})
		require.NoError(t, err)
		require.NotNil(t, sch)

		rows, err := db.Query(`SELECT * FROM users WHERE id = ?`, 1)
		require.NoError(t, err)
		defer rows.Close()

		u := &User{}
		err = sch.Scan(rows, u)
		require.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("not a pointer", func(t *testing.T) {
		db := setup(t)

		sch, err := Parse(&User{})
		require.NoError(t, err)
		require.NotNil(t, sch)

		rows, err := db.Query(`SELECT * FROM users WHERE id = ?`, 1)
		require.NoError(t, err)
		defer rows.Close()

		err = sch.Scan(rows, User{})
		require.ErrorIs(t, err, utils.ErrNotPtr)
	})

	t.Run("nil pointer", func(t *testing.T) {
		db := setup(t)

		sch, err := Parse(&User{})
		require.NoError(t, err)
		require.NotNil(t, sch)

		rows, err := db.Query(`SELECT * FROM users WHERE id = ?`, 1)
		require.NoError(t, err)
		defer rows.Close()

		var u *User
		err = sch.Scan(rows, u)
		require.ErrorIs(t, err, utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Benchmark_Create(b *testing.B) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(b, err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY, 
			name TEXT NOT NULL,
			age TEXT NOT NULL,
			number INTEGER
		);
	`)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sch, err := Parse(&User{})
		require.NoError(b, err)

		u := &User{
			Base: Base{
				ID: i,
			},
			Name: "John",
			Age:  30,
		}

		builder := sch.InsertBuilder(u)
		query, args, _ := builder.ToSql()

		_, err = db.Exec(query, args...)
		require.NoError(b, err)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Benchmark_Scan(b *testing.B) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(b, err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY, 
			name TEXT NOT NULL,
			age TEXT NOT NULL,
			number INTEGER
		);
	`)
	require.NoError(b, err)

	for i := 0; i < 1000; i++ {
		_, err = db.Exec(`INSERT INTO users (id, name, age) VALUES (?, ?, ?)`, i, "John", 30)
		require.NoError(b, err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sch, err := Parse(&User{})
		require.NoError(b, err)

		options := &database.Options{Where: squirrel.Eq{"id": (i % 1000)}}
		builder := sch.SelectBuilder(options).Limit(1)
		query, args, _ := builder.ToSql()

		rows, err := db.Query(query, args...)
		require.NoError(b, err)

		u := &User{}
		err = sch.Scan(rows, u)
		require.NoError(b, err)
	}
}
