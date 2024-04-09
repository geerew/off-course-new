package daos

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetProgressDao is the data access object for assets progress
type AssetProgressDao struct {
	db    database.Database
	table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewAssetProgressDao returns a new AssetProgressDao
func NewAssetProgressDao(db database.Database) *AssetProgressDao {
	return &AssetProgressDao{
		db:    db,
		table: TableAssetsProgress(),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TableAssetProgress returns the name of the assets progress table
func TableAssetsProgress() string {
	return "assets_progress"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create inserts a new asset progress
//
// The function must be run within a transaction as it also updates the course progress. If `tx` is
// nil, the function will create a new transaction
func (dao *AssetProgressDao) Create(ap *models.AssetProgress, tx *sql.Tx) error {
	if tx == nil {
		return dao.db.RunInTransaction(func(tx *sql.Tx) error {
			return dao.create(ap, tx)
		})
	} else {
		return dao.create(ap, tx)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get selects an asset progress with the given asset ID
func (dao *AssetProgressDao) Get(assetId string, tx *sql.Tx) (*models.AssetProgress, error) {
	generic := NewGenericDao(dao.db, dao.table)

	dbParams := &database.DatabaseParams{
		Columns: dao.selectColumns(),
		Where:   squirrel.Eq{generic.table + ".asset_id": assetId},
	}

	row, err := generic.Get(dao.baseSelect(), dbParams, tx)
	if err != nil {
		return nil, err
	}

	cp, err := dao.scanRow(row)
	if err != nil {
		return nil, err
	}

	return cp, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update updates the `video_pos` (for video assets) and `completed`
//
// When `completed` is true, `completed_at` is set to the current time. When the `completed`
// is false, `completed_at` is set to null
//
// Note: Only the `video_pos` and `completed` may be updated
//
// The function must be run within a transaction as it also updates the course progress. If `tx` is
// nil, the function will create a new transaction
func (dao *AssetProgressDao) Update(ap *models.AssetProgress, tx *sql.Tx) error {
	if tx == nil {
		return dao.db.RunInTransaction(func(tx *sql.Tx) error {
			return dao.update(ap, tx)
		})
	} else {
		return dao.update(ap, tx)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (dao *AssetProgressDao) create(ap *models.AssetProgress, tx *sql.Tx) error {
	if ap.ID == "" {
		ap.RefreshId()
	}

	ap.RefreshCreatedAt()
	ap.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.table).
		SetMap(dao.data(ap)).
		ToSql()

	_, err := tx.Exec(query, args...)

	if err != nil {
		return err
	}

	// Refresh course progress
	cpDao := NewCourseProgressDao(dao.db)
	return cpDao.Refresh(ap.CourseID, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (dao *AssetProgressDao) update(ap *models.AssetProgress, tx *sql.Tx) error {
	if ap.AssetID == "" {
		return ErrEmptyId
	}

	if tx == nil {
		return ErrNilTransaction
	}

	// Normalize video position
	if ap.VideoPos < 0 {
		ap.VideoPos = 0
	}

	// Get the asset
	assetDao := NewAssetDao(dao.db)
	asset, err := assetDao.Get(ap.AssetID, nil, tx)
	if err != nil {
		return err
	}

	// Return when nothing has changed
	if asset.VideoPos == ap.VideoPos && asset.Completed == ap.Completed {
		return nil
	}

	// Set an id (if empty)
	if ap.ID == "" {
		ap.RefreshId()
	}

	// Set course id (if empty)
	if ap.CourseID == "" {
		ap.CourseID = asset.CourseID
	}

	ap.RefreshCreatedAt()
	ap.RefreshUpdatedAt()

	if ap.Completed {
		if !asset.CompletedAt.IsZero() {
			ap.CompletedAt = asset.CompletedAt
		} else {
			ap.CompletedAt = types.NowDateTime()
		}
	} else {
		ap.CompletedAt = types.DateTime{}
	}

	// Update (or create if it doesn't exist)
	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.table).
		SetMap(dao.data(ap)).
		Suffix(
			"ON CONFLICT (asset_id) DO UPDATE SET video_pos = ?, completed = ?, completed_at = ?, updated_at = ?",
			ap.VideoPos, ap.Completed, NilStr(ap.CompletedAt.String()), ap.UpdatedAt,
		).
		ToSql()

	_, err = tx.Exec(query, args...)
	if err != nil {
		return err
	}

	// Refresh course progress
	cpDao := NewCourseProgressDao(dao.db)
	return cpDao.Refresh(ap.CourseID, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *AssetProgressDao) baseSelect() squirrel.SelectBuilder {
	return squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.table).
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// selectColumns returns the columns to select
func (dao *AssetProgressDao) selectColumns() []string {
	return []string{
		dao.table + ".*",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// data generates a map of key/values for thane asset progress
func (dao *AssetProgressDao) data(ap *models.AssetProgress) map[string]any {
	return map[string]any{
		"id":           ap.ID,
		"asset_id":     NilStr(ap.AssetID),
		"course_id":    NilStr(ap.CourseID),
		"video_pos":    ap.VideoPos,
		"completed":    ap.Completed,
		"completed_at": ap.CompletedAt,
		"created_at":   ap.CreatedAt,
		"updated_at":   ap.UpdatedAt,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans an assets progress row
func (dao *AssetProgressDao) scanRow(scannable Scannable) (*models.AssetProgress, error) {
	var ap models.AssetProgress

	err := scannable.Scan(
		&ap.ID,
		&ap.AssetID,
		&ap.CourseID,
		&ap.VideoPos,
		&ap.Completed,
		&ap.CompletedAt,
		&ap.CreatedAt,
		&ap.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &ap, nil
}
