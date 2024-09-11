package daos

import (
	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UserDao is the data access object for users
type UserDao struct {
	BaseDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewUserDao returns a new UserDao
func NewUserDao(db database.Database) *UserDao {
	return &UserDao{
		BaseDao: BaseDao{
			db:    db,
			table: "users",
		},
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count counts the users
func (dao *UserDao) Count(dbParams *database.DatabaseParams, tx *database.Tx) (int, error) {
	return genericCount(dao, dbParams, tx)
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
		SetMap((toDBMapOrPanic(u))).
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
	selectColumns, _ := tableColumnsOrPanic(models.User{}, dao.Table())

	dbParams := &database.DatabaseParams{
		Columns: selectColumns,
		Where:   squirrel.Eq{dao.Table() + ".username": username},
	}

	return genericGet(dao, dbParams, dao.scanRow, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List lists users
func (dao *UserDao) List(dbParams *database.DatabaseParams, tx *database.Tx) ([]*models.User, error) {
	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	selectColumns, orderByColumns := tableColumnsOrPanic(models.User{}, dao.Table())

	dbParams.Columns = selectColumns

	// Remove invalid orderBy columns
	dbParams.OrderBy = genericProcessOrderBy(dbParams.OrderBy, orderByColumns, dao, false)

	return genericList(dao, dbParams, dao.scanRow, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes users based on the where clause
func (dao *UserDao) Delete(dbParams *database.DatabaseParams, tx *database.Tx) error {
	return genericDelete(dao, dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans a course row
func (dao *UserDao) scanRow(scannable Scannable) (*models.User, error) {
	var u models.User

	err := scannable.Scan(
		&u.ID,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.Username,
		&u.PasswordHash,
		&u.Role,
	)

	if err != nil {
		return nil, err
	}

	return &u, nil
}
