package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ntiGideon/ent/user"
	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /announcements
func (app *application) announcementsList(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	churchID := 0
	if u != nil && u.Role != user.RoleSuperAdmin && u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}

	announcements, err := app.announcementModel.ListByChurch(r.Context(), churchID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"announcements": announcements,
	}
	app.render(w, r, http.StatusOK, "announcements.gohtml", data)
}

// GET /announcements/new
func (app *application) announcementNewGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = models.AnnouncementDto{}
	app.render(w, r, http.StatusOK, "announcement_new.gohtml", data)
}

// POST /announcements/new
func (app *application) announcementNewPost(w http.ResponseWriter, r *http.Request) {
	var dto models.AnnouncementDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Title), "title", "Title is required")
	dto.CheckField(validator.NotBlank(dto.Content), "content", "Content is required")
	dto.CheckField(validator.NotBlank(dto.Category), "category", "Category is required")

	if !dto.Valid() {
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "announcement_new.gohtml", data)
		return
	}

	u := app.getAuthenticatedUser(r)
	churchID := 0
	authorID := 0
	if u != nil {
		authorID = u.ID
		if u.Edges.Church != nil {
			churchID = u.Edges.Church.ID
		}
	}

	_, err := app.announcementModel.Create(r.Context(), &dto, churchID, authorID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Announcement published successfully.")
	http.Redirect(w, r, "/announcements", http.StatusSeeOther)
}

// GET /announcements/{id}
func (app *application) announcementDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	ann, err := app.announcementModel.GetByID(r.Context(), id)
	if err != nil {
		if err == models.ErrRecordNotFound {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{"announcement": ann}
	app.render(w, r, http.StatusOK, "announcement_detail.gohtml", data)
}

// GET /announcements/{id}/edit
func (app *application) announcementEditGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	ann, err := app.announcementModel.GetByID(r.Context(), id)
	if err != nil {
		if err == models.ErrRecordNotFound {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	dto := models.AnnouncementDto{
		Title:       ann.Title,
		Content:     ann.Content,
		Category:    string(ann.Category),
		IsPublished: ann.IsPublished,
	}

	data := app.newTemplateData(r)
	data.Form = dto
	data.Data = map[string]interface{}{"announcement": ann}
	app.render(w, r, http.StatusOK, "announcement_edit.gohtml", data)
}

// POST /announcements/{id}/edit
func (app *application) announcementEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	ann, err := app.announcementModel.GetByID(r.Context(), id)
	if err != nil {
		if err == models.ErrRecordNotFound {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	var dto models.AnnouncementDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Title), "title", "Title is required")
	dto.CheckField(validator.NotBlank(dto.Content), "content", "Content is required")
	dto.CheckField(validator.NotBlank(dto.Category), "category", "Category is required")

	if !dto.Valid() {
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{"announcement": ann}
		app.render(w, r, http.StatusUnprocessableEntity, "announcement_edit.gohtml", data)
		return
	}

	if err := app.announcementModel.Update(r.Context(), id, &dto); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Announcement updated successfully.")
	http.Redirect(w, r, fmt.Sprintf("/announcements/%d", id), http.StatusSeeOther)
}

// POST /announcements/{id}/publish  — toggles published/draft state
func (app *application) announcementTogglePublish(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	ann, err := app.announcementModel.GetByID(r.Context(), id)
	if err != nil {
		if err == models.ErrRecordNotFound {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	if ann.IsPublished {
		if err := app.announcementModel.Unpublish(r.Context(), id); err != nil {
			app.serverError(w, r, err)
			return
		}
		app.sessionManager.Put(r.Context(), "flash", "Announcement moved to draft.")
	} else {
		if err := app.announcementModel.Publish(r.Context(), id); err != nil {
			app.serverError(w, r, err)
			return
		}
		app.sessionManager.Put(r.Context(), "flash", "Announcement published.")
	}

	http.Redirect(w, r, fmt.Sprintf("/announcements/%d", id), http.StatusSeeOther)
}

// POST /announcements/{id}/delete
func (app *application) announcementDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.announcementModel.Delete(r.Context(), id); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Announcement deleted.")
	http.Redirect(w, r, "/announcements", http.StatusSeeOther)
}
