package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /workers
func (app *application) workersGet(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	churchID := 0
	if u != nil && u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}

	workers, err := app.memberModel.ListWorkersByChurch(r.Context(), churchID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"workers": workers,
	}
	app.render(w, r, http.StatusOK, "workers.gohtml", data)
}

// GET /workers/new
func (app *application) workerInviteGet(w http.ResponseWriter, r *http.Request) {
	churchID := app.churchID(r)
	customRoles, err := app.customRoleModel.ListByChurch(r.Context(), churchID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.Form = models.MemberInviteDto{}
	data.Data = map[string]interface{}{
		"customRoles": customRoles,
	}
	app.render(w, r, http.StatusOK, "invite_member.gohtml", data)
}

// POST /workers/new
func (app *application) workerInvitePost(w http.ResponseWriter, r *http.Request) {
	var dto models.MemberInviteDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Email), "email", "Email is required")
	dto.CheckField(validator.Matches(dto.Email, validator.EmailRX), "email", "Must be a valid email address")
	// Either a built-in role or a custom role must be selected
	dto.CheckField(dto.Role != "" || dto.CustomRoleID > 0, "role", "Please select a role")

	u := app.getAuthenticatedUser(r)
	churchID := 0
	inviterID := 0
	if u != nil {
		inviterID = u.ID
		if u.Edges.Church != nil {
			churchID = u.Edges.Church.ID
		}
	}

	if !dto.Valid() {
		customRoles, _ := app.customRoleModel.ListByChurch(r.Context(), churchID)
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{
			"customRoles": customRoles,
		}
		app.render(w, r, http.StatusUnprocessableEntity, "invite_member.gohtml", data)
		return
	}

	if churchID == 0 {
		customRoles, _ := app.customRoleModel.ListByChurch(r.Context(), 0)
		dto.AddNonFieldError("Cannot determine church — please contact support.")
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{
			"customRoles": customRoles,
		}
		app.render(w, r, http.StatusUnprocessableEntity, "invite_member.gohtml", data)
		return
	}

	token, err := app.churchModel.GenerateToken(dto.Email, 72)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	expiresAt := time.Now().Add(72 * time.Hour)
	_, err = app.invitationModel.Create(r.Context(), churchID, inviterID, &dto, token, expiresAt)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	acceptURL := fmt.Sprintf("http://localhost:3000/invite/accept?token=%s", token)
	inviteeName := dto.Name
	if inviteeName == "" {
		inviteeName = dto.Email
	}
	htmlBody := buildMemberInviteEmail(inviteeName, acceptURL)
	go func() { _ = sendHTMLEmail(dto.Email, "You're invited to join FaithConnect", htmlBody) }()

	app.sessionManager.Put(r.Context(), "flash", "Invitation sent to "+dto.Email)
	http.Redirect(w, r, "/workers", http.StatusSeeOther)
}
