package daos

import (
	"database/sql"
	"slices"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetDao is the data access object for assets
type AssetDao struct {
	BaseDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewAssetDao returns a new AssetDao
func NewAssetDao(db database.Database) *AssetDao {
	return &AssetDao{
		BaseDao: BaseDao{
			db:    db,
			table: "assets",
		},
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count counts the assets
func (dao *AssetDao) Count(dbParams *database.DatabaseParams, tx *database.Tx) (int, error) {
	return GenericCount(dao, dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create creates an asset
func (dao *AssetDao) Create(a *models.Asset, tx *database.Tx) error {
	if a.Prefix.Valid && a.Prefix.Int16 < 0 {
		return ErrInvalidPrefix
	}

	if a.ID == "" {
		a.RefreshId()
	}

	a.RefreshCreatedAt()
	a.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.Table()).
		SetMap(dao.data(a)).
		ToSql()

	execFn := dao.db.Exec
	if tx != nil {
		execFn = tx.Exec
	}

	_, err := execFn(query, args...)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get gets an asset with the given ID
func (dao *AssetDao) Get(id string, dbParams *database.DatabaseParams, tx *database.Tx) (*models.Asset, error) {
	assetDbParams := &database.DatabaseParams{
		Columns: dao.columns(),
		Where:   squirrel.Eq{dao.Table() + ".id": id},
	}

	asset, err := GenericGet(dao, assetDbParams, dao.scanRow, tx)
	if err != nil {
		return nil, err
	}

	// Get the attachments
	attachmentDao := NewAttachmentDao(dao.db)
	if dbParams != nil && slices.Contains(dbParams.IncludeRelations, attachmentDao.Table()) {
		// Set the DB params
		attachmentDbParams := &database.DatabaseParams{
			OrderBy: GenericProcessOrderBy(dbParams.OrderBy, attachmentDao.columns(), true),
			Where:   squirrel.Eq{"asset_id": asset.ID},
		}

		attachments, err := attachmentDao.List(attachmentDbParams, tx)
		if err != nil {
			return nil, err
		}

		asset.Attachments = attachments
	}

	return asset, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List lists assets
func (dao *AssetDao) List(dbParams *database.DatabaseParams, tx *database.Tx) ([]*models.Asset, error) {
	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	origOrderBy := dbParams.OrderBy
	dbParams.OrderBy = GenericProcessOrderBy(dbParams.OrderBy, dao.columns(), false)

	// Default the columns if not specified
	if len(dbParams.Columns) == 0 {
		dbParams.Columns = dao.columns()
	}

	assets, err := GenericList(dao, dbParams, dao.scanRow, tx)
	if err != nil {
		return nil, err
	}

	// Get the attachments
	attachmentDao := NewAttachmentDao(dao.db)
	if len(assets) > 0 && slices.Contains(dbParams.IncludeRelations, attachmentDao.Table()) {
		assetIds := []string{}
		for _, asset := range assets {
			assetIds = append(assetIds, asset.ID)
		}

		// Reduce the order by clause to only include columns specific to the attachments table
		reducedOrderBy := GenericProcessOrderBy(origOrderBy, attachmentDao.columns(), true)

		dbParams = &database.DatabaseParams{
			OrderBy: reducedOrderBy,
			Where:   squirrel.Eq{"asset_id": assetIds},
		}

		// Get the attachments
		attachments, err := attachmentDao.List(dbParams, tx)
		if err != nil {
			return nil, err
		}

		// Store in a map for easy lookup
		attachmentsMap := map[string][]*models.Attachment{}
		for _, a := range attachments {
			attachmentsMap[a.AssetID] = append(attachmentsMap[a.AssetID], a)
		}

		// Add attachments to its asset
		for _, a := range assets {
			a.Attachments = attachmentsMap[a.ID]
		}
	}

	return assets, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update updates select columns of an asset
//
//   - title
//   - prefix
//   - chapter
//   - type
//   - path
func (dao *AssetDao) Update(asset *models.Asset, tx *database.Tx) error {
	if asset.ID == "" {
		return ErrEmptyId
	}

	asset.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Update(dao.Table()).
		Set("title", NilStr(asset.Title)).
		Set("prefix", asset.Prefix).
		Set("chapter", NilStr(asset.Chapter)).
		Set("type", NilStr(asset.Type.String())).
		Set("path", NilStr(asset.Path)).
		Set("updated_at", FormatTime(asset.UpdatedAt)).
		Where("id = ?", asset.ID).
		ToSql()

	execFn := dao.db.Exec
	if tx != nil {
		execFn = tx.Exec
	}

	_, err := execFn(query, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes assets based upon the where clause
func (dao *AssetDao) Delete(dbParams *database.DatabaseParams, tx *database.Tx) error {
	return GenericDelete(dao, dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// countSelect returns the default count select builder
func (dao *AssetDao) countSelect() squirrel.SelectBuilder {
	apDao := NewAssetProgressDao(dao.db)

	return squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.Table()).
		LeftJoin(apDao.Table() + " ON " + dao.Table() + ".id = " + apDao.Table() + ".asset_id").
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
//
// It performs 1 left join
//   - assets progress table to get `video_pos`, `completed` and `completed_at`
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *AssetDao) baseSelect() squirrel.SelectBuilder {
	return dao.countSelect()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// columns returns the columns to select
func (dao *AssetDao) columns() []string {
	apDao := NewAssetProgressDao(dao.db)

	return []string{
		dao.Table() + ".*",
		apDao.Table() + ".video_pos",
		apDao.Table() + ".completed",
		apDao.Table() + ".completed_at",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// data generates a map of key/values for an asset
func (dao *AssetDao) data(a *models.Asset) map[string]any {
	return map[string]any{
		"id":         a.ID,
		"course_id":  NilStr(a.CourseID),
		"title":      NilStr(a.Title),
		"prefix":     a.Prefix,
		"chapter":    NilStr(a.Chapter),
		"type":       NilStr(a.Type.String()),
		"path":       NilStr(a.Path),
		"hash":       NilStr(a.Hash),
		"created_at": FormatTime(a.CreatedAt),
		"updated_at": FormatTime(a.UpdatedAt),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans an asset row
func (dao *AssetDao) scanRow(scannable Scannable) (*models.Asset, error) {
	var a models.Asset

	// Nullable fields
	var chapter sql.NullString
	var videoPos sql.NullInt16
	var completed sql.NullBool
	var createdAt string
	var updatedAt string
	var completedAt sql.NullString

	err := scannable.Scan(
		&a.ID,
		&a.CourseID,
		&a.Title,
		&a.Prefix,
		&chapter,
		&a.Type,
		&a.Path,
		&a.Hash,
		&createdAt,
		&updatedAt,

		// Asset progress
		&videoPos,
		&completed,
		&completedAt,
	)

	if err != nil {
		return nil, err
	}

	if chapter.Valid {
		a.Chapter = chapter.String
	}

	if a.CreatedAt, err = ParseTime(createdAt); err != nil {
		return nil, err
	}

	if a.UpdatedAt, err = ParseTime(updatedAt); err != nil {
		return nil, err
	}

	if completedAt.Valid {
		if a.CompletedAt, err = ParseTime(completedAt.String); err != nil {
			return nil, err
		}
	} else {
		a.CompletedAt = time.Time{}
	}

	a.VideoPos = int(videoPos.Int16)
	a.Completed = completed.Bool

	return &a, nil
}
