package main

import (
	"math"
	"net/http"
	"strconv"

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

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}

	const pageSize = 10
	filter := models.FinanceFilter{
		TxType:   q.Get("type"),
		Category: q.Get("category"),
		DateFrom: q.Get("date_from"),
		DateTo:   q.Get("date_to"),
		Page:     page,
		PageSize: pageSize,
	}

	transactions, total, err := app.financeModel.ListFiltered(r.Context(), churchID, filter)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	summary, _ := app.financeModel.Summary(r.Context(), churchID)

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages < 1 {
		totalPages = 1
	}

	fromIdx := (page-1)*pageSize + 1
	toIdx := page * pageSize
	if toIdx > total {
		toIdx = total
	}
	if total == 0 {
		fromIdx = 0
	}

	data := app.newTemplateData(r)
	pageNums := paginationWindow(page, totalPages)
	firstPN, lastPN := 0, 0
	if len(pageNums) > 0 {
		firstPN = pageNums[0]
		lastPN = pageNums[len(pageNums)-1]
	}

	data.Data = map[string]interface{}{
		"transactions":      transactions,
		"summary":           summary,
		"filter":            filter,
		"total":             total,
		"page":              page,
		"totalPages":        totalPages,
		"totalPagesMinus1":  totalPages - 1,
		"pageSize":          pageSize,
		"pageNums":          pageNums,
		"firstPageNum":      firstPN,
		"lastPageNum":       lastPN,
		"fromIdx":           fromIdx,
		"toIdx":             toIdx,
	}
	app.render(w, r, http.StatusOK, "finance.gohtml", data)
}

// paginationWindow returns a slice of page numbers centered around the current page.
func paginationWindow(page, totalPages int) []int {
	const window = 5
	start := page - window/2
	if start < 1 {
		start = 1
	}
	end := start + window - 1
	if end > totalPages {
		end = totalPages
		start = end - window + 1
		if start < 1 {
			start = 1
		}
	}
	nums := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		nums = append(nums, i)
	}
	return nums
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
