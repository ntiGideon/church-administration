package main

import (
	"net/http"
	"strconv"

	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /sermons
func (app *application) sermonsList(w http.ResponseWriter, r *http.Request) {
	cid := app.churchID(r)
	sermons, err := app.sermonModel.ListByChurch(r.Context(), cid)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{"sermons": sermons}
	app.render(w, r, http.StatusOK, "sermons.gohtml", data)
}

// GET /sermons/new
func (app *application) sermonNewGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = models.SermonDto{}
	app.render(w, r, http.StatusOK, "sermon_new.gohtml", data)
}

// POST /sermons/new
func (app *application) sermonNewPost(w http.ResponseWriter, r *http.Request) {
	var dto models.SermonDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Title), "title", "Title is required")
	dto.CheckField(validator.NotBlank(dto.Speaker), "speaker", "Speaker name is required")
	dto.CheckField(validator.NotBlank(dto.ServiceDate), "service_date", "Service date is required")

	if !dto.Valid() {
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "sermon_new.gohtml", data)
		return
	}

	cid := app.churchID(r)
	s, err := app.sermonModel.Create(r.Context(), &dto, cid)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Sermon added to the library!")
	http.Redirect(w, r, "/sermons/"+strconv.Itoa(s.ID), http.StatusSeeOther)
}

// GET /sermons/{id}
func (app *application) sermonDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	s, err := app.sermonModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{"sermon": s}
	app.render(w, r, http.StatusOK, "sermon_detail.gohtml", data)
}

// GET /sermons/{id}/edit
func (app *application) sermonEditGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	s, err := app.sermonModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	dto := models.SermonDto{
		Title:       s.Title,
		Speaker:     s.Speaker,
		Series:      s.Series,
		Scripture:   s.Scripture,
		Description: s.Description,
		MediaURL:    s.MediaURL,
		ServiceDate: s.ServiceDate.Format("2006-01-02"),
		IsPublished: s.IsPublished,
	}

	data := app.newTemplateData(r)
	data.Form = dto
	data.Data = map[string]interface{}{"sermon": s}
	app.render(w, r, http.StatusOK, "sermon_edit.gohtml", data)
}

// POST /sermons/{id}/edit
func (app *application) sermonEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	var dto models.SermonDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Title), "title", "Title is required")
	dto.CheckField(validator.NotBlank(dto.Speaker), "speaker", "Speaker name is required")
	dto.CheckField(validator.NotBlank(dto.ServiceDate), "service_date", "Service date is required")

	if !dto.Valid() {
		s, _ := app.sermonModel.GetByID(r.Context(), id)
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{"sermon": s}
		app.render(w, r, http.StatusUnprocessableEntity, "sermon_edit.gohtml", data)
		return
	}

	if err := app.sermonModel.Update(r.Context(), id, &dto); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Sermon updated!")
	http.Redirect(w, r, "/sermons/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /sermons/{id}/publish
func (app *application) sermonTogglePublish(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	s, err := app.sermonModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	newState := !s.IsPublished
	if err := app.sermonModel.TogglePublish(r.Context(), id, newState); err != nil {
		app.serverError(w, r, err)
		return
	}

	msg := "Sermon published."
	if !newState {
		msg = "Sermon unpublished."
	}
	app.sessionManager.Put(r.Context(), "flash", msg)
	http.Redirect(w, r, "/sermons/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /sermons/{id}/delete
func (app *application) sermonDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.sermonModel.Delete(r.Context(), id); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Sermon deleted.")
	http.Redirect(w, r, "/sermons", http.StatusSeeOther)
}
