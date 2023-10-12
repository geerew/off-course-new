package models

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/security"
	"github.com/stretchr/testify/require"
)

// import (
// 	"time"

// 	"github.com/geerew/off-course/database"
// 	"github.com/geerew/off-course/utils/security"
// 	"gorm.io/gorm"
// )

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
func CountAttachments(db database.Database, params *database.DatabaseParams, ctx context.Context) (int, error) {
	q := db.DB().NewSelect().Model((*Attachment)(nil))

	if params != nil && params.Where != nil {
		q = selectWhere(q, params)
	}

	return q.Count(ctx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAttachments returns a slice of attachments
func GetAttachments(db database.Database, params *database.DatabaseParams, ctx context.Context) ([]*Attachment, error) {
	var attachments []*Attachment

	q := db.DB().NewSelect().Model(&attachments)

	if params != nil {
		// // Pagination
		// if params.Pagination != nil {
		// 	if count, err := CountScans(db, params); err != nil {
		// 		return nil, err
		// 	} else {
		// 		params.Pagination.SetCount(count)
		// 	}

		// 	q = q.Scopes(params.Pagination.Paginate())
		// }

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
				q = selectWhere(q, params)
			}
		}
	}

	err := q.Scan(ctx)

	return attachments, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAttachmentById returns an attachment for the given ID
func GetAttachmentById(db database.Database, id string, params *database.DatabaseParams, ctx context.Context) (*Attachment, error) {
	attachment := &Attachment{}
	attachment.SetId(id)

	q := db.DB().NewSelect().Model(attachment)

	if params != nil && params.Relation != nil {
		q = selectRelation(q, params)
	}

	if err := q.Where("attachment.id = ?", id).Scan(ctx); err != nil {
		return nil, err
	}

	return attachment, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAttachment inserts a new attachment
func CreateAttachment(db database.Database, attachment *Attachment, ctx context.Context) error {
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
func DeleteAttachment(db database.Database, id string, ctx context.Context) (int, error) {
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
