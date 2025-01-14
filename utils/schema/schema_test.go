package schema

import (
	"database/sql"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Parse(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		sch, err := Parse(&TestUser{})
		require.NoError(t, err)
		require.NotNil(t, sch)

		require.Equal(t, "users", sch.Table)
		require.Len(t, sch.Fields, 1)
		require.Len(t, sch.Relations, 3)
	})

	t.Run("slice", func(t *testing.T) {
		var users []*TestUser
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
		var user *TestUser
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

func Test_Select(t *testing.T) {
	t.Run("struct success", func(t *testing.T) {
		db := setup(t)

		sch, err := Parse(&TestUser{})
		require.NoError(t, err)
		require.NotNil(t, sch)

		u := &TestUser{}
		err = sch.Select(u, &database.Options{Where: squirrel.Eq{"id": 1}}, db)
		require.NoError(t, err)

		require.Equal(t, 1, u.ID)
		require.Equal(t, 1, u.Profile.ID)
		require.Len(t, u.Posts, 2)
		require.Equal(t, "Post 1 by John", u.Posts[0].Title)
		require.Equal(t, "Post 2 by John", u.Posts[1].Title)
		require.Len(t, *u.PtrPosts, 2)
	})

	t.Run("slice success", func(t *testing.T) {
		db := setup(t)

		sch, err := Parse(&TestUser{})
		require.NoError(t, err)
		require.NotNil(t, sch)

		u := []TestUser{}
		err = sch.Select(&u, nil, db)
		require.NoError(t, err)
		require.Len(t, u, 2)

		require.Equal(t, 1, u[0].ID)
		require.Equal(t, 1, u[0].Profile.ID)
		require.Len(t, u[0].Posts, 2)
		require.Equal(t, "Post 1 by John", u[0].Posts[0].Title)
		require.Equal(t, "Post 2 by John", u[0].Posts[1].Title)
		require.Len(t, *u[0].PtrPosts, 2)

		require.Equal(t, 2, u[1].ID)
		require.Equal(t, 2, u[1].Profile.ID)
		require.Len(t, u[1].Posts, 1)
		require.Equal(t, "Post by Jane", u[1].Posts[0].Title)
		require.Len(t, *u[1].PtrPosts, 1)
	})

	t.Run("no relation found", func(t *testing.T) {
		db := setup(t)

		userSchema, err := Parse(&TestUser{})
		require.NoError(t, err)
		require.NotNil(t, userSchema)

		profileSchema, err := Parse(&TestProfile{})
		require.NoError(t, err)
		require.NotNil(t, profileSchema)

		postSchema, err := Parse(&TestPost{})
		require.NoError(t, err)
		require.NotNil(t, postSchema)

		// Delete all profiles
		_, err = profileSchema.Delete(nil, db)
		require.NoError(t, err)

		// Delete all posts
		_, err = postSchema.Delete(nil, db)
		require.NoError(t, err)

		u := &TestUser{}
		err = userSchema.Select(u, &database.Options{Where: squirrel.Eq{"id": 1}}, db)
		require.NoError(t, err)

		require.Equal(t, 1, u.ID)
		require.Equal(t, TestProfile{}, u.Profile)
		require.Len(t, u.Posts, 0)
	})

	t.Run("some relations found", func(t *testing.T) {
		db := setup(t)

		userSchema, err := Parse(&TestUser{})
		require.NoError(t, err)
		require.NotNil(t, userSchema)

		profileSchema, err := Parse(&TestProfile{})
		require.NoError(t, err)
		require.NotNil(t, profileSchema)

		postSchema, err := Parse(&TestPost{})
		require.NoError(t, err)
		require.NotNil(t, postSchema)

		// Delete Johns profile
		_, err = profileSchema.Delete(&database.Options{Where: squirrel.Eq{"id": 1}}, db)
		require.NoError(t, err)

		// Delete Johns posts
		_, err = postSchema.Delete(&database.Options{Where: squirrel.Eq{"user_id": 1}}, db)
		require.NoError(t, err)

		u := []*TestUser{}
		err = userSchema.Select(&u, nil, db)
		require.NoError(t, err)

		require.Len(t, u, 2)
		require.Equal(t, 1, u[0].ID)
		require.Equal(t, TestProfile{}, u[0].Profile)
		require.Len(t, u[0].Posts, 0)

		require.Equal(t, 2, u[1].ID)
		require.Equal(t, 2, u[1].Profile.ID)
		require.Len(t, u[1].Posts, 1)
	})

	t.Run("no rows", func(t *testing.T) {
		db := setup(t)

		sch, err := Parse(&TestUser{})
		require.NoError(t, err)
		require.NotNil(t, sch)

		// Delete everything from the users table
		_, err = sch.Delete(nil, db)
		require.NoError(t, err)

		u := &TestUser{}
		err = sch.Select(u, &database.Options{Where: squirrel.Eq{"id": 1}}, db)
		require.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("not a pointer", func(t *testing.T) {
		db := setup(t)

		sch, err := Parse(&TestUser{})
		require.NoError(t, err)
		require.NotNil(t, sch)

		err = sch.Select(TestUser{}, nil, db)
		require.ErrorIs(t, err, utils.ErrNotPtr)
	})

	t.Run("nil pointer", func(t *testing.T) {
		db := setup(t)

		sch, err := Parse(&TestUser{})
		require.NoError(t, err)
		require.NotNil(t, sch)

		var u *TestUser
		err = sch.Select(u, nil, db)
		require.ErrorIs(t, err, utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Count(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db := setup(t)

		sch, err := Parse(&TestUser{})
		require.NoError(t, err)
		require.NotNil(t, sch)

		count, err := sch.Count(&database.Options{Where: squirrel.Eq{"id": 1}}, db)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})

	t.Run("no rows", func(t *testing.T) {
		db := setup(t)

		sch, err := Parse(&TestUser{})
		require.NoError(t, err)
		require.NotNil(t, sch)

		// Delete everything from the users table
		_, err = sch.Delete(nil, db)
		require.NoError(t, err)

		count, err := sch.Count(nil, db)
		require.NoError(t, err)
		require.Zero(t, count)

	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Benchmark_Create(b *testing.B) {
	db := setup(b)

	userSchema, err := Parse(&TestUser{})
	require.NoError(b, err)
	require.NotNil(b, userSchema)

	_, err = userSchema.Delete(nil, db)
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sch, err := Parse(&TestUser{})
		require.NoError(b, err)

		u := &TestUser{
			TestBase: TestBase{
				ID: i,
			},
		}

		builder := sch.InsertBuilder(u)
		query, args, _ := builder.ToSql()

		_, err = db.Exec(query, args...)
		require.NoError(b, err)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Benchmark_Scan(b *testing.B) {
	db := setup(b)

	userSchema, err := Parse(&TestUser{})
	require.NoError(b, err)
	require.NotNil(b, userSchema)

	// Empty users
	_, err = userSchema.Delete(nil, db)
	require.NoError(b, err)

	profileSchema, err := Parse(&TestProfile{})
	require.NoError(b, err)
	require.NotNil(b, profileSchema)

	// Empty profiles
	_, err = profileSchema.Delete(nil, db)
	require.NoError(b, err)

	// Insert 1000 users and profiles
	for i := 0; i < 1000; i++ {
		_, err = db.Exec(`INSERT INTO users (id) VALUES (?)`, i)
		require.NoError(b, err)

		_, err = db.Exec(`INSERT INTO profiles (user_id, name, username, email) VALUES (?, ?, ?, ?)`, i, "John Doe", "johndoe", "john@test.com")
		require.NoError(b, err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		u := &TestUser{}
		err = userSchema.Select(u, &database.Options{Where: squirrel.Eq{"id": (i % 1000)}}, db)
		require.NoError(b, err)
	}
}
