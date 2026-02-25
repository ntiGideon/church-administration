package main

import (
	"net/http"

	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /church/settings — branch admin manages their church profile
func (app *application) churchSettings(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if u.Edges.Church == nil {
		app.clientError(w, http.StatusForbidden)
		return
	}

	c := u.Edges.Church
	dto := models.ChurchSettingsDto{
		Name:             c.Name,
		Address:          c.Address,
		City:             c.City,
		Country:          c.Country,
		Phone:            c.Phone,
		Website:          c.Website,
		CongregationSize: c.CongregationSize,
	}

	memberCount, _ := app.memberModel.CountByChurch(r.Context(), c.ID)
	eventCount, _ := app.db.Event.Query().Count(r.Context())

	offlineCount := c.CongregationSize - memberCount
	if offlineCount < 0 {
		offlineCount = 0
	}

	data := app.newTemplateData(r)
	data.Form = dto
	data.Data = map[string]interface{}{
		"church":       c,
		"memberCount":  memberCount,
		"eventCount":   eventCount,
		"offlineCount": offlineCount,
	}
	app.render(w, r, http.StatusOK, "church_settings.gohtml", data)
}

// POST /church/settings
func (app *application) churchSettingsPost(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if u.Edges.Church == nil {
		app.clientError(w, http.StatusForbidden)
		return
	}

	var dto models.ChurchSettingsDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Name), "name", "Church name is required")
	dto.CheckField(validator.NotBlank(dto.Address), "address", "Address is required")
	dto.CheckField(dto.CongregationSize >= 0, "congregation_size", "Must be 0 or greater")

	if !dto.Valid() {
		memberCount, _ := app.memberModel.CountByChurch(r.Context(), u.Edges.Church.ID)
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{
			"church":      u.Edges.Church,
			"memberCount": memberCount,
		}
		app.render(w, r, http.StatusUnprocessableEntity, "church_settings.gohtml", data)
		return
	}

	if err := app.churchModel.UpdateSettings(r.Context(), u.Edges.Church.ID, &dto); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Church settings updated successfully.")
	http.Redirect(w, r, "/church/settings", http.StatusSeeOther)
}

// POST /church/settings/logo
func (app *application) churchLogoPost(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	if u == nil || u.Edges.Church == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		app.sessionManager.Put(r.Context(), "flash_error", "File too large — maximum size is 5 MB.")
		http.Redirect(w, r, "/church/settings", http.StatusSeeOther)
		return
	}

	file, header, err := r.FormFile("logo")
	if err != nil {
		app.sessionManager.Put(r.Context(), "flash_error", "No file selected.")
		http.Redirect(w, r, "/church/settings", http.StatusSeeOther)
		return
	}
	defer file.Close()

	url, err := app.uploadImage(file, header, "logos")
	if err != nil {
		app.sessionManager.Put(r.Context(), "flash_error", err.Error())
		http.Redirect(w, r, "/church/settings", http.StatusSeeOther)
		return
	}

	if _, err := app.db.Church.UpdateOneID(u.Edges.Church.ID).
		SetLogoURL(url).
		Save(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Church logo updated successfully.")
	http.Redirect(w, r, "/church/settings", http.StatusSeeOther)
}
