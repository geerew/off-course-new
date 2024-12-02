package schema

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// BASE
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TestBase defines a test base struct
type TestBase struct {
	ID int
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Defined implements the `Definer` interface
func (b *TestBase) Define(c *ModelConfig) {
	c.Field("ID")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// POST
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TestPost defines a test post struct
type TestPost struct {
	TestBase
	UserID  int
	Title   string
	Content string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `Modeler` interface
func (p *TestPost) Table() string {
	return "posts"

}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Define implements the `Modeler` interface
func (p *TestPost) Define(c *ModelConfig) {
	c.Embedded("TestBase")

	c.Field("UserID").NotNull()
	c.Field("Title").NotNull().Mutable()
	c.Field("Content").NotNull().Mutable()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// PROFILE
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TestProfile defines a test profile struct
type TestProfile struct {
	TestBase
	UserID   int
	Name     string
	Username string
	Email    string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `Modeler` interface
func (p *TestProfile) Table() string {
	return "profiles"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Define implements the `Modeler` interface
func (p *TestProfile) Define(c *ModelConfig) {
	c.Embedded("TestBase")

	c.Field("UserID").NotNull()
	c.Field("Name").NotNull().Mutable()
	c.Field("Username").NotNull().Mutable()
	c.Field("Email").NotNull().Mutable()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// USER
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type TestUser struct {
	TestBase
	Profile  TestProfile
	Posts    []TestPost
	PtrPosts *[]*TestPost
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `Modeler` interface
func (u *TestUser) Table() string {
	return "users"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Define implements the `Modeler` interface
func (u *TestUser) Define(c *ModelConfig) {
	c.Embedded("TestBase")

	c.Relation("Profile").MatchOn("user_id")
	c.Relation("Posts").MatchOn("user_id")
	c.Relation("PtrPosts").MatchOn("user_id")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// setup creates a new in-memory SQLite database
func setup(tb testing.TB) *sql.DB {
	tb.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(tb, err)

	// Create tables
	_, err = db.Exec(`
		-- Create the users table
		CREATE TABLE users (
		    id INTEGER PRIMARY KEY AUTOINCREMENT
		);

		-- Create the profiles table
		CREATE TABLE profiles (
		    id INTEGER PRIMARY KEY AUTOINCREMENT,
		    user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
		    username TEXT NOT NULL,
		    email TEXT NOT NULL,
		    FOREIGN KEY (user_id) REFERENCES users(id)
		);

		-- Create the posts table
		CREATE TABLE posts (
		    id INTEGER PRIMARY KEY AUTOINCREMENT,
		    user_id INTEGER NOT NULL,
		    title TEXT NOT NULL,
		    content TEXT NOT NULL,
		    FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`)
	require.NoError(tb, err)

	// Schemas
	userSchema, err := Parse(&TestUser{})
	require.NoError(tb, err)

	profileSchema, err := Parse(&TestProfile{})
	require.NoError(tb, err)

	postSchema, err := Parse(&TestPost{})
	require.NoError(tb, err)

	// Insert John with a profile and 2 posts
	john := &TestUser{TestBase: TestBase{ID: 1}}
	_, err = userSchema.Insert(john, db)
	require.NoError(tb, err)

	_, err = profileSchema.Insert(&TestProfile{TestBase: TestBase{ID: 1}, UserID: john.ID, Name: "John", Username: "john_doe", Email: "john@test.com"}, db)
	require.NoError(tb, err)

	_, err = postSchema.Insert(&TestPost{TestBase: TestBase{ID: 1}, UserID: john.ID, Title: "Post 1 by John", Content: "This is the first post by John."}, db)
	require.NoError(tb, err)

	_, err = postSchema.Insert(&TestPost{TestBase: TestBase{ID: 2}, UserID: john.ID, Title: "Post 2 by John", Content: "This is the second post by John."}, db)
	require.NoError(tb, err)

	// Insert Jane with a profile and 1 post
	jane := &TestUser{TestBase: TestBase{ID: 2}}
	_, err = userSchema.Insert(jane, db)
	require.NoError(tb, err)

	_, err = profileSchema.Insert(&TestProfile{TestBase: TestBase{ID: 2}, UserID: jane.ID, Name: "Jane", Username: "jane_doe", Email: "jane@test.com"}, db)
	require.NoError(tb, err)

	_, err = postSchema.Insert(&TestPost{TestBase: TestBase{ID: 3}, UserID: jane.ID, Title: "Post by Jane", Content: "This is the post by Jane."}, db)
	require.NoError(tb, err)

	return db
}
