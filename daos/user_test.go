package daos

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func userSetup(t *testing.T) (*UserDao, database.Database) {
	t.Helper()

	dbManager := setup(t)
	userDao := NewUserDao(dbManager.DataDb)
	return userDao, dbManager.DataDb
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUser_Count(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		dao, _ := userSetup(t)

		count, err := dao.Count(nil)
		require.Nil(t, err)
		require.Zero(t, count)
	})

	t.Run("entries", func(t *testing.T) {
		dao, _ := userSetup(t)

		users := []*models.User{
			{
				Username:     "admin",
				PasswordHash: "admin",
				Role:         types.UserRoleAdmin,
			},
			{
				Username:     "user",
				PasswordHash: "user",
				Role:         types.UserRoleUser,
			},
		}
		require.Nil(t, dao.Create(users[0], nil))
		require.Nil(t, dao.Create(users[1], nil))

		count, err := dao.Count(nil)
		require.Nil(t, err)
		require.Equal(t, count, 2)
	})

	t.Run("where", func(t *testing.T) {
		dao, _ := userSetup(t)

		users := []*models.User{
			{
				Username:     "admin",
				PasswordHash: "admin",
				Role:         types.UserRoleAdmin,
			},
			{
				Username:     "user",
				PasswordHash: "user",
				Role:         types.UserRoleUser,
			},
		}
		require.Nil(t, dao.Create(users[0], nil))
		require.Nil(t, dao.Create(users[1], nil))

		// ----------------------------
		// EQUALS username admin
		// ----------------------------
		count, err := dao.Count(&database.DatabaseParams{Where: squirrel.Eq{dao.Table() + ".username": users[0].Username}})
		require.Nil(t, err)
		require.Equal(t, 1, count)

		// ----------------------------
		// EQUALS role user
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.NotEq{dao.Table() + ".role": types.UserRoleUser}})
		require.Nil(t, err)
		require.Equal(t, 1, count)

		// ----------------------------
		// ERROR
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		require.Zero(t, count)
	})

	t.Run("db error", func(t *testing.T) {
		dao, db := userSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
		require.Nil(t, err)

		_, err = dao.Count(nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUser_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, _ := userSetup(t)

		u := &models.User{
			BaseModel: models.BaseModel{
				ID:        "1234",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Username:     "test",
			PasswordHash: "test",
			Role:         types.UserRoleAdmin,
		}

		err := dao.Create(u, nil)
		require.Nil(t, err, "Failed to create user")

		foundUser, err := dao.Get(u.Username, nil)
		require.Nil(t, err)
		require.Equal(t, u.ID, foundUser.ID)
		require.Equal(t, u.Username, foundUser.Username)
		require.Equal(t, u.PasswordHash, foundUser.PasswordHash)
		require.False(t, foundUser.CreatedAt.IsZero())
		require.False(t, foundUser.UpdatedAt.IsZero())
	})

	t.Run("duplicate course id", func(t *testing.T) {
		dao, _ := userSetup(t)

		u := &models.User{
			BaseModel: models.BaseModel{
				ID:        "1234",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Username:     "test",
			PasswordHash: "test",
			Role:         types.UserRoleAdmin,
		}

		err := dao.Create(u, nil)
		require.Nil(t, err, "Failed to create user")

		err = dao.Create(u, nil)
		require.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.username", dao.Table()))
	})

	t.Run("constraint errors", func(t *testing.T) {
		dao, _ := userSetup(t)

		// Missing course ID
		u := &models.User{}
		require.ErrorContains(t, dao.Create(u, nil), fmt.Sprintf("NOT NULL constraint failed: %s.username", dao.Table()))
		u.Username = ""
		require.ErrorContains(t, dao.Create(u, nil), fmt.Sprintf("NOT NULL constraint failed: %s.username", dao.Table()))
		u.Username = "test"

		// Password hash
		require.ErrorContains(t, dao.Create(u, nil), fmt.Sprintf("NOT NULL constraint failed: %s.password_hash", dao.Table()))
		u.PasswordHash = ""
		require.ErrorContains(t, dao.Create(u, nil), fmt.Sprintf("NOT NULL constraint failed: %s.password_hash", dao.Table()))
		u.PasswordHash = "test"

		// User role
		require.ErrorContains(t, dao.Create(u, nil), fmt.Sprintf("NOT NULL constraint failed: %s.role", dao.Table()))
		u.Role = types.UserRoleAdmin

		// Success
		require.Nil(t, dao.Create(u, nil))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUser_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		dao, _ := userSetup(t)

		users := []*models.User{
			{
				Username:     "admin",
				PasswordHash: "admin",
				Role:         types.UserRoleAdmin,
			},
			{
				Username:     "user",
				PasswordHash: "user",
				Role:         types.UserRoleUser,
			},
		}
		require.Nil(t, dao.Create(users[0], nil))
		require.Nil(t, dao.Create(users[1], nil))

		u, err := dao.Get(users[0].Username, nil)
		require.Nil(t, err)
		require.Equal(t, users[0].ID, u.ID)
		require.Equal(t, users[0].Username, u.Username)
		require.Equal(t, users[0].PasswordHash, u.PasswordHash)
		require.Equal(t, users[0].Role, u.Role)
	})

	t.Run("not found", func(t *testing.T) {
		dao, _ := userSetup(t)

		s, err := dao.Get("1234", nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		require.Nil(t, s)
	})

	t.Run("empty id", func(t *testing.T) {
		dao, _ := userSetup(t)

		s, err := dao.Get("", nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		require.Nil(t, s)
	})

	t.Run("db error", func(t *testing.T) {
		dao, db := userSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
		require.Nil(t, err)

		_, err = dao.Get("1234", nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUser_List(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		dao, _ := userSetup(t)

		courses, err := dao.List(nil, nil)
		require.Nil(t, err)
		require.Zero(t, courses)
	})

	t.Run("found", func(t *testing.T) {
		dao, _ := userSetup(t)

		// Create 5 users
		for i := 0; i < 5; i++ {
			u := &models.User{
				Username:     fmt.Sprintf("user%d", i),
				PasswordHash: fmt.Sprintf("password%d", i),
				Role:         types.UserRoleUser,
			}
			require.Nil(t, dao.Create(u, nil))
		}

		result, err := dao.List(nil, nil)
		require.Nil(t, err)
		require.Len(t, result, 5)
	})

	// t.Run("orderby", func(t *testing.T) {
	// 	dao, db := userSetup(t)

	// 	testData := NewTestBuilder(t).Db(db).Courses(3).Build()

	// 	// ----------------------------
	// 	// CREATED_AT DESC
	// 	// ----------------------------
	// 	dbParams := &database.DatabaseParams{OrderBy: []string{"created_at desc"}}
	// 	result, err := dao.List(dbParams, nil)
	// 	require.Nil(t, err)
	// 	require.Len(t, result, 3)
	// 	require.Equal(t, testData[2].ID, result[0].ID)

	// 	// ----------------------------
	// 	// CREATED_AT ASC
	// 	// ----------------------------
	// 	result, err = dao.List(&database.DatabaseParams{OrderBy: []string{"created_at asc"}}, nil)
	// 	require.Nil(t, err)
	// 	require.Len(t, result, 3)
	// 	require.Equal(t, testData[0].ID, result[0].ID)

	// 	// ----------------------------
	// 	// SCAN_STATUS DESC
	// 	// ----------------------------

	// 	// Create a scan for course 2 and 3
	// 	scanDao := NewScanDao(db)

	// 	testData[1].Scan = &models.Scan{CourseID: testData[1].ID}
	// 	require.Nil(t, scanDao.Create(testData[1].Scan, nil))
	// 	testData[2].Scan = &models.Scan{CourseID: testData[2].ID}
	// 	require.Nil(t, scanDao.Create(testData[2].Scan, nil))

	// 	// Set course 3 to processing
	// 	testData[2].Scan.Status = types.NewScanStatus(types.ScanStatusProcessing)
	// 	require.Nil(t, scanDao.Update(testData[2].Scan, nil))

	// 	result, err = dao.List(&database.DatabaseParams{OrderBy: []string{scanDao.Table() + ".status desc"}}, nil)
	// 	require.Nil(t, err)
	// 	require.Len(t, result, 3)

	// 	require.Equal(t, testData[0].ID, result[2].ID)
	// 	require.Equal(t, testData[1].ID, result[1].ID)
	// 	require.Equal(t, testData[2].ID, result[0].ID)

	// 	// ----------------------------
	// 	// SCAN_STATUS ASC
	// 	// ----------------------------
	// 	result, err = dao.List(&database.DatabaseParams{OrderBy: []string{scanDao.Table() + ".status asc"}}, nil)
	// 	require.Nil(t, err)
	// 	require.Len(t, result, 3)

	// 	require.Equal(t, testData[0].ID, result[0].ID)
	// 	require.Equal(t, testData[1].ID, result[1].ID)
	// 	require.Equal(t, testData[2].ID, result[2].ID)

	// 	// ----------------------------
	// 	// Error
	// 	// ----------------------------
	// 	dbParams = &database.DatabaseParams{OrderBy: []string{"unit_test asc"}}
	// 	result, err = dao.List(dbParams, nil)
	// 	require.ErrorContains(t, err, "no such column")
	// 	require.Nil(t, result)
	// })

	t.Run("where", func(t *testing.T) {
		dao, _ := userSetup(t)

		// Create 5 users (1 admin, 4 users)
		users := []*models.User{}
		for i := 0; i < 5; i++ {
			role := types.UserRoleUser
			if i == 0 {
				role = types.UserRoleAdmin
			}

			u := &models.User{
				Username:     fmt.Sprintf("user%d", i),
				PasswordHash: fmt.Sprintf("password%d", i),
				Role:         role,
			}
			require.Nil(t, dao.Create(u, nil))
			users = append(users, u)
		}

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		result, err := dao.List(&database.DatabaseParams{Where: squirrel.Eq{dao.Table() + ".username": users[2].Username}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 1)
		require.Equal(t, users[2].ID, result[0].ID)

		// ----------------------------
		// EQUALS role user
		// ----------------------------
		dbParams := &database.DatabaseParams{
			Where:   squirrel.Eq{dao.Table() + ".role": types.UserRoleUser},
			OrderBy: []string{"created_at asc"},
		}
		result, err = dao.List(dbParams, nil)
		require.Nil(t, err)
		require.Len(t, result, 4)

		// ----------------------------
		// ERROR
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Eq{"": ""}}, nil)
		require.ErrorContains(t, err, "syntax error")
		require.Nil(t, result)
	})

	t.Run("pagination", func(t *testing.T) {
		dao, _ := userSetup(t)

		// Create 17 users
		users := []*models.User{}
		for i := 0; i < 17; i++ {
			u := &models.User{
				Username:     fmt.Sprintf("user%d", i),
				PasswordHash: fmt.Sprintf("password%d", i),
				Role:         types.UserRoleUser,
			}
			require.Nil(t, dao.Create(u, nil))
			time.Sleep(1 * time.Millisecond)
			users = append(users, u)
		}

		// ----------------------------
		// Page 1 with 10 items
		// ----------------------------
		p := pagination.New(1, 10)

		result, err := dao.List(&database.DatabaseParams{Pagination: p}, nil)
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 17, p.TotalItems())
		require.Equal(t, users[0].ID, result[0].ID)
		require.Equal(t, users[9].ID, result[9].ID)

		// ----------------------------
		// Page 2 with 7 items
		// ----------------------------
		p = pagination.New(2, 10)

		result, err = dao.List(&database.DatabaseParams{Pagination: p}, nil)
		require.Nil(t, err)
		require.Len(t, result, 7)
		require.Equal(t, 17, p.TotalItems())
		require.Equal(t, users[10].ID, result[0].ID)
		require.Equal(t, users[16].ID, result[6].ID)
	})

	t.Run("db error", func(t *testing.T) {
		dao, db := userSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
		require.Nil(t, err)

		_, err = dao.List(nil, nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUser_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, _ := userSetup(t)

		// Create 2 users
		users := []*models.User{}
		for i := 0; i < 2; i++ {
			u := &models.User{
				Username:     fmt.Sprintf("user%d", i),
				PasswordHash: fmt.Sprintf("password%d", i),
				Role:         types.UserRoleUser,
			}
			require.Nil(t, dao.Create(u, nil))
			time.Sleep(1 * time.Millisecond)
			users = append(users, u)
		}

		err := dao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"username": users[1].Username}}, nil)
		require.Nil(t, err)
	})

	t.Run("no db params", func(t *testing.T) {
		dao, _ := userSetup(t)

		err := dao.Delete(nil, nil)
		require.ErrorIs(t, err, ErrMissingWhere)
	})

	t.Run("db error", func(t *testing.T) {
		dao, db := userSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table())
		require.Nil(t, err)

		err = dao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": "1234"}}, nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table())
	})
}
