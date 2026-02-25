package main

import (
	"net/http"
	"strconv"

	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /events
func (app *application) eventsList(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	churchID := 0
	if u != nil && u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}

	events, err := app.eventModel.List(r.Context(), churchID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"events": events,
	}
	app.render(w, r, http.StatusOK, "events.gohtml", data)
}

// GET /events/new
func (app *application) eventNewGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = models.EventDto{}
	app.render(w, r, http.StatusOK, "event_new.gohtml", data)
}

// POST /events/new
func (app *application) eventNewPost(w http.ResponseWriter, r *http.Request) {
	var dto models.EventDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Title), "title", "Title is required")
	dto.CheckField(validator.NotBlank(dto.Description), "description", "Description is required")
	dto.CheckField(validator.NotBlank(dto.StartTime), "start_time", "Start time is required")
	dto.CheckField(validator.NotBlank(dto.EndTime), "end_time", "End time is required")
	dto.CheckField(validator.NotBlank(dto.Location), "location", "Location is required")
	dto.CheckField(validator.NotBlank(dto.EventType), "event_type", "Event type is required")

	if !dto.Valid() {
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "event_new.gohtml", data)
		return
	}

	u := app.getAuthenticatedUser(r)
	churchID := 0
	if u != nil && u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}

	_, err := app.eventModel.Create(r.Context(), &dto, churchID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Event created successfully!")
	http.Redirect(w, r, "/events", http.StatusSeeOther)
}

// GET /events/{id}
func (app *application) eventDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	e, err := app.eventModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"event": e,
	}
	app.render(w, r, http.StatusOK, "event_detail.gohtml", data)
}

// POST /events/{id}/publish — toggle published state
func (app *application) eventTogglePublish(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	e, err := app.eventModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.eventModel.SetPublished(r.Context(), id, !e.IsPublished); err != nil {
		app.serverError(w, r, err)
		return
	}

	msg := "Event published."
	if e.IsPublished {
		msg = "Event moved to draft."
	}
	app.sessionManager.Put(r.Context(), "flash", msg)
	http.Redirect(w, r, "/events/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /events/{id}/attendance — update attendance count
func (app *application) eventUpdateAttendance(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	count, err := strconv.Atoi(r.FormValue("attendance_count"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if err := app.eventModel.UpdateAttendance(r.Context(), id, count); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Attendance updated successfully!")
	http.Redirect(w, r, "/events/"+strconv.Itoa(id), http.StatusSeeOther)
}
