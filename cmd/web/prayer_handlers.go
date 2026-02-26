package main

import (
	"net/http"
	"strconv"

	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /prayer
func (app *application) prayerList(w http.ResponseWriter, r *http.Request) {
	cid := app.churchID(r)
	requests, err := app.prayerModel.ListByChurch(r.Context(), cid)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	counts, _ := app.prayerModel.CountByStatus(r.Context(), cid)
	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"requests": requests,
		"counts":   counts,
	}
	app.render(w, r, http.StatusOK, "prayer.gohtml", data)
}

// GET /prayer/new
func (app *application) prayerNewGet(w http.ResponseWriter, r *http.Request) {
	cid := app.churchID(r)
	members, _ := app.memberModel.ListContactsByChurch(r.Context(), cid)
	data := app.newTemplateData(r)
	data.Form = models.PrayerRequestDto{}
	data.Data = map[string]interface{}{"members": members}
	app.render(w, r, http.StatusOK, "prayer_new.gohtml", data)
}

// POST /prayer/new
func (app *application) prayerNewPost(w http.ResponseWriter, r *http.Request) {
	var dto models.PrayerRequestDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Title), "title", "Title is required")
	dto.CheckField(validator.NotBlank(dto.Body), "body", "Request details are required")

	if !dto.Valid() {
		cid := app.churchID(r)
		members, _ := app.memberModel.ListContactsByChurch(r.Context(), cid)
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{"members": members}
		app.render(w, r, http.StatusUnprocessableEntity, "prayer_new.gohtml", data)
		return
	}

	cid := app.churchID(r)
	pr, err := app.prayerModel.Create(r.Context(), &dto, cid)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Prayer request submitted!")
	http.Redirect(w, r, "/prayer/"+strconv.Itoa(pr.ID), http.StatusSeeOther)
}

// GET /prayer/{id}
func (app *application) prayerDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	pr, err := app.prayerModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{"request": pr}
	app.render(w, r, http.StatusOK, "prayer_detail.gohtml", data)
}

// GET /prayer/{id}/edit
func (app *application) prayerEditGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	pr, err := app.prayerModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	dto := models.PrayerRequestDto{
		Title:         pr.Title,
		Body:          pr.Body,
		RequesterName: pr.RequesterName,
		IsAnonymous:   pr.IsAnonymous,
		IsPrivate:     pr.IsPrivate,
	}

	data := app.newTemplateData(r)
	data.Form = dto
	data.Data = map[string]interface{}{"request": pr}
	app.render(w, r, http.StatusOK, "prayer_edit.gohtml", data)
}

// POST /prayer/{id}/edit
func (app *application) prayerEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	var dto models.PrayerRequestDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Title), "title", "Title is required")
	dto.CheckField(validator.NotBlank(dto.Body), "body", "Request details are required")

	if !dto.Valid() {
		pr, _ := app.prayerModel.GetByID(r.Context(), id)
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{"request": pr}
		app.render(w, r, http.StatusUnprocessableEntity, "prayer_edit.gohtml", data)
		return
	}

	if err := app.prayerModel.Update(r.Context(), id, &dto); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Prayer request updated!")
	http.Redirect(w, r, "/prayer/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /prayer/{id}/status
func (app *application) prayerUpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	status := r.FormValue("status")
	if status == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if err := app.prayerModel.UpdateStatus(r.Context(), id, status); err != nil {
		app.serverError(w, r, err)
		return
	}

	msg := map[string]string{
		"active":   "Marked as active.",
		"answered": "Praise God! Marked as answered.",
		"closed":   "Request closed.",
	}[status]
	if msg == "" {
		msg = "Status updated."
	}

	app.sessionManager.Put(r.Context(), "flash", msg)
	http.Redirect(w, r, "/prayer/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /prayer/{id}/delete
func (app *application) prayerDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.prayerModel.Delete(r.Context(), id); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Prayer request deleted.")
	http.Redirect(w, r, "/prayer", http.StatusSeeOther)
}
