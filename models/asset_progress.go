package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Defines a model for the table `assets_progress`
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetProgress defines the model for a course progress
//
// As this changes, update `scanAssetProgressRow()`
type AssetProgress struct {
	BaseModel

	AssetID     string
	CourseID    string
	VideoPos    int
	Completed   bool
	CompletedAt types.DateTime

	// Course progress information for convenience
	CourseStarted bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TableAssetsProgress returns the table name for the courses progress table
func TableAssetsProgress() string {
	return "assets_progress"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CountAssetsProgress counts the number of assets progress
func CountAssetsProgress(db database.Database, params *database.DatabaseParams) (int, error) {
	builder := assetsProgressBaseSelect().Columns("COUNT(DISTINCT " + TableAssetsProgress() + ".id)")

	if params != nil && params.Where != "" {
		builder = builder.Where(params.Where)
	}

	// Build the query
	query, args, err := builder.ToSql()
	if err != nil {
		return -1, err
	}

	// Execute the query
	var count int
	err = db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAssetsProgress selects courses progress
func GetAssetsProgress(db database.Database, params *database.DatabaseParams) ([]*AssetProgress, error) {
	var aps []*AssetProgress

	builder := assetsProgressBaseSelect().
		Columns(TableAssetsProgress()+".*", TableCoursesProgress()+".started as course_started")

	if params != nil {
		// ORDER BY
		if params != nil && len(params.OrderBy) > 0 {
			builder = builder.OrderBy(params.OrderBy...)
		}

		// WHERE
		if params.Where != "" {
			builder = builder.Where(params.Where)
		}

		// PAGINATION
		if params.Pagination != nil {
			var err error
			if builder, err = paginate(db, params, builder, CountAssetsProgress); err != nil {
				return nil, err
			}
		}
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		ap, err := scanAssetProgressRow(rows)
		if err != nil {
			return nil, err
		}

		aps = append(aps, ap)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return aps, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAssetProgress selects an asset progress for the given asset ID
func GetAssetProgress(db database.Database, assetId string) (*AssetProgress, error) {
	if assetId == "" {
		return nil, errors.New("id cannot be empty")
	}

	builder := assetsProgressBaseSelect().
		Columns(TableAssetsProgress()+".*", TableCoursesProgress()+".started as course_started").
		Where(sq.Eq{TableAssetsProgress() + ".asset_id": assetId})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	row := db.QueryRow(query, args...)
	if row.Err() != nil {
		return nil, row.Err()
	}

	ap, err := scanAssetProgressRow(row)
	if err != nil {
		return nil, err
	}

	return ap, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAssetProgress inserts a new asset progress
func CreateAssetProgress(db database.Database, ap *AssetProgress) error {
	ap.RefreshId()
	ap.RefreshCreatedAt()
	ap.RefreshUpdatedAt()

	builder := sq.StatementBuilder.
		Insert(TableAssetsProgress()).
		Columns("id", "asset_id", "course_id", "video_pos", "completed", "completed_at", "created_at", "updated_at").
		Values(ap.ID, NilStr(ap.AssetID), NilStr(ap.CourseID), ap.VideoPos, ap.Completed, ap.CompletedAt, ap.CreatedAt, ap.UpdatedAt)

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(query, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateAssetProgressVideoPos updates `video_pos`. When a video progress entry does not exist, it
// will be created. When an entry does exist, it will be updated.
func UpdateAssetProgressVideoPos(db database.Database, assetId string, position int) (*AssetProgress, error) {
	if assetId == "" {
		return nil, errors.New("id cannot be empty")
	}

	ap, err := GetAssetProgress(db, assetId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Keep the position >= 0
	if position < 0 {
		position = 0
	}

	// --------------------------------
	// Create if it does not exist
	// --------------------------------
	if err == sql.ErrNoRows {
		// This will only happen once for each asset. Subsequent calls will update the existing
		// row
		asset, err := GetAsset(db, assetId)
		if err != nil {
			return nil, err
		}

		ap = &AssetProgress{
			AssetID:  assetId,
			CourseID: asset.CourseID,
			VideoPos: position,
		}

		err = CreateAssetProgress(db, ap)
		if err != nil {
			return nil, err
		}

		// Mark the course as started
		if !ap.CourseStarted {
			_, err := UpdateCourseProgressStarted(db, ap.CourseID, true)
			return ap, err
		}

		_, err = UpdateCourseProgressUpdatedAt(db, ap.CourseID)
		return ap, err
	}

	// --------------------------------
	// Update existing
	// --------------------------------

	// Nothing to do
	if ap.VideoPos == position {
		return ap, nil
	}

	updatedAt := types.NowDateTime()

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Update(TableAssetsProgress()).
		Set("video_pos", position).
		Set("updated_at", updatedAt).
		Where("id = ?", ap.ID)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	ap.VideoPos = position
	ap.UpdatedAt = updatedAt

	// Mark the course as started. This will also update the updated_at timestamp for the course
	// progress
	if !ap.CourseStarted {
		_, err := UpdateCourseProgressStarted(db, ap.CourseID, true)
		return ap, err
	}

	// If the course is already started, just update the updated_at timestamp for the course
	// progress
	_, err = UpdateCourseProgressUpdatedAt(db, ap.CourseID)
	return ap, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateAssetProgressCompleted updates `completed` and `completed_at`. When a video progress entry
// does not exist, it will be created. When an entry does exist, it will be updated.
//
// When an `completed` is true, `completed_at` is set to the current time. When `completed` is
// false, `completed_at` is set to null
func UpdateAssetProgressCompleted(db database.Database, assetId string, completed bool) (*AssetProgress, error) {
	if assetId == "" {
		return nil, errors.New("id cannot be empty")
	}

	ap, err := GetAssetProgress(db, assetId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// --------------------------------
	// Create if it does not exist
	// --------------------------------
	if err == sql.ErrNoRows {
		// This will only happen once for each asset as subsequent calls will update the existing
		// row
		asset, err := GetAsset(db, assetId)
		if err != nil {
			return nil, err
		}

		ap = &AssetProgress{
			AssetID:   assetId,
			CourseID:  asset.CourseID,
			Completed: completed,
		}

		if completed {
			ap.CompletedAt = types.NowDateTime()
		}

		err = CreateAssetProgress(db, ap)
		if err != nil {
			return nil, err
		}

		// When this asset is completed, update the course progress percent
		if completed {
			_, err = UpdateCourseProgressPercent(db, ap.CourseID)
			return ap, err
		}

		return ap, nil
	}

	if ap.Completed == completed {
		return ap, nil
	}

	updatedAt := types.NowDateTime()

	var completedAt types.DateTime
	if completed {
		completedAt = updatedAt
	}

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Update(TableAssetsProgress()).
		Set("completed", completed).
		Set("completed_at", completedAt).
		Set("updated_at", updatedAt).
		Where("id = ?", ap.ID)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	ap.Completed = completed
	ap.CompletedAt = completedAt
	ap.UpdatedAt = updatedAt

	// Update the course progress percent for this course
	_, err = UpdateCourseProgressPercent(db, ap.CourseID)
	return ap, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// assetsProgressBaseSelect returns a select builder for the assets_progress table. It does not
// include any columns by default and as such, you must specify the columns with `.Columns(...)`
func assetsProgressBaseSelect() sq.SelectBuilder {
	return sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select("").
		From(TableAssetsProgress()).
		LeftJoin(TableCoursesProgress() + " ON " + TableAssetsProgress() + ".course_id = " + TableCoursesProgress() + ".course_id").
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanAssetProgressRow scans an asset progress row
func scanAssetProgressRow(scannable Scannable) (*AssetProgress, error) {
	var ap AssetProgress

	err := scannable.Scan(
		&ap.ID,
		&ap.AssetID,
		&ap.CourseID,
		&ap.VideoPos,
		&ap.Completed,
		&ap.CompletedAt,
		&ap.CreatedAt,
		&ap.UpdatedAt,
		&ap.CourseStarted,
	)

	if err != nil {
		return nil, err
	}

	return &ap, nil
}
