package daos

import (
	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AttachmentDao is the data access object for attachments
type AttachmentDao struct {
	BaseDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewAttachmentDao returns a new AttachmentDao
func NewAttachmentDao(db database.Database) *AttachmentDao {
	return &AttachmentDao{
		BaseDao: BaseDao{
			db:    db,
			table: "attachments",
		},
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count counts the attachments
func (dao *AttachmentDao) Count(dbParams *database.DatabaseParams, tx *database.Tx) (int, error) {
	return genericCount(dao, dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create creates an attachment
func (dao *AttachmentDao) Create(a *models.Attachment, tx *database.Tx) error {
	if a.ID == "" {
		a.RefreshId()
	}

	a.RefreshCreatedAt()
	a.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.Table()).
		SetMap(modelToMapOrPanic(a)).
		ToSql()

	execFn := dao.db.Exec
	if tx != nil {
		execFn = tx.Exec
	}

	_, err := execFn(query, args...)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get gets attachment with the given ID
func (dao *AttachmentDao) Get(id string, tx *database.Tx) (*models.Attachment, error) {
	selectColumns, _ := tableColumnsOrPanic(models.Attachment{}, dao.Table())

	dbParams := &database.DatabaseParams{
		Columns: selectColumns,
		Where:   squirrel.Eq{dao.Table() + ".id": id},
	}

	return genericGet(dao, dbParams, dao.scanRow, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List lists attachments
func (dao *AttachmentDao) List(dbParams *database.DatabaseParams, tx *database.Tx) ([]*models.Attachment, error) {
	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	selectColumns, orderByColumns := tableColumnsOrPanic(models.Attachment{}, dao.Table())

	dbParams.Columns = selectColumns

	// Remove invalid orderBy columns
	dbParams.OrderBy = genericProcessOrderBy(dbParams.OrderBy, orderByColumns, dao, false)

	return genericList(dao, dbParams, dao.scanRow, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes attachments based upon the where clause
func (dao *AttachmentDao) Delete(dbParams *database.DatabaseParams, tx *database.Tx) error {
	return genericDelete(dao, dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans an attachment row
func (dao *AttachmentDao) scanRow(scannable Scannable) (*models.Attachment, error) {
	var a models.Attachment

	err := scannable.Scan(
		&a.ID,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.CourseID,
		&a.AssetID,
		&a.Title,
		&a.Path,
	)

	if err != nil {
		return nil, err
	}

	return &a, nil
}
