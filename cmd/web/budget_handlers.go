package main

import (
	"net/http"
	"strconv"

	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /budgets
func (app *application) budgetsList(w http.ResponseWriter, r *http.Request) {
	cid := app.churchID(r)
	budgets, err := app.budgetModel.ListByChurch(r.Context(), cid)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{"budgets": budgets}
	app.render(w, r, http.StatusOK, "budgets.gohtml", data)
}

// GET /budgets/new
func (app *application) budgetNewGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = models.BudgetDto{}
	app.render(w, r, http.StatusOK, "budget_new.gohtml", data)
}

// POST /budgets/new
func (app *application) budgetNewPost(w http.ResponseWriter, r *http.Request) {
	var dto models.BudgetDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Name), "name", "Name is required")
	dto.CheckField(dto.FiscalYear > 0, "fiscal_year", "Fiscal year is required")
	dto.CheckField(validator.NotBlank(dto.StartDate), "start_date", "Start date is required")
	dto.CheckField(validator.NotBlank(dto.EndDate), "end_date", "End date is required")

	if !dto.Valid() {
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "budget_new.gohtml", data)
		return
	}

	cid := app.churchID(r)
	b, err := app.budgetModel.Create(r.Context(), &dto, cid)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Budget created successfully!")
	http.Redirect(w, r, "/budgets/"+strconv.Itoa(b.ID), http.StatusSeeOther)
}

// GET /budgets/{id}
func (app *application) budgetDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	summary, err := app.budgetModel.GetVsActual(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"summary": summary,
	}
	app.render(w, r, http.StatusOK, "budget_detail.gohtml", data)
}

// GET /budgets/{id}/edit
func (app *application) budgetEditGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	b, err := app.budgetModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	dto := models.BudgetDto{
		Name:       b.Name,
		FiscalYear: b.FiscalYear,
		Period:     b.Period.String(),
		StartDate:  b.StartDate.Format("2006-01-02"),
		EndDate:    b.EndDate.Format("2006-01-02"),
		Status:     b.Status.String(),
		Notes:      b.Notes,
	}

	data := app.newTemplateData(r)
	data.Form = dto
	data.Data = map[string]interface{}{"budget": b}
	app.render(w, r, http.StatusOK, "budget_edit.gohtml", data)
}

// POST /budgets/{id}/edit
func (app *application) budgetEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	var dto models.BudgetDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Name), "name", "Name is required")
	dto.CheckField(dto.FiscalYear > 0, "fiscal_year", "Fiscal year is required")
	dto.CheckField(validator.NotBlank(dto.StartDate), "start_date", "Start date is required")
	dto.CheckField(validator.NotBlank(dto.EndDate), "end_date", "End date is required")

	if !dto.Valid() {
		b, _ := app.budgetModel.GetByID(r.Context(), id)
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{"budget": b}
		app.render(w, r, http.StatusUnprocessableEntity, "budget_edit.gohtml", data)
		return
	}

	if _, err := app.budgetModel.Update(r.Context(), id, &dto); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Budget updated successfully!")
	http.Redirect(w, r, "/budgets/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /budgets/{id}/lines/add
func (app *application) budgetLineAdd(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	var dto models.BudgetLineDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if dto.Category == "" || dto.AllocatedAmount <= 0 {
		app.sessionManager.Put(r.Context(), "flash_error", "Please provide a category and a positive amount.")
		http.Redirect(w, r, "/budgets/"+strconv.Itoa(id), http.StatusSeeOther)
		return
	}

	if _, err := app.budgetModel.AddLine(r.Context(), id, &dto); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Budget line added.")
	http.Redirect(w, r, "/budgets/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /budgets/{id}/lines/{lid}/delete
func (app *application) budgetLineDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	lid, err := strconv.Atoi(r.PathValue("lid"))
	if err != nil || lid < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.budgetModel.DeleteLine(r.Context(), lid); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Budget line removed.")
	http.Redirect(w, r, "/budgets/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /budgets/{id}/delete
func (app *application) budgetDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.budgetModel.Delete(r.Context(), id); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Budget deleted.")
	http.Redirect(w, r, "/budgets", http.StatusSeeOther)
}
