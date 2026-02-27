package main

import (
	"net/http"
	"strconv"

	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /milestones — church-wide list
func (app *application) milestonesListGet(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	churchID := 0
	if u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}

	list, err := app.milestoneModel.ListByChurch(r.Context(), churchID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"milestones": list,
	}
	app.render(w, r, http.StatusOK, "milestones.gohtml", data)
}

// POST /members/{id}/milestones/new
func (app *application) milestoneNewPost(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	contactID, err := strconv.Atoi(idStr)
	if err != nil || contactID < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	u := app.getAuthenticatedUser(r)
	if u == nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}
	churchID := 0
	if u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}

	var dto models.MilestoneDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.MilestoneType), "milestone_type", "Please select a milestone type")
	dto.CheckField(validator.NotBlank(dto.EventDate), "event_date", "Please enter a date")

	if !dto.Valid() {
		app.sessionManager.Put(r.Context(), "flash_error", "Please fill in the required milestone fields")
		http.Redirect(w, r, "/members/"+idStr, http.StatusSeeOther)
		return
	}

	if _, err := app.milestoneModel.Create(r.Context(), &dto, contactID, churchID); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Milestone recorded successfully")
	http.Redirect(w, r, "/members/"+idStr, http.StatusSeeOther)
}

// POST /members/{id}/milestones/{mid}/delete
func (app *application) milestoneDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	mid, err := strconv.Atoi(r.PathValue("mid"))
	if err != nil || mid < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.milestoneModel.Delete(r.Context(), mid); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Milestone removed")
	http.Redirect(w, r, "/members/"+idStr, http.StatusSeeOther)
}
