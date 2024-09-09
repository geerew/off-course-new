package daos

import (
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
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
		table: "assets_progress",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table returns the table name
func (dao *AssetProgressDao) Table() string {
	return dao.table
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create creates an asset progress, then refreshes the course progress
//
// A new transaction is created if `tx` is nil
func (dao *AssetProgressDao) Create(ap *models.AssetProgress, tx *database.Tx) error {
	createFn := func(tx *database.Tx) error {
		if ap.ID == "" {
			ap.RefreshId()
		}

		ap.RefreshCreatedAt()
		ap.RefreshUpdatedAt()

		query, args, _ := squirrel.
			StatementBuilder.
			Insert(dao.Table()).
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

	if tx == nil {
		return dao.db.RunInTransaction(func(tx *database.Tx) error {
			return createFn(tx)
		})
	} else {
		return createFn(tx)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get gets an asset progress with the given ID
func (dao *AssetProgressDao) Get(assetId string, tx *database.Tx) (*models.AssetProgress, error) {
	generic := NewGenericDao(dao.db, dao)

	dbParams := &database.DatabaseParams{
		Columns: dao.columns(),
		Where:   squirrel.Eq{dao.Table() + ".asset_id": assetId},
	}

	row, err := generic.Get(dbParams, tx)
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

// Update updates select columns of an asset progress, then refresh course progress
//
//   - `video_pos` (for video assets)
//   - `completed` (for all assets)
//
// When `completed` is true, `completed_at` is set to the current time, else it will be null
//
// A new transaction is created if `tx` is nil
func (dao *AssetProgressDao) Update(ap *models.AssetProgress, tx *database.Tx) error {
	updateFn := func(tx *database.Tx) error {
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
				ap.CompletedAt = time.Now()
			}
		} else {
			ap.CompletedAt = time.Time{}
		}

		// Update (or create if it doesn't exist)
		query, args, _ := squirrel.
			StatementBuilder.
			Insert(dao.Table()).
			SetMap(dao.data(ap)).
			Suffix(
				"ON CONFLICT (asset_id) DO UPDATE SET video_pos = ?, completed = ?, completed_at = ?, updated_at = ?",
				ap.VideoPos, ap.Completed, FormatTime(ap.CompletedAt), FormatTime(ap.UpdatedAt),
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

	if tx == nil {
		return dao.db.RunInTransaction(func(tx *database.Tx) error {
			return updateFn(tx)
		})
	} else {
		return updateFn(tx)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// countSelect returns the default count select builder
func (dao *AssetProgressDao) countSelect() squirrel.SelectBuilder {
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
func (dao *AssetProgressDao) baseSelect() squirrel.SelectBuilder {
	return dao.countSelect()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// columns returns the columns to select
func (dao *AssetProgressDao) columns() []string {
	return []string{
		dao.Table() + ".*",
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
		"completed_at": FormatTime(ap.CompletedAt),
		"created_at":   FormatTime(ap.CreatedAt),
		"updated_at":   FormatTime(ap.UpdatedAt),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans an assets progress row
func (dao *AssetProgressDao) scanRow(scannable Scannable) (*models.AssetProgress, error) {
	var ap models.AssetProgress

	var createdAt string
	var updatedAt string
	var completedAt string

	err := scannable.Scan(
		&ap.ID,
		&ap.AssetID,
		&ap.CourseID,
		&ap.VideoPos,
		&ap.Completed,
		&completedAt,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	if ap.CompletedAt, err = ParseTime(completedAt); err != nil {
		return nil, err
	}

	if ap.CreatedAt, err = ParseTime(createdAt); err != nil {
		return nil, err
	}

	if ap.UpdatedAt, err = ParseTime(updatedAt); err != nil {
		return nil, err
	}

	return &ap, nil
}
