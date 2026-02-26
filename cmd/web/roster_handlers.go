package main

import (
	"net/http"
	"strconv"

	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /rosters
func (app *application) rostersList(w http.ResponseWriter, r *http.Request) {
	cid := app.churchID(r)
	rosters, err := app.rosterModel.ListByChurch(r.Context(), cid)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{"rosters": rosters}
	app.render(w, r, http.StatusOK, "rosters.gohtml", data)
}

// GET /rosters/new
func (app *application) rosterNewGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = models.RosterDto{}
	app.render(w, r, http.StatusOK, "roster_new.gohtml", data)
}

// POST /rosters/new
func (app *application) rosterNewPost(w http.ResponseWriter, r *http.Request) {
	var dto models.RosterDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Title), "title", "Title is required")
	dto.CheckField(validator.NotBlank(dto.ServiceDate), "service_date", "Service date is required")

	if !dto.Valid() {
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "roster_new.gohtml", data)
		return
	}

	cid := app.churchID(r)
	roster, err := app.rosterModel.Create(r.Context(), &dto, cid)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Roster created!")
	http.Redirect(w, r, "/rosters/"+strconv.Itoa(roster.ID), http.StatusSeeOther)
}

// GET /rosters/{id}
func (app *application) rosterDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	roster, err := app.rosterModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	cid := app.churchID(r)
	allMembers, _ := app.memberModel.ListContactsByChurch(r.Context(), cid)

	// Filter out already-assigned contacts
	assigned := map[int]bool{}
	for _, e := range roster.Edges.Entries {
		assigned[e.ContactID] = true
	}
	type memberOpt struct {
		ID   int
		Name string
	}
	var available []memberOpt
	for _, c := range allMembers {
		if !assigned[c.ID] {
			available = append(available, memberOpt{ID: c.ID, Name: c.FirstName + " " + c.LastName})
		}
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"roster":    roster,
		"available": available,
	}
	app.render(w, r, http.StatusOK, "roster_detail.gohtml", data)
}

// GET /rosters/{id}/edit
func (app *application) rosterEditGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	roster, err := app.rosterModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	dto := models.RosterDto{
		Title:       roster.Title,
		ServiceDate: roster.ServiceDate.Format("2006-01-02"),
		Department:  roster.Department,
		Notes:       roster.Notes,
	}

	data := app.newTemplateData(r)
	data.Form = dto
	data.Data = map[string]interface{}{"roster": roster}
	app.render(w, r, http.StatusOK, "roster_edit.gohtml", data)
}

// POST /rosters/{id}/edit
func (app *application) rosterEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	var dto models.RosterDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Title), "title", "Title is required")
	dto.CheckField(validator.NotBlank(dto.ServiceDate), "service_date", "Service date is required")

	if !dto.Valid() {
		roster, _ := app.rosterModel.GetByID(r.Context(), id)
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{"roster": roster}
		app.render(w, r, http.StatusUnprocessableEntity, "roster_edit.gohtml", data)
		return
	}

	if err := app.rosterModel.Update(r.Context(), id, &dto); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Roster updated!")
	http.Redirect(w, r, "/rosters/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /rosters/{id}/entries/add
func (app *application) rosterAddEntry(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	contactID, err := strconv.Atoi(r.FormValue("contact_id"))
	if err != nil || contactID < 1 {
		app.sessionManager.Put(r.Context(), "flash_error", "Please select a valid member.")
		http.Redirect(w, r, "/rosters/"+strconv.Itoa(id), http.StatusSeeOther)
		return
	}

	role := r.FormValue("role")
	if role == "" {
		app.sessionManager.Put(r.Context(), "flash_error", "Please enter a role.")
		http.Redirect(w, r, "/rosters/"+strconv.Itoa(id), http.StatusSeeOther)
		return
	}

	already, err := app.rosterModel.IsAssigned(r.Context(), id, contactID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if already {
		app.sessionManager.Put(r.Context(), "flash_error", "This member is already on this roster.")
		http.Redirect(w, r, "/rosters/"+strconv.Itoa(id), http.StatusSeeOther)
		return
	}

	notes := r.FormValue("notes")
	if _, err := app.rosterModel.AddEntry(r.Context(), id, contactID, role, notes); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Volunteer added to roster.")
	http.Redirect(w, r, "/rosters/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /rosters/{id}/entries/{eid}/remove
func (app *application) rosterRemoveEntry(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	eid, err := strconv.Atoi(r.PathValue("eid"))
	if err != nil || eid < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.rosterModel.RemoveEntry(r.Context(), eid); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Volunteer removed from roster.")
	http.Redirect(w, r, "/rosters/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /rosters/{id}/delete
func (app *application) rosterDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.rosterModel.Delete(r.Context(), id); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Roster deleted.")
	http.Redirect(w, r, "/rosters", http.StatusSeeOther)
}
