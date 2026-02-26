package main

import (
	"net/http"
	"strconv"

	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

func (app *application) churchID(r *http.Request) int {
	u := app.getAuthenticatedUser(r)
	if u != nil && u.Edges.Church != nil {
		return u.Edges.Church.ID
	}
	return 0
}

// GET /groups
func (app *application) groupsList(w http.ResponseWriter, r *http.Request) {
	cid := app.churchID(r)
	groups, err := app.groupModel.ListByChurch(r.Context(), cid)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{"groups": groups}
	app.render(w, r, http.StatusOK, "groups.gohtml", data)
}

// GET /groups/new
func (app *application) groupNewGet(w http.ResponseWriter, r *http.Request) {
	cid := app.churchID(r)
	members, _ := app.memberModel.ListContactsByChurch(r.Context(), cid)
	data := app.newTemplateData(r)
	data.Form = models.GroupDto{}
	data.Data = map[string]interface{}{"members": members}
	app.render(w, r, http.StatusOK, "group_new.gohtml", data)
}

// POST /groups/new
func (app *application) groupNewPost(w http.ResponseWriter, r *http.Request) {
	var dto models.GroupDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Name), "name", "Name is required")
	dto.CheckField(validator.NotBlank(dto.GroupType), "group_type", "Group type is required")

	if !dto.Valid() {
		cid := app.churchID(r)
		members, _ := app.memberModel.ListContactsByChurch(r.Context(), cid)
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{"members": members}
		app.render(w, r, http.StatusUnprocessableEntity, "group_new.gohtml", data)
		return
	}

	cid := app.churchID(r)
	g, err := app.groupModel.Create(r.Context(), &dto, cid)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Group created successfully!")
	http.Redirect(w, r, "/groups/"+strconv.Itoa(g.ID), http.StatusSeeOther)
}

// GET /groups/{id}
func (app *application) groupDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	g, err := app.groupModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	cid := app.churchID(r)
	allMembers, _ := app.memberModel.ListContactsByChurch(r.Context(), cid)

	// Build set of already-in-group contact IDs
	inGroup := map[int]bool{}
	for _, m := range g.Edges.Members {
		inGroup[m.ID] = true
	}
	// Contacts not yet in the group (available to add)
	type memberOpt struct {
		ID   int
		Name string
	}
	var availableMembers []memberOpt
	for _, c := range allMembers {
		if !inGroup[c.ID] {
			availableMembers = append(availableMembers, memberOpt{ID: c.ID, Name: c.FirstName + " " + c.LastName})
		}
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"group":            g,
		"availableMembers": availableMembers,
	}
	app.render(w, r, http.StatusOK, "group_detail.gohtml", data)
}

// GET /groups/{id}/edit
func (app *application) groupEditGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	g, err := app.groupModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	cid := app.churchID(r)
	members, _ := app.memberModel.ListContactsByChurch(r.Context(), cid)

	dto := models.GroupDto{
		Name:        g.Name,
		Description: g.Description,
		GroupType:   g.GroupType.String(),
		MeetingTime: g.MeetingTime,
		Location:    g.Location,
		IsActive:    g.IsActive,
	}
	if string(g.MeetingDay) != "" {
		dto.MeetingDay = g.MeetingDay.String()
	}
	if g.Edges.Leader != nil {
		dto.LeaderID = g.Edges.Leader.ID
	}

	data := app.newTemplateData(r)
	data.Form = dto
	data.Data = map[string]interface{}{
		"group":   g,
		"members": members,
	}
	app.render(w, r, http.StatusOK, "group_edit.gohtml", data)
}

// POST /groups/{id}/edit
func (app *application) groupEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	var dto models.GroupDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Name), "name", "Name is required")
	dto.CheckField(validator.NotBlank(dto.GroupType), "group_type", "Group type is required")

	if !dto.Valid() {
		g, _ := app.groupModel.GetByID(r.Context(), id)
		cid := app.churchID(r)
		members, _ := app.memberModel.ListContactsByChurch(r.Context(), cid)
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{"group": g, "members": members}
		app.render(w, r, http.StatusUnprocessableEntity, "group_edit.gohtml", data)
		return
	}

	if err := app.groupModel.Update(r.Context(), id, &dto); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Group updated successfully!")
	http.Redirect(w, r, "/groups/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /groups/{id}/members/add
func (app *application) groupAddMember(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	contactID, err := strconv.Atoi(r.FormValue("contact_id"))
	if err != nil || contactID < 1 {
		app.sessionManager.Put(r.Context(), "flash_error", "Please select a valid member.")
		http.Redirect(w, r, "/groups/"+strconv.Itoa(id), http.StatusSeeOther)
		return
	}

	already, err := app.groupModel.IsMember(r.Context(), id, contactID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if already {
		app.sessionManager.Put(r.Context(), "flash_error", "This member is already in the group.")
		http.Redirect(w, r, "/groups/"+strconv.Itoa(id), http.StatusSeeOther)
		return
	}

	if err := app.groupModel.AddMember(r.Context(), id, contactID); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Member added to group.")
	http.Redirect(w, r, "/groups/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /groups/{id}/members/{cid}/remove
func (app *application) groupRemoveMember(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	cid, err := strconv.Atoi(r.PathValue("cid"))
	if err != nil || cid < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.groupModel.RemoveMember(r.Context(), id, cid); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Member removed from group.")
	http.Redirect(w, r, "/groups/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /groups/{id}/delete
func (app *application) groupDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.groupModel.Delete(r.Context(), id); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Group deleted.")
	http.Redirect(w, r, "/groups", http.StatusSeeOther)
}
