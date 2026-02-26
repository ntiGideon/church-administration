package main

import (
	"net/http"
	"strconv"

	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /visitors
func (app *application) visitorsList(w http.ResponseWriter, r *http.Request) {
	cid := app.churchID(r)
	visitors, err := app.visitorModel.ListByChurch(r.Context(), cid)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	counts, _ := app.visitorModel.CountByStatus(r.Context(), cid)
	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"visitors": visitors,
		"counts":   counts,
	}
	app.render(w, r, http.StatusOK, "visitors.gohtml", data)
}

// GET /visitors/new
func (app *application) visitorNewGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = models.VisitorDto{FollowUpStatus: "new"}
	app.render(w, r, http.StatusOK, "visitor_new.gohtml", data)
}

// POST /visitors/new
func (app *application) visitorNewPost(w http.ResponseWriter, r *http.Request) {
	var dto models.VisitorDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.FirstName), "first_name", "First name is required")
	dto.CheckField(validator.NotBlank(dto.LastName), "last_name", "Last name is required")
	dto.CheckField(validator.NotBlank(dto.VisitDate), "visit_date", "Visit date is required")

	if !dto.Valid() {
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "visitor_new.gohtml", data)
		return
	}

	cid := app.churchID(r)
	v, err := app.visitorModel.Create(r.Context(), &dto, cid)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Visitor record created!")
	http.Redirect(w, r, "/visitors/"+strconv.Itoa(v.ID), http.StatusSeeOther)
}

// GET /visitors/{id}
func (app *application) visitorDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	v, err := app.visitorModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{"visitor": v}
	app.render(w, r, http.StatusOK, "visitor_detail.gohtml", data)
}

// GET /visitors/{id}/edit
func (app *application) visitorEditGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	v, err := app.visitorModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	dto := models.VisitorDto{
		FirstName:      v.FirstName,
		LastName:       v.LastName,
		Email:          v.Email,
		Phone:          v.Phone,
		Address:        v.Address,
		VisitDate:      v.VisitDate.Format("2006-01-02"),
		HowHeard:       string(v.HowHeard),
		InvitedBy:      v.InvitedBy,
		Notes:          v.Notes,
		FollowUpStatus: string(v.FollowUpStatus),
	}

	data := app.newTemplateData(r)
	data.Form = dto
	data.Data = map[string]interface{}{"visitor": v}
	app.render(w, r, http.StatusOK, "visitor_edit.gohtml", data)
}

// POST /visitors/{id}/edit
func (app *application) visitorEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	var dto models.VisitorDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.FirstName), "first_name", "First name is required")
	dto.CheckField(validator.NotBlank(dto.LastName), "last_name", "Last name is required")
	dto.CheckField(validator.NotBlank(dto.VisitDate), "visit_date", "Visit date is required")

	if !dto.Valid() {
		v, _ := app.visitorModel.GetByID(r.Context(), id)
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{"visitor": v}
		app.render(w, r, http.StatusUnprocessableEntity, "visitor_edit.gohtml", data)
		return
	}

	if err := app.visitorModel.Update(r.Context(), id, &dto); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Visitor record updated!")
	http.Redirect(w, r, "/visitors/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /visitors/{id}/status
func (app *application) visitorUpdateStatus(w http.ResponseWriter, r *http.Request) {
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

	if err := app.visitorModel.UpdateStatus(r.Context(), id, status); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Follow-up status updated.")
	http.Redirect(w, r, "/visitors/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /visitors/{id}/delete
func (app *application) visitorDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.visitorModel.Delete(r.Context(), id); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Visitor record deleted.")
	http.Redirect(w, r, "/visitors", http.StatusSeeOther)
}
