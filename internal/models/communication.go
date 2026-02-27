package models

import (
	"context"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/communication"
)

type CommunicationModel struct {
	Db *ent.Client
}

// Create persists a communication log entry after a mass email is sent.
func (m *CommunicationModel) Create(
	ctx context.Context,
	subject, bodyHTML, filter, filterLabel string,
	recipientCount, sentCount, failCount int,
	churchID, senderID int,
) (*ent.Communication, error) {
	b := m.Db.Communication.Create().
		SetSubject(subject).
		SetBodyHTML(bodyHTML).
		SetRecipientFilter(filter).
		SetRecipientFilterLabel(filterLabel).
		SetRecipientCount(recipientCount).
		SetSentCount(sentCount).
		SetFailCount(failCount).
		SetChurchID(churchID)
	if senderID > 0 {
		b = b.SetSenderID(senderID)
	}
	c, err := b.Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return c, nil
}

// ListByChurch returns all communications for a church, newest first.
func (m *CommunicationModel) ListByChurch(ctx context.Context, churchID int) ([]*ent.Communication, error) {
	q := m.Db.Communication.Query().
		WithSender(func(uq *ent.UserQuery) { uq.WithContact() }).
		Order(ent.Desc(communication.FieldSentAt))
	if churchID > 0 {
		q = q.Where(communication.HasChurchWith(church.IDEQ(churchID)))
	}
	return q.All(ctx)
}

// GetByID returns a single communication with sender edge loaded.
func (m *CommunicationModel) GetByID(ctx context.Context, id int) (*ent.Communication, error) {
	c, err := m.Db.Communication.Query().
		Where(communication.IDEQ(id)).
		WithSender(func(uq *ent.UserQuery) { uq.WithContact() }).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return c, nil
}
