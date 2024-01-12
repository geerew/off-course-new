package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Defines a model for the table `attachments`
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Attachments defines the model for an attachment
//
// As this changes, update `scanAttachmentRow()`
type Attachment struct {
	BaseModel

	CourseID string
	AssetID  string
	Title    string
	Path     string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TableAttachments returns the table name for the attachments table
func TableAttachments() string {
	return "attachments"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CountAttachments counts the number of attachments
func CountAttachments(db database.Database, params *database.DatabaseParams) (int, error) {
	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select("COUNT(*)").
		From(TableAttachments())

	// Add where clauses if necessary
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

// GetAttachments selects attachments
func GetAttachments(db database.Database, params *database.DatabaseParams) ([]*Attachment, error) {
	var attachments []*Attachment

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select(TableAttachments() + ".*").
		From(TableAttachments())

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
			if builder, err = paginate(db, params, builder, CountAttachments); err != nil {
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
		a, err := scanAttachmentRow(rows)
		if err != nil {
			return nil, err
		}

		attachments = append(attachments, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return attachments, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAttachment selects an attachment for the given ID
func GetAttachment(db database.Database, id string) (*Attachment, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}
	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select(TableAttachments() + ".*").
		From(TableAttachments()).
		Where(sq.Eq{TableAttachments() + ".id": id})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	row := db.QueryRow(query, args...)
	if row.Err() != nil {
		return nil, row.Err()
	}

	attachment, err := scanAttachmentRow(row)
	if err != nil {
		return nil, err
	}

	return attachment, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAttachment inserts a new attachment
func CreateAttachment(db database.Database, a *Attachment) error {
	a.RefreshId()
	a.RefreshCreatedAt()
	a.RefreshUpdatedAt()

	builder := sq.StatementBuilder.
		Insert(TableAttachments()).
		Columns("id", "course_id", "asset_id", "title", "path", "created_at", "updated_at").
		Values(a.ID, NilStr(a.CourseID), NilStr(a.AssetID), NilStr(a.Title), NilStr(a.Path), a.CreatedAt, a.UpdatedAt)

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(query, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteAttachment deletes an attachment with the given ID
func DeleteAttachment(db database.Database, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Delete(TableAttachments()).
		Where(sq.Eq{"id": id})

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(query, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanAttachmentRow scans an attachment row
func scanAttachmentRow(scannable Scannable) (*Attachment, error) {
	var a Attachment

	err := scannable.Scan(
		&a.ID,
		&a.CourseID,
		&a.AssetID,
		&a.Title,
		&a.Path,
		&a.CreatedAt,
		&a.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &a, nil
}
