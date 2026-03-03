package main

import (
	"net/http"
	"strconv"

	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/permissions"
	"github.com/ntiGideon/internal/validator"
)

// GET /church/settings/roles
func (app *application) customRolesList(w http.ResponseWriter, r *http.Request) {
	churchID := app.churchID(r)
	roles, err := app.customRoleModel.ListByChurch(r.Context(), churchID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"roles":      roles,
		"activeTab":  "roles",
	}
	app.render(w, r, http.StatusOK, "custom_roles.gohtml", data)
}

// GET /church/settings/roles/new
func (app *application) customRoleNewGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = models.CustomRoleDto{IsActive: true}
	data.Data = map[string]interface{}{
		"permGroups":   permissions.PermissionGroups,
		"selectedPerms": map[string]bool{},
		"activeTab":    "roles",
	}
	app.render(w, r, http.StatusOK, "custom_role_form.gohtml", data)
}

// POST /church/settings/roles/new
func (app *application) customRoleNewPost(w http.ResponseWriter, r *http.Request) {
	var dto models.CustomRoleDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Name), "name", "Role name is required")

	if !dto.Valid() {
		selected := make(map[string]bool, len(dto.Permissions))
		for _, p := range dto.Permissions {
			selected[p] = true
		}
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{
			"permGroups":   permissions.PermissionGroups,
			"selectedPerms": selected,
			"activeTab":    "roles",
		}
		app.render(w, r, http.StatusUnprocessableEntity, "custom_role_form.gohtml", data)
		return
	}

	churchID := app.churchID(r)
	if _, err := app.customRoleModel.Create(r.Context(), churchID, &dto); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Custom role created successfully.")
	http.Redirect(w, r, "/church/settings/roles", http.StatusSeeOther)
}

// GET /church/settings/roles/{id}
func (app *application) customRoleEditGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	role, err := app.customRoleModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	perms := models.ParsePermissions(role.Permissions)
	selected := make(map[string]bool, len(perms))
	for _, p := range perms {
		selected[p] = true
	}

	dto := models.CustomRoleDto{
		Name:        role.Name,
		Description: role.Description,
		Permissions: perms,
		IsActive:    role.IsActive,
	}

	data := app.newTemplateData(r)
	data.Form = dto
	data.Data = map[string]interface{}{
		"permGroups":    permissions.PermissionGroups,
		"selectedPerms": selected,
		"roleID":        id,
		"activeTab":     "roles",
	}
	app.render(w, r, http.StatusOK, "custom_role_form.gohtml", data)
}

// POST /church/settings/roles/{id}
func (app *application) customRoleEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	var dto models.CustomRoleDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Name), "name", "Role name is required")

	if !dto.Valid() {
		selected := make(map[string]bool, len(dto.Permissions))
		for _, p := range dto.Permissions {
			selected[p] = true
		}
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{
			"permGroups":    permissions.PermissionGroups,
			"selectedPerms": selected,
			"roleID":        id,
			"activeTab":     "roles",
		}
		app.render(w, r, http.StatusUnprocessableEntity, "custom_role_form.gohtml", data)
		return
	}

	if _, err := app.customRoleModel.Update(r.Context(), id, &dto); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Custom role updated successfully.")
	http.Redirect(w, r, "/church/settings/roles", http.StatusSeeOther)
}

// POST /church/settings/roles/{id}/delete
func (app *application) customRoleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.customRoleModel.Delete(r.Context(), id); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Custom role deleted.")
	http.Redirect(w, r, "/church/settings/roles", http.StatusSeeOther)
}
