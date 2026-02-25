package main

import (
	"net/http"

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
