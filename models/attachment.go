package models

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/security"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type Attachment struct {
	BaseModel
	CourseID string
	AssetID  string
	Title    string `bun:",notnull,default:null"`
	Path     string `bun:",unique,notnull,default:null"`

	// Belongs to
	Course *Course `bun:"rel:belongs-to,join:course_id=id"`
	Asset  *Asset  `bun:"rel:belongs-to,join:asset_id=id"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CountAttachments returns the number of attachments
func CountAttachments(ctx context.Context, db database.Database, params *database.DatabaseParams) (int, error) {
	q := db.DB().NewSelect().Model((*Attachment)(nil))

	if params != nil && params.Where != nil {
		q = selectWhere(q, params, "attachment")
	}

	return q.Count(ctx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAttachments returns a slice of attachments
func GetAttachments(ctx context.Context, db database.Database, params *database.DatabaseParams) ([]*Attachment, error) {
	var attachments []*Attachment

	q := db.DB().NewSelect().Model(&attachments)

	if params != nil {
		// Pagination
		if params.Pagination != nil {
			if count, err := CountAttachments(ctx, db, params); err != nil {
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
			selectOrderBy(q, params, "attachment")
		}

		// Where
		if params.Where != nil {
			if params.Where != nil {
				q = selectWhere(q, params, "attachment")
			}
		}
	}

	err := q.Scan(ctx)

	return attachments, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAttachment returns an attachment based upon the where clause in the database params
func GetAttachment(ctx context.Context, db database.Database, params *database.DatabaseParams) (*Attachment, error) {
	if params == nil || params.Where == nil {
		return nil, errors.New("where clause required")
	}

	attachment := &Attachment{}

	q := db.DB().NewSelect().Model(attachment)

	// Where
	if params.Where != nil {
		q = selectWhere(q, params, "attachment")
	}

	// Relations
	if params.Relation != nil {
		q = selectRelation(q, params)
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return attachment, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAttachmentById returns an attachment for the given ID
func GetAttachmentById(ctx context.Context, db database.Database, params *database.DatabaseParams, id string) (*Attachment, error) {
	attachment := &Attachment{}

	q := db.DB().NewSelect().Model(attachment).Where("attachment.id = ?", id)

	if params != nil && params.Relation != nil {
		q = selectRelation(q, params)
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return attachment, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAttachmentById returns a slice of attachments for the given asset ID
func GetAttachmentsByAssetId(ctx context.Context, db database.Database, params *database.DatabaseParams, id string) ([]*Attachment, error) {
	var attachments []*Attachment

	q := db.DB().NewSelect().Model(&attachments).Where("attachment.asset_id = ?", id)

	if params != nil {
		// Pagination
		if params.Pagination != nil {
			// Set the where to the asset ID
			params.Where = []database.Where{{Column: "attachment.asset_id", Value: id}}

			if count, err := CountAttachments(ctx, db, params); err != nil {
				return nil, err
			} else {
				params.Pagination.SetCount(count)
			}

			q = q.Offset(params.Pagination.Offset()).Limit(params.Pagination.Limit())
		}

		// Order by
		if len(params.OrderBy) > 0 {
			selectOrderBy(q, params, "attachment")
		}

		// Relation
		if params.Relation != nil {
			q = selectRelation(q, params)
		}
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return attachments, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAttachmentsByCourseId returns a slice of attachments for the given course ID
func GetAttachmentsByCourseId(ctx context.Context, db database.Database, params *database.DatabaseParams, id string) ([]*Attachment, error) {
	var attachments []*Attachment

	q := db.DB().NewSelect().Model(&attachments).Where("attachment.course_id = ?", id)

	if params != nil {
		// Pagination
		if params.Pagination != nil {
			// Set the where to the course ID
			params.Where = []database.Where{{Column: "attachment.course_id", Value: id}}

			if count, err := CountAttachments(ctx, db, params); err != nil {
				return nil, err
			} else {
				params.Pagination.SetCount(count)
			}

			q = q.Offset(params.Pagination.Offset()).Limit(params.Pagination.Limit())
		}

		// Order by
		if len(params.OrderBy) > 0 {
			selectOrderBy(q, params, "attachment")
		}

		// Relation
		if params.Relation != nil {
			q = selectRelation(q, params)
		}
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return attachments, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAttachment inserts a new attachment
func CreateAttachment(ctx context.Context, db database.Database, attachment *Attachment) error {
	attachment.RefreshId()
	attachment.RefreshCreatedAt()
	attachment.RefreshUpdatedAt()

	_, err := db.DB().NewInsert().Model(attachment).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteAttachment deletes an attachment with the given ID
func DeleteAttachment(ctx context.Context, db database.Database, id string) (int, error) {
	attachment := &Attachment{}
	attachment.SetId(id)

	if res, err := db.DB().NewDelete().Model(attachment).WherePK().Exec(ctx); err != nil {
		return 0, err
	} else {
		count, _ := res.RowsAffected()
		return int(count), err
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTestAttachments creates n number of attachments for each asset in the slice. If a db is
// provided, the attachments will be inserted into the db
//
// THIS IS FOR TESTING PURPOSES
func NewTestAttachments(t *testing.T, db database.Database, assets []*Asset, attachmentsPerAsset int) []*Attachment {
	attachments := []*Attachment{}
	for i := 0; i < len(assets); i++ {
		for j := 0; j < attachmentsPerAsset; j++ {
			title := fmt.Sprintf("%s.txt", security.PseudorandomString(8))

			a := &Attachment{
				CourseID: assets[i].CourseID,
				AssetID:  assets[i].ID,
				Title:    title,
				Path:     fmt.Sprintf("%s/%d %s", filepath.Dir(assets[i].Path), assets[i].Prefix, title),
			}

			a.RefreshId()
			a.RefreshCreatedAt()
			a.RefreshUpdatedAt()

			if db != nil {
				_, err := db.DB().NewInsert().Model(a).Exec(context.Background())
				require.Nil(t, err)
			}

			attachments = append(attachments, a)
			time.Sleep(1 * time.Millisecond)
		}
	}

	return attachments
}
