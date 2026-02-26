package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /documents
func (app *application) documentsList(w http.ResponseWriter, r *http.Request) {
	cid := app.churchID(r)
	docs, err := app.documentModel.ListByChurch(r.Context(), cid)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Count per category
	counts := map[string]int{"all": len(docs)}
	for _, d := range docs {
		counts[string(d.Category)]++
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"documents": docs,
		"counts":    counts,
	}
	app.render(w, r, http.StatusOK, "documents.gohtml", data)
}

// GET /documents/upload
func (app *application) documentUploadGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = models.DocumentDto{Category: "other", IsPublic: true}
	app.render(w, r, http.StatusOK, "document_upload.gohtml", data)
}

// POST /documents/upload
func (app *application) documentUploadPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxDocumentSize); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	var dto models.DocumentDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Title), "title", "Title is required")
	dto.CheckField(validator.NotBlank(dto.Category), "category", "Please select a category")

	file, header, fileErr := r.FormFile("document_file")
	if fileErr != nil {
		dto.AddFieldError("document_file", "Please choose a file to upload")
	}

	if !dto.Valid() {
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "document_upload.gohtml", data)
		return
	}
	defer file.Close()

	fileURL, mimeType, err := app.uploadDocument(file, header, "documents")
	if err != nil {
		dto.AddFieldError("document_file", fmt.Sprintf("Upload failed: %s", err.Error()))
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "document_upload.gohtml", data)
		return
	}

	cid := app.churchID(r)
	_, err = app.documentModel.Create(r.Context(), &dto, fileURL, header.Filename, mimeType, header.Size, cid)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Document uploaded successfully!")
	http.Redirect(w, r, "/documents", http.StatusSeeOther)
}

// POST /documents/{id}/delete
func (app *application) documentDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.documentModel.Delete(r.Context(), id); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Document deleted.")
	http.Redirect(w, r, "/documents", http.StatusSeeOther)
}
