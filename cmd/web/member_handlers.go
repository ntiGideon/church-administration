package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ntiGideon/ent/user"
	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /members/{id}/edit
func (app *application) memberEditGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}
	member, err := app.memberModel.GetByID(r.Context(), id)
	if err != nil {
		if err == models.ErrUserNotFound {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	// Build DTO from existing contact
	dto := models.MemberDto{}
	if c := member.Edges.Contact; c != nil {
		dto.FirstName = c.FirstName
		dto.LastName = c.LastName
		dto.MiddleName = c.MiddleName
		dto.Email = c.Email
		dto.Phone = c.Phone
		dto.Gender = string(c.Gender)
		if !c.DateOfBirth.IsZero() {
			dto.DateOfBirth = c.DateOfBirth.Format("2006-01-02")
		}
		dto.MaritalStatus = string(c.MaritalStatus)
		dto.Occupation = c.Occupation
		dto.AddressLine1 = c.AddressLine1
		dto.City = c.City
		dto.Country = c.Country
		dto.IDNumber = c.IDNumber
		dto.Hometown = c.Hometown
		dto.Region = c.Region
		dto.SundaySchoolClass = c.SundaySchoolClass
		dto.DayBorn = string(c.DayBorn)
		dto.MembershipYear = c.MembershipYear
		dto.HasSpouse = c.HasSpouse
		dto.SpouseID = c.SpouseID
		dto.IsBaptized = c.IsBaptized
		dto.BaptizedBy = c.BaptizedBy
		dto.BaptismChurch = c.BaptismChurch
		dto.BaptismCertNumber = c.BaptismCertNumber
		if !c.BaptismDate.IsZero() {
			dto.BaptismDate = c.BaptismDate.Format("2006-01-02")
		}
	}

	// Load church members for spouse dropdown
	u := app.getAuthenticatedUser(r)
	churchID := 0
	if u != nil && u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}
	allMembers, _ := app.memberModel.ListByChurch(r.Context(), churchID)

	data := app.newTemplateData(r)
	data.Form = dto
	data.Data = map[string]interface{}{
		"member":     member,
		"allMembers": allMembers,
	}
	app.render(w, r, http.StatusOK, "member_edit.gohtml", data)
}

// POST /members/{id}/edit
func (app *application) memberEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}
	member, err := app.memberModel.GetByID(r.Context(), id)
	if err != nil {
		if err == models.ErrUserNotFound {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	var dto models.MemberDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.FirstName), "first_name", "First name is required")
	dto.CheckField(validator.NotBlank(dto.LastName), "last_name", "Last name is required")

	if !dto.Valid() {
		u := app.getAuthenticatedUser(r)
		churchID := 0
		if u != nil && u.Edges.Church != nil {
			churchID = u.Edges.Church.ID
		}
		allMembers, _ := app.memberModel.ListByChurch(r.Context(), churchID)
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{
			"member":     member,
			"allMembers": allMembers,
		}
		app.render(w, r, http.StatusUnprocessableEntity, "member_edit.gohtml", data)
		return
	}

	if member.Edges.Contact == nil {
		app.serverError(w, r, models.ErrUserNotFound)
		return
	}
	contactID := member.Edges.Contact.ID

	if err := app.memberModel.UpdateContact(r.Context(), contactID, &dto); err != nil {
		app.serverError(w, r, err)
		return
	}

	// Sync spouse link bidirectionally
	if dto.HasSpouse && dto.SpouseID > 0 && dto.SpouseID != contactID {
		// Set this contact's spouse_id
		_, _ = app.db.Contact.UpdateOneID(contactID).SetSpouseID(dto.SpouseID).Save(r.Context())
		// Set the other side too (if not already pointing to someone else)
		_, _ = app.db.Contact.UpdateOneID(dto.SpouseID).SetSpouseID(contactID).SetHasSpouse(true).Save(r.Context())
	} else if !dto.HasSpouse {
		// Clear both sides
		old := member.Edges.Contact.SpouseID
		_, _ = app.db.Contact.UpdateOneID(contactID).ClearSpouseID().Save(r.Context())
		if old != 0 {
			_, _ = app.db.Contact.UpdateOneID(old).ClearSpouseID().SetHasSpouse(false).Save(r.Context())
		}
	}

	app.sessionManager.Put(r.Context(), "flash", "Member details updated successfully.")
	http.Redirect(w, r, fmt.Sprintf("/members/%d", id), http.StatusSeeOther)
}

// derefStr safely dereferences an optional string pointer.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// derefInt safely dereferences an optional int pointer.
func derefInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// GET /members
func (app *application) membersList(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	// super_admin sees all members; others see only their church
	churchID := 0
	if u != nil && u.Role != user.RoleSuperAdmin && u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}

	members, err := app.memberModel.ListByChurch(r.Context(), churchID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"members": members,
	}
	app.render(w, r, http.StatusOK, "members.gohtml", data)
}

// GET /members/{id}
func (app *application) memberDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	member, err := app.memberModel.GetByID(r.Context(), id)
	if err != nil {
		if err == models.ErrUserNotFound {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"member": member,
	}
	app.render(w, r, http.StatusOK, "member_detail.gohtml", data)
}

// POST /members/{id}/deactivate
func (app *application) memberDeactivate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.memberModel.Deactivate(r.Context(), id); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Member has been deactivated.")
	http.Redirect(w, r, "/members", http.StatusSeeOther)
}

// GET /members/new — show invite member form
func (app *application) inviteMemberGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = models.MemberInviteDto{}
	app.render(w, r, http.StatusOK, "invite_member.gohtml", data)
}

// POST /members/new — create invitation record and send email
func (app *application) inviteMemberPost(w http.ResponseWriter, r *http.Request) {
	var dto models.MemberInviteDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Email), "email", "Email is required")
	dto.CheckField(validator.Matches(dto.Email, validator.EmailRX), "email", "Must be a valid email address")
	dto.CheckField(validator.NotBlank(dto.Role), "role", "Role is required")

	if !dto.Valid() {
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "invite_member.gohtml", data)
		return
	}

	u := app.getAuthenticatedUser(r)
	churchID := 0
	inviterID := 0
	if u != nil {
		inviterID = u.ID
		if u.Edges.Church != nil {
			churchID = u.Edges.Church.ID
		}
	}

	if churchID == 0 {
		dto.AddNonFieldError("Cannot determine church — please contact support.")
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "invite_member.gohtml", data)
		return
	}

	// Generate a JWT token that embeds the invitee's email
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

	// Send invitation email (fire-and-forget)
	acceptURL := fmt.Sprintf("http://localhost:3000/invite/accept?token=%s", token)
	inviteeName := dto.Name
	if inviteeName == "" {
		inviteeName = dto.Email
	}
	htmlBody := buildMemberInviteEmail(inviteeName, acceptURL)
	go func() { _ = sendHTMLEmail(dto.Email, "You're invited to join FaithConnect", htmlBody) }()

	app.sessionManager.Put(r.Context(), "flash", "Invitation sent to "+dto.Email)
	http.Redirect(w, r, "/members", http.StatusSeeOther)
}

func buildMemberInviteEmail(name, acceptURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family:sans-serif;background:#f8f9fa;padding:40px 20px">
<div style="max-width:520px;margin:0 auto;background:white;border-radius:12px;overflow:hidden;box-shadow:0 4px 20px rgba(0,0,0,0.1)">
  <div style="background:linear-gradient(135deg,#50222D,#6B2F3D);padding:32px;text-align:center">
    <h1 style="color:white;margin:0;font-size:24px">You're Invited!</h1>
    <p style="color:rgba(255,255,255,0.8);margin:8px 0 0">FaithConnect Church Management</p>
  </div>
  <div style="padding:32px">
    <p style="color:#212529;font-size:16px">Hello %s,</p>
    <p style="color:#495057;font-size:15px;line-height:1.6">
      You have been invited to join your church on FaithConnect — a modern church management platform.
      Click the button below to create your account and get started.
    </p>
    <div style="text-align:center;margin:32px 0">
      <a href="%s" style="background:#50222D;color:white;text-decoration:none;padding:14px 32px;border-radius:8px;font-weight:600;font-size:15px;display:inline-block">
        Accept Invitation
      </a>
    </div>
    <p style="color:#868E96;font-size:13px">This invitation link expires in 72 hours.</p>
  </div>
</div>
</body></html>`, name, acceptURL)
}
