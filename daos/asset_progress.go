package daos

import (
	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetProgressDao is the data access object for assets progress
type AssetProgressDao struct {
	BaseDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewAssetProgressDao returns a new AssetProgressDao
func NewAssetProgressDao(db database.Database) *AssetProgressDao {
	return &AssetProgressDao{
		BaseDao: BaseDao{
			db:    db,
			table: "assets_progress",
		},
	}
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
			SetMap(toDBMapOrPanic(ap)).
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
	selectColumns, _ := tableColumnsOrPanic(models.AssetProgress{}, dao.Table())

	dbParams := &database.DatabaseParams{
		Columns: selectColumns,
		Where:   squirrel.Eq{dao.Table() + ".asset_id": assetId},
	}

	return genericGet(dao, dbParams, dao.scanRow, tx)
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
				ap.CompletedAt = types.NowDateTime()
			}
		} else {
			ap.CompletedAt = types.DateTime{}
		}

		// Update (or create if it doesn't exist)
		query, args, _ := squirrel.
			StatementBuilder.
			Insert(dao.Table()).
			SetMap(toDBMapOrPanic(ap)).
			Suffix(
				"ON CONFLICT (asset_id) DO UPDATE SET video_pos = ?, completed = ?, completed_at = ?, updated_at = ?",
				ap.VideoPos, ap.Completed, ap.CompletedAt, ap.UpdatedAt,
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

// scanRow scans an assets progress row
func (dao *AssetProgressDao) scanRow(scannable Scannable) (*models.AssetProgress, error) {
	var ap models.AssetProgress

	err := scannable.Scan(
		&ap.ID,
		&ap.CreatedAt,
		&ap.UpdatedAt,
		&ap.AssetID,
		&ap.CourseID,
		&ap.VideoPos,
		&ap.Completed,
		&ap.CompletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &ap, nil
}
