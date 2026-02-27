package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /members
func (app *application) membersList(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	churchID := 0
	if u != nil && u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}

	members, err := app.memberModel.ListContactsByChurch(r.Context(), churchID)
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

// GET /members/new
func (app *application) memberNewGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = models.MemberDto{}
	app.render(w, r, http.StatusOK, "member_new.gohtml", data)
}

// POST /members/new
func (app *application) memberNewPost(w http.ResponseWriter, r *http.Request) {
	var dto models.MemberDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.FirstName), "first_name", "First name is required")
	dto.CheckField(validator.NotBlank(dto.LastName), "last_name", "Last name is required")

	if !dto.Valid() {
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "member_new.gohtml", data)
		return
	}

	u := app.getAuthenticatedUser(r)
	churchID := 0
	if u != nil && u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}

	if churchID == 0 {
		dto.AddNonFieldError("Cannot determine church — please contact support.")
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "member_new.gohtml", data)
		return
	}

	c, err := app.memberModel.CreateContact(r.Context(), &dto, churchID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", fmt.Sprintf("%s %s has been added to the congregation.", dto.FirstName, dto.LastName))
	http.Redirect(w, r, fmt.Sprintf("/members/%d", c.ID), http.StatusSeeOther)
}

// GET /members/{id}
func (app *application) memberDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	member, err := app.memberModel.GetContactByID(r.Context(), id)
	if err != nil {
		if err == models.ErrRecordNotFound {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	milestones, _ := app.milestoneModel.ListByContact(r.Context(), id)

	// Load relationships and build grouped family tree
	rels, _ := app.relationshipModel.ListByContact(r.Context(), id)
	familyGroups := buildFamilyTree(id, rels)

	// All church contacts for the "add relationship" dropdown (exclude self)
	u2 := app.getAuthenticatedUser(r)
	churchIDForRel := 0
	if u2 != nil && u2.Edges.Church != nil {
		churchIDForRel = u2.Edges.Church.ID
	}
	allContacts, _ := app.memberModel.ListContactsByChurch(r.Context(), churchIDForRel)

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"member":       member,
		"milestones":   milestones,
		"familyGroups": familyGroups,
		"allContacts":  allContacts,
	}
	app.render(w, r, http.StatusOK, "member_detail.gohtml", data)
}

// GET /members/{id}/edit
func (app *application) memberEditGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	member, err := app.memberModel.GetContactByID(r.Context(), id)
	if err != nil {
		if err == models.ErrRecordNotFound {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	// Build DTO from existing contact
	dto := models.MemberDto{
		FirstName:         member.FirstName,
		LastName:          member.LastName,
		MiddleName:        member.MiddleName,
		Email:             member.Email,
		Phone:             member.Phone,
		Gender:            string(member.Gender),
		MaritalStatus:     string(member.MaritalStatus),
		Occupation:        member.Occupation,
		AddressLine1:      member.AddressLine1,
		City:              member.City,
		Country:           member.Country,
		IDNumber:          member.IDNumber,
		Hometown:          member.Hometown,
		Region:            member.Region,
		SundaySchoolClass: member.SundaySchoolClass,
		DayBorn:           string(member.DayBorn),
		MembershipYear:    member.MembershipYear,
		HasSpouse:         member.HasSpouse,
		SpouseID:          member.SpouseID,
		IsBaptized:        member.IsBaptized,
		BaptizedBy:        member.BaptizedBy,
		BaptismChurch:     member.BaptismChurch,
		BaptismCertNumber: member.BaptismCertNumber,
	}
	if !member.DateOfBirth.IsZero() {
		dto.DateOfBirth = member.DateOfBirth.Format("2006-01-02")
	}
	if !member.BaptismDate.IsZero() {
		dto.BaptismDate = member.BaptismDate.Format("2006-01-02")
	}

	// Load church contacts for spouse dropdown
	u := app.getAuthenticatedUser(r)
	churchID := 0
	if u != nil && u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}
	allContacts, _ := app.memberModel.ListContactsByChurch(r.Context(), churchID)

	data := app.newTemplateData(r)
	data.Form = dto
	data.Data = map[string]interface{}{
		"member":      member,
		"allContacts": allContacts,
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

	member, err := app.memberModel.GetContactByID(r.Context(), id)
	if err != nil {
		if err == models.ErrRecordNotFound {
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
		allContacts, _ := app.memberModel.ListContactsByChurch(r.Context(), churchID)
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{
			"member":      member,
			"allContacts": allContacts,
		}
		app.render(w, r, http.StatusUnprocessableEntity, "member_edit.gohtml", data)
		return
	}

	if err := app.memberModel.UpdateContact(r.Context(), id, &dto); err != nil {
		app.serverError(w, r, err)
		return
	}

	// Sync spouse link bidirectionally
	if dto.HasSpouse && dto.SpouseID > 0 && dto.SpouseID != id {
		_, _ = app.db.Contact.UpdateOneID(id).SetSpouseID(dto.SpouseID).Save(r.Context())
		_, _ = app.db.Contact.UpdateOneID(dto.SpouseID).SetSpouseID(id).SetHasSpouse(true).Save(r.Context())
	} else if !dto.HasSpouse {
		oldSpouseID := member.SpouseID
		_, _ = app.db.Contact.UpdateOneID(id).ClearSpouseID().Save(r.Context())
		if oldSpouseID != 0 {
			_, _ = app.db.Contact.UpdateOneID(oldSpouseID).ClearSpouseID().SetHasSpouse(false).Save(r.Context())
		}
	}

	app.sessionManager.Put(r.Context(), "flash", "Member details updated successfully.")
	http.Redirect(w, r, fmt.Sprintf("/members/%d", id), http.StatusSeeOther)
}

// POST /members/{id}/avatar
func (app *application) memberAvatarPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	redirectURL := fmt.Sprintf("/members/%d", id)

	if err := r.ParseMultipartForm(maxImageInputSize); err != nil {
		app.sessionManager.Put(r.Context(), "flash_error", "File too large — maximum accepted size is 20 MB.")
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		app.sessionManager.Put(r.Context(), "flash_error", "No file selected.")
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}
	defer file.Close()

	url, err := app.uploadImage(file, header, "member-avatars")
	if err != nil {
		app.sessionManager.Put(r.Context(), "flash_error", err.Error())
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	if _, err := app.db.Contact.UpdateOneID(id).SetProfilePictureURL(url).Save(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Profile picture updated.")
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// POST /members/{id}/delete
func (app *application) memberDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.memberModel.DeleteContact(r.Context(), id); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Member record has been removed.")
	http.Redirect(w, r, "/members", http.StatusSeeOther)
}

// GET /members/{id}/attendance
func (app *application) memberAttendanceGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	member, err := app.memberModel.GetContactByID(r.Context(), id)
	if err != nil {
		if err == models.ErrRecordNotFound {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	records, err := app.attendanceModel.ListByContact(r.Context(), id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Compute summary stats
	totalPresent, totalLate := 0, 0
	for _, a := range records {
		switch string(a.Status) {
		case "present":
			totalPresent++
		case "late":
			totalLate++
		}
	}

	onTimeRate := 0
	if len(records) > 0 {
		onTimeRate = totalPresent * 100 / len(records)
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"member":       member,
		"records":      records,
		"totalPresent": totalPresent,
		"totalLate":    totalLate,
		"totalEvents":  len(records),
		"onTimeRate":   onTimeRate,
	}
	app.render(w, r, http.StatusOK, "member_attendance.gohtml", data)
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

// buildMemberInviteEmail constructs the HTML invitation email body.
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

