package main

import (
	"net/http"
	"strconv"

	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /pastoral
func (app *application) pastoralList(w http.ResponseWriter, r *http.Request) {
	cid := app.churchID(r)
	notes, err := app.pastoralModel.ListByChurch(r.Context(), cid)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	pending := app.pastoralModel.CountPendingFollowUp(r.Context(), cid)

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"notes":   notes,
		"pending": pending,
	}
	app.render(w, r, http.StatusOK, "pastoral.gohtml", data)
}

// GET /pastoral/new
func (app *application) pastoralNewGet(w http.ResponseWriter, r *http.Request) {
	cid := app.churchID(r)
	members, _ := app.memberModel.ListContactsByChurch(r.Context(), cid)

	data := app.newTemplateData(r)
	data.Form = models.PastoralNoteDto{}
	data.Data = map[string]interface{}{"members": members}
	app.render(w, r, http.StatusOK, "pastoral_new.gohtml", data)
}

// POST /pastoral/new
func (app *application) pastoralNewPost(w http.ResponseWriter, r *http.Request) {
	cid := app.churchID(r)
	u := app.getAuthenticatedUser(r)

	var dto models.PastoralNoteDto
	if err := app.formDecoder.Decode(&dto, r.PostForm); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.VisitDate), "visit_date", "Visit date is required.")
	dto.CheckField(validator.NotBlank(dto.CareType), "care_type", "Care type is required.")
	dto.CheckField(validator.NotBlank(dto.Notes), "notes", "Notes are required.")
	dto.CheckField(dto.ContactID > 0, "contact_id", "Please select a member.")

	if !dto.Valid() {
		members, _ := app.memberModel.ListContactsByChurch(r.Context(), cid)
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{"members": members}
		app.render(w, r, http.StatusUnprocessableEntity, "pastoral_new.gohtml", data)
		return
	}

	userID := 0
	if u != nil {
		userID = u.ID
	}

	n, err := app.pastoralModel.Create(r.Context(), &dto, dto.ContactID, cid, userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Pastoral care note recorded successfully.")
	http.Redirect(w, r, "/pastoral/"+strconv.Itoa(n.ID), http.StatusSeeOther)
}

// GET /pastoral/{id}
func (app *application) pastoralDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	note, err := app.pastoralModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{"note": note}
	app.render(w, r, http.StatusOK, "pastoral_detail.gohtml", data)
}

// GET /pastoral/{id}/edit
func (app *application) pastoralEditGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	note, err := app.pastoralModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	cid := app.churchID(r)
	members, _ := app.memberModel.ListContactsByChurch(r.Context(), cid)

	dto := models.PastoralNoteDto{
		VisitDate:     note.VisitDate.Format("2006-01-02"),
		CareType:      note.CareType.String(),
		Notes:         note.Notes,
		NeedsFollowUp: note.NeedsFollowUp,
		ContactID:     note.ContactID,
	}
	if !note.FollowUpDate.IsZero() {
		dto.FollowUpDate = note.FollowUpDate.Format("2006-01-02")
	}

	data := app.newTemplateData(r)
	data.Form = dto
	data.Data = map[string]interface{}{
		"note":    note,
		"members": members,
	}
	app.render(w, r, http.StatusOK, "pastoral_edit.gohtml", data)
}

// POST /pastoral/{id}/edit
func (app *application) pastoralEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	var dto models.PastoralNoteDto
	if err := app.formDecoder.Decode(&dto, r.PostForm); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.VisitDate), "visit_date", "Visit date is required.")
	dto.CheckField(validator.NotBlank(dto.CareType), "care_type", "Care type is required.")
	dto.CheckField(validator.NotBlank(dto.Notes), "notes", "Notes are required.")

	if !dto.Valid() {
		cid := app.churchID(r)
		members, _ := app.memberModel.ListContactsByChurch(r.Context(), cid)
		note, _ := app.pastoralModel.GetByID(r.Context(), id)
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{
			"note":    note,
			"members": members,
		}
		app.render(w, r, http.StatusUnprocessableEntity, "pastoral_edit.gohtml", data)
		return
	}

	if _, err := app.pastoralModel.Update(r.Context(), id, &dto); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Pastoral care note updated.")
	http.Redirect(w, r, "/pastoral/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /pastoral/{id}/followup
func (app *application) pastoralMarkFollowUpDone(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.pastoralModel.MarkFollowUpDone(r.Context(), id); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Follow-up marked as complete.")
	http.Redirect(w, r, "/pastoral/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /pastoral/{id}/delete
func (app *application) pastoralDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.pastoralModel.Delete(r.Context(), id); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Pastoral care note deleted.")
	http.Redirect(w, r, "/pastoral", http.StatusSeeOther)
}
