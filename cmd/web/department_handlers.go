package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /departments
func (app *application) departmentsList(w http.ResponseWriter, r *http.Request) {
	churchID := app.churchID(r)
	depts, err := app.departmentModel.ListByChurch(r.Context(), churchID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{"departments": depts}
	app.render(w, r, http.StatusOK, "departments.gohtml", data)
}

// GET /departments/new
func (app *application) departmentNewGet(w http.ResponseWriter, r *http.Request) {
	members, _ := app.memberModel.ListContactsByChurch(r.Context(), app.churchID(r))
	data := app.newTemplateData(r)
	data.Form = models.DepartmentDto{IsActive: true}
	data.Data = map[string]interface{}{"members": members}
	app.render(w, r, http.StatusOK, "department_new.gohtml", data)
}

// POST /departments/new
func (app *application) departmentNewPost(w http.ResponseWriter, r *http.Request) {
	var dto models.DepartmentDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Name), "name", "Department name is required")
	dto.CheckField(validator.NotBlank(dto.DepartmentType), "department_type", "Please select a type")

	if !dto.Valid() {
		members, _ := app.memberModel.ListContactsByChurch(r.Context(), app.churchID(r))
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{"members": members}
		app.render(w, r, http.StatusUnprocessableEntity, "department_new.gohtml", data)
		return
	}

	d, err := app.departmentModel.Create(r.Context(), &dto, app.churchID(r))
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", fmt.Sprintf("%s department created.", dto.Name))
	http.Redirect(w, r, fmt.Sprintf("/departments/%d", d.ID), http.StatusSeeOther)
}

// GET /departments/{id}
func (app *application) departmentDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	dept, err := app.departmentModel.GetByID(r.Context(), id)
	if err != nil {
		if err == models.ErrRecordNotFound {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	members, _ := app.memberModel.ListContactsByChurch(r.Context(), app.churchID(r))

	// Build set of member IDs already in department
	inDept := map[int]bool{}
	for _, m := range dept.Edges.Members {
		inDept[m.ID] = true
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"department": dept,
		"allMembers": members,
		"inDept":     inDept,
	}
	app.render(w, r, http.StatusOK, "department_detail.gohtml", data)
}

// GET /departments/{id}/edit
func (app *application) departmentEditGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	dept, err := app.departmentModel.GetByID(r.Context(), id)
	if err != nil {
		if err == models.ErrRecordNotFound {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	members, _ := app.memberModel.ListContactsByChurch(r.Context(), app.churchID(r))

	data := app.newTemplateData(r)
	data.Form = models.DepartmentDto{
		Name:           dept.Name,
		Description:    dept.Description,
		DepartmentType: dept.DepartmentType.String(),
		LeaderID:       dept.LeaderID,
		IsActive:       dept.IsActive,
	}
	data.Data = map[string]interface{}{
		"department": dept,
		"members":    members,
	}
	app.render(w, r, http.StatusOK, "department_edit.gohtml", data)
}

// POST /departments/{id}/edit
func (app *application) departmentEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	var dto models.DepartmentDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Name), "name", "Department name is required")
	dto.CheckField(validator.NotBlank(dto.DepartmentType), "department_type", "Please select a type")

	if !dto.Valid() {
		dept, _ := app.departmentModel.GetByID(r.Context(), id)
		members, _ := app.memberModel.ListContactsByChurch(r.Context(), app.churchID(r))
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{"department": dept, "members": members}
		app.render(w, r, http.StatusUnprocessableEntity, "department_edit.gohtml", data)
		return
	}

	if err := app.departmentModel.Update(r.Context(), id, &dto); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Department updated.")
	http.Redirect(w, r, fmt.Sprintf("/departments/%d", id), http.StatusSeeOther)
}

// POST /departments/{id}/members/add
func (app *application) departmentAddMember(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	cid, err := strconv.Atoi(r.FormValue("contact_id"))
	if err != nil || cid < 1 {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if err := app.departmentModel.AddMember(r.Context(), id, cid); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Member added to department.")
	http.Redirect(w, r, fmt.Sprintf("/departments/%d", id), http.StatusSeeOther)
}

// POST /departments/{id}/members/{cid}/remove
func (app *application) departmentRemoveMember(w http.ResponseWriter, r *http.Request) {
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

	if err := app.departmentModel.RemoveMember(r.Context(), id, cid); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Member removed from department.")
	http.Redirect(w, r, fmt.Sprintf("/departments/%d", id), http.StatusSeeOther)
}

// POST /departments/{id}/delete
func (app *application) departmentDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.departmentModel.Delete(r.Context(), id); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Department deleted.")
	http.Redirect(w, r, "/departments", http.StatusSeeOther)
}
