package main

import (
	"net/http"

	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /giving/donations  (finance list)
func (app *application) financeList(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	churchID := 0
	if u != nil && u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}

	transactions, err := app.financeModel.List(r.Context(), churchID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	summary, _ := app.financeModel.Summary(r.Context(), churchID)

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"transactions": transactions,
		"summary":      summary,
	}
	app.render(w, r, http.StatusOK, "finance.gohtml", data)
}

// GET /giving/new
func (app *application) financeNewGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = models.FinanceDto{}
	app.render(w, r, http.StatusOK, "finance_new.gohtml", data)
}

// POST /giving/new
func (app *application) financeNewPost(w http.ResponseWriter, r *http.Request) {
	var dto models.FinanceDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Description), "description", "Description is required")
	dto.CheckField(validator.NotBlank(dto.TransactionType), "transaction_type", "Transaction type is required")
	dto.CheckField(dto.Amount > 0, "amount", "Amount must be greater than zero")
	dto.CheckField(validator.NotBlank(dto.Category), "category", "Category is required")
	dto.CheckField(validator.NotBlank(dto.TransactionDate), "transaction_date", "Date is required")

	if !dto.Valid() {
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "finance_new.gohtml", data)
		return
	}

	u := app.getAuthenticatedUser(r)
	churchID := 0
	userID := 0
	if u != nil {
		userID = u.ID
		if u.Edges.Church != nil {
			churchID = u.Edges.Church.ID
		}
	}

	_, err := app.financeModel.Create(r.Context(), &dto, churchID, userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Transaction recorded successfully!")
	http.Redirect(w, r, "/giving/donations", http.StatusSeeOther)
}
