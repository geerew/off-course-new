package daos

import (
	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UserDao is the data access object for users
type UserDao struct {
	db    database.Database
	table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewUserDao returns a new UserDao
func NewUserDao(db database.Database) *UserDao {
	return &UserDao{
		db:    db,
		table: "users",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table returns the table name
func (dao *UserDao) Table() string {
	return dao.table
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count counts the users
func (dao *UserDao) Count(dbParams *database.DatabaseParams, tx *database.Tx) (int, error) {
	queryRowFn := dao.db.QueryRow
	if tx != nil {
		queryRowFn = tx.QueryRow
	}

	return GenericCount(dao.countSelect(), dao.Table(), dbParams, queryRowFn)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create creates a user
func (dao *UserDao) Create(u *models.User, tx *database.Tx) error {
	if u.ID == "" {
		u.RefreshId()
	}

	u.RefreshCreatedAt()
	u.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.Table()).
		SetMap(dao.data(u)).
		ToSql()

	execFn := dao.db.Exec
	if tx != nil {
		execFn = tx.Exec
	}

	_, err := execFn(query, args...)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get gets user with the given username
func (dao *UserDao) Get(username string, tx *database.Tx) (*models.User, error) {
	generic := NewGenericDao(dao.db, dao)

	courseDbParams := &database.DatabaseParams{
		Columns: dao.columns(),
		Where:   squirrel.Eq{dao.Table() + ".username": username},
	}

	row, err := generic.Get(courseDbParams, tx)
	if err != nil {
		return nil, err
	}

	course, err := dao.scanRow(row)
	if err != nil {
		return nil, err
	}

	return course, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List lists users
func (dao *UserDao) List(dbParams *database.DatabaseParams, tx *database.Tx) ([]*models.User, error) {
	generic := NewGenericDao(dao.db, dao)

	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	// Process the order by clauses
	dbParams.OrderBy = dao.ProcessOrderBy(dbParams.OrderBy, false)

	// Default the columns if not specified
	if len(dbParams.Columns) == 0 {
		dbParams.Columns = dao.columns()
	}

	rows, err := generic.List(dbParams, tx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User

	for rows.Next() {
		u, err := dao.scanRow(rows)
		if err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes users based on the where clause
func (dao *UserDao) Delete(dbParams *database.DatabaseParams, tx *database.Tx) error {
	if dbParams == nil || dbParams.Where == nil {
		return ErrMissingWhere
	}

	generic := NewGenericDao(dao.db, dao)
	return generic.Delete(dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// countSelect returns the default select builder for counting
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *UserDao) countSelect() squirrel.SelectBuilder {
	return squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.Table()).
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *UserDao) baseSelect() squirrel.SelectBuilder {
	return dao.countSelect()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// columns returns the columns to select
func (dao *UserDao) columns() []string {
	return []string{
		dao.Table() + ".*",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// data generates a map of key/values for a course
func (dao *UserDao) data(u *models.User) map[string]any {
	return map[string]any{
		"id":            u.ID,
		"username":      NilStr(u.Username),
		"password_hash": NilStr(u.PasswordHash),
		"role":          NilStr(u.Role.String()),
		"created_at":    FormatTime(u.CreatedAt),
		"updated_at":    FormatTime(u.UpdatedAt),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ProcessOrderBy takes an array of strings representing orderBy clauses and returns a processed
// version of this array
//
// It will creates a new list of valid table columns based upon columns() for the current
// DAO
func (dao *UserDao) ProcessOrderBy(orderBy []string, explicit bool) []string {
	if len(orderBy) == 0 {
		return orderBy
	}

	generic := NewGenericDao(dao.db, dao)
	return generic.ProcessOrderBy(orderBy, dao.columns(), explicit)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans a course row
func (dao *UserDao) scanRow(scannable Scannable) (*models.User, error) {
	var u models.User

	// Nullable fields
	var createdAt string
	var updatedAt string

	err := scannable.Scan(
		&u.ID,
		&u.Username,
		&u.PasswordHash,
		&u.Role,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	if u.CreatedAt, err = ParseTime(createdAt); err != nil {
		return nil, err
	}

	if u.UpdatedAt, err = ParseTime(updatedAt); err != nil {
		return nil, err
	}

	return &u, nil
}
