package models

import (
	"context"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/document"
)

type DocumentModel struct {
	Db *ent.Client
}

// Create adds a new document record.
func (m *DocumentModel) Create(ctx context.Context, dto *DocumentDto, fileURL, fileName, mimeType string, fileSize int64, churchID int) (*ent.Document, error) {
	d, err := m.Db.Document.Create().
		SetTitle(dto.Title).
		SetNillableDescription(nullStr(dto.Description)).
		SetCategory(document.Category(dto.Category)).
		SetFileURL(fileURL).
		SetFileName(fileName).
		SetFileSize(fileSize).
		SetNillableMimeType(nullStr(mimeType)).
		SetIsPublic(dto.IsPublic).
		SetChurchID(churchID).
		Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return d, nil
}

// GetByID returns a single document.
func (m *DocumentModel) GetByID(ctx context.Context, id int) (*ent.Document, error) {
	d, err := m.Db.Document.Query().
		Where(document.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return d, nil
}

// ListByChurch returns all documents for a church, newest first.
func (m *DocumentModel) ListByChurch(ctx context.Context, churchID int) ([]*ent.Document, error) {
	q := m.Db.Document.Query()
	if churchID > 0 {
		q = q.Where(document.HasChurchWith(church.IDEQ(churchID)))
	}
	return q.Order(ent.Desc(document.FieldCreatedAt)).All(ctx)
}

// Delete removes a document record (caller is responsible for removing the file from MinIO if needed).
func (m *DocumentModel) Delete(ctx context.Context, id int) error {
	return m.Db.Document.DeleteOneID(id).Exec(ctx)
}
