package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/contact"
	"github.com/ntiGideon/ent/group"
	"github.com/ntiGideon/internal/models"
)

// GET /communications — sent-mail log
func (app *application) communicationsList(w http.ResponseWriter, r *http.Request) {
	churchID := app.churchID(r)
	comms, err := app.communicationModel.ListByChurch(r.Context(), churchID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.Data = map[string]any{"communications": comms}
	app.render(w, r, http.StatusOK, "communications.gohtml", data)
}

// GET /communications/compose — compose form
func (app *application) communicationComposeGet(w http.ResponseWriter, r *http.Request) {
	churchID := app.churchID(r)
	groups, err := app.groupModel.ListByChurch(r.Context(), churchID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.Data = map[string]any{
		"groups": groups,
		"form":   models.CommunicationDto{},
	}
	app.render(w, r, http.StatusOK, "communication_compose.gohtml", data)
}

// POST /communications/compose — send and log
func (app *application) communicationComposePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	churchID := app.churchID(r)
	senderID := app.sessionManager.GetInt(ctx, "authenticatedUserID")

	var form models.CommunicationDto
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(form.Subject != "", "subject", "Subject is required")
	form.CheckField(form.Body != "", "body", "Message body is required")
	form.CheckField(form.RecipientFilter != "", "recipient_filter", "Please select a recipient group")

	if !form.Valid() {
		groups, _ := app.groupModel.ListByChurch(ctx, churchID)
		data := app.newTemplateData(r)
		data.Data = map[string]any{"groups": groups, "form": form}
		app.render(w, r, http.StatusUnprocessableEntity, "communication_compose.gohtml", data)
		return
	}

	contacts, filterLabel, err := app.resolveRecipients(ctx, churchID, form.RecipientFilter)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	var (
		mu        sync.Mutex
		sentCount int
		failCount int
	)
	var wg sync.WaitGroup
	for _, c := range contacts {
		if c.Email == "" {
			continue
		}
		wg.Add(1)
		go func(email string) {
			defer wg.Done()
			if sendErr := sendHTMLEmail(email, form.Subject, form.Body); sendErr != nil {
				mu.Lock()
				failCount++
				mu.Unlock()
			} else {
				mu.Lock()
				sentCount++
				mu.Unlock()
			}
		}(c.Email)
	}
	wg.Wait()

	recipientCount := len(contacts)
	_, err = app.communicationModel.Create(
		ctx,
		form.Subject, form.Body,
		form.RecipientFilter, filterLabel,
		recipientCount, sentCount, failCount,
		churchID, senderID,
	)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(ctx, "flash", fmt.Sprintf(
		"Email sent to %d recipient(s) — %d delivered, %d failed.",
		recipientCount, sentCount, failCount,
	))
	http.Redirect(w, r, "/communications", http.StatusSeeOther)
}

// GET /communications/{id} — detail view
func (app *application) communicationDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}
	comm, err := app.communicationModel.GetByID(r.Context(), id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.Data = map[string]any{"communication": comm}
	app.render(w, r, http.StatusOK, "communication_detail.gohtml", data)
}

// resolveRecipients returns contacts matching the filter and a human-readable label.
func (app *application) resolveRecipients(ctx context.Context, churchID int, filter string) ([]*ent.Contact, string, error) {
	q := app.db.Contact.Query().
		Where(contact.HasChurchWith(church.IDEQ(churchID)))

	var label string
	switch {
	case filter == "all":
		label = "All Members"
	case filter == "with_email":
		q = q.Where(contact.EmailNotNil(), contact.EmailNEQ(""))
		label = "Members with Email"
	case strings.HasPrefix(filter, "gender:"):
		g := strings.TrimPrefix(filter, "gender:")
		q = q.Where(
			contact.GenderEQ(contact.Gender(g)),
			contact.EmailNotNil(),
			contact.EmailNEQ(""),
		)
		label = strings.ToUpper(g[:1]) + g[1:] + " Members"
	case strings.HasPrefix(filter, "group:"):
		gidStr := strings.TrimPrefix(filter, "group:")
		gid, err := strconv.Atoi(gidStr)
		if err != nil {
			label = "Group"
		} else {
			grp, _ := app.groupModel.GetByID(ctx, gid)
			if grp != nil {
				label = "Group: " + grp.Name
			} else {
				label = "Group"
			}
			q = q.Where(
				contact.HasGroupsWith(group.IDEQ(gid)),
				contact.EmailNotNil(),
				contact.EmailNEQ(""),
			)
		}
	default:
		label = "All Members"
	}

	contacts, err := q.All(ctx)
	return contacts, label, err
}
