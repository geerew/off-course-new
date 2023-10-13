package models

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type Scan struct {
	BaseModel
	CourseID string           `bun:",unique,notnull"`
	Status   types.ScanStatus `bun:",notnull"`

	// Belongs to
	Course *Course `bun:"rel:belongs-to,join:course_id=id"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CountScans returns the number of scans
func CountScans(ctx context.Context, db database.Database, params *database.DatabaseParams) (int, error) {
	q := db.DB().NewSelect().Model((*Scan)(nil))

	if params != nil && params.Where != nil {
		q = selectWhere(q, params, "scan")
	}

	return q.Count(ctx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetScans returns a slice of scans
func GetScans(ctx context.Context, db database.Database, params *database.DatabaseParams) ([]*Scan, error) {
	var scans []*Scan

	q := db.DB().NewSelect().Model(&scans)

	if params != nil {
		// Pagination
		if params.Pagination != nil {
			if count, err := CountScans(ctx, db, params); err != nil {
				return nil, err
			} else {
				params.Pagination.SetCount(count)
			}

			q = q.Offset(params.Pagination.Offset()).Limit(params.Pagination.Limit())
		}

		if params.Relation != nil {
			q = selectRelation(q, params)
		}

		// Order by
		if len(params.OrderBy) > 0 {
			q = q.Order(params.OrderBy...)
		}

		// Where
		if params.Where != nil {
			if params.Where != nil {
				q = selectWhere(q, params, "scan")
			}
		}
	}

	err := q.Scan(ctx)

	return scans, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetScan returns a scan based upon the where clause in the database params
func GetScan(ctx context.Context, db database.Database, params *database.DatabaseParams) (*Scan, error) {
	if params == nil || params.Where == nil {
		return nil, errors.New("where clause required")
	}

	scan := &Scan{}

	q := db.DB().NewSelect().Model(scan)

	// Where
	if params.Where != nil {
		q = selectWhere(q, params, "scan")
	}

	// Relations
	if params.Relation != nil {
		q = selectRelation(q, params)
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return scan, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateScan inserts a new scan with a status of waiting
func CreateScan(ctx context.Context, db database.Database, scan *Scan) error {
	scan.RefreshId()
	scan.RefreshCreatedAt()
	scan.RefreshUpdatedAt()
	scan.Status = types.NewScanStatus(types.ScanStatusWaiting)

	_, err := db.DB().NewInsert().Model(scan).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateScanStatus updates the scan status
func UpdateScanStatus(ctx context.Context, db database.Database, scan *Scan, newStatus types.ScanStatusType) error {
	// Do nothing when the status is the same
	ss := types.NewScanStatus(newStatus)
	if scan.Status == ss {
		return nil
	}

	// Require an ID
	if scan.ID == "" {
		return errors.New("scan ID cannot be empty")
	}

	// Set a new timestamp
	ts := types.NowDateTime()

	// Update the status
	if res, err := db.DB().NewUpdate().Model(scan).
		Set("status = ?", ss).
		Set("updated_at = ?", ts).
		WherePK().Exec(ctx); err != nil {
		return err
	} else {
		count, _ := res.RowsAffected()
		if count == 0 {
			return nil
		}
	}

	// Update the original scan struct
	scan.Status = ss
	scan.UpdatedAt = ts

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteScan deletes a scan with the given ID
func DeleteScan(ctx context.Context, db database.Database, id string) (int, error) {
	scan := &Scan{}
	scan.SetId(id)

	if res, err := db.DB().NewDelete().Model(scan).WherePK().Exec(ctx); err != nil {
		return 0, err
	} else {
		count, _ := res.RowsAffected()
		return int(count), err
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NextScan returns the next scan to be processed whose status is `waiting“
func NextScan(ctx context.Context, db database.Database) (*Scan, error) {
	var scan Scan

	err := db.DB().NewSelect().
		Model(&scan).
		Relation("Course").
		Where("scan.status = ?", types.ScanStatusWaiting).
		Order("scan.created_at ASC").
		Limit(1).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &scan, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTestScans creates a scan for each course in the slice. If a db is provided, the scans will
// be inserted into the db
//
// THIS IS FOR TESTING PURPOSES
func NewTestScans(t *testing.T, db database.Database, courses []*Course) []*Scan {
	scans := []*Scan{}
	for i := 0; i < len(courses); i++ {
		s := &Scan{
			CourseID: courses[i].ID,
			Status:   types.NewScanStatus(types.ScanStatusWaiting),
		}

		s.RefreshId()
		s.RefreshCreatedAt()
		s.RefreshUpdatedAt()

		if db != nil {
			_, err := db.DB().NewInsert().Model(s).Exec(context.Background())
			require.Nil(t, err)
		}

		scans = append(scans, s)
		time.Sleep(1 * time.Millisecond)
	}

	return scans
}
