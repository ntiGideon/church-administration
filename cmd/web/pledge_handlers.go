package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /members/{id}/giving
func (app *application) memberGiving(w http.ResponseWriter, r *http.Request) {
	memberID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || memberID < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	member, err := app.memberModel.GetContactByID(r.Context(), memberID)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	givingRecords, err := app.financeModel.ListByContact(r.Context(), memberID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	pledges, err := app.pledgeModel.ListByContact(r.Context(), memberID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	summary := app.financeModel.SumByContact(r.Context(), memberID)

	// Compute fulfillment for each pledge
	type pledgeRow struct {
		ID        int
		Title     string
		Category  string
		Amount    float64
		Currency  string
		StartDate time.Time
		EndDate   string
		Frequency string
		Notes     string
		Fulfilled float64
		Percent   int
	}
	var pledgeRows []pledgeRow
	for _, p := range pledges {
		var fulfilled float64
		for _, f := range givingRecords {
			if f.TransactionDate.Before(p.StartDate) {
				continue
			}
			if !p.EndDate.IsZero() && f.TransactionDate.After(p.EndDate) {
				continue
			}
			fulfilled += f.Amount
		}
		pct := 0
		if p.Amount > 0 {
			pct = int((fulfilled / p.Amount) * 100)
			if pct > 100 {
				pct = 100
			}
		}
		endStr := ""
		if !p.EndDate.IsZero() {
			endStr = p.EndDate.Format("Jan 2, 2006")
		}
		pledgeRows = append(pledgeRows, pledgeRow{
			ID:        p.ID,
			Title:     p.Title,
			Category:  p.Category,
			Amount:    p.Amount,
			Currency:  p.Currency,
			StartDate: p.StartDate,
			EndDate:   endStr,
			Frequency: p.Frequency.String(),
			Notes:     p.Notes,
			Fulfilled: fulfilled,
			Percent:   pct,
		})
	}

	data := app.newTemplateData(r)
	data.Form = models.PledgeDto{}
	data.Data = map[string]interface{}{
		"member":        member,
		"givingRecords": givingRecords,
		"pledges":       pledgeRows,
		"summary":       summary,
	}
	app.render(w, r, http.StatusOK, "member_giving.gohtml", data)
}

// POST /members/{id}/pledges/new
func (app *application) memberPledgeNewPost(w http.ResponseWriter, r *http.Request) {
	memberID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || memberID < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	var dto models.PledgeDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Title), "title", "Title is required")
	dto.CheckField(dto.Amount > 0, "amount", "Amount must be greater than zero")
	dto.CheckField(validator.NotBlank(dto.StartDate), "start_date", "Start date is required")

	if !dto.Valid() {
		app.sessionManager.Put(r.Context(), "flash_error", "Please fix the form errors.")
		http.Redirect(w, r, "/members/"+strconv.Itoa(memberID)+"/giving", http.StatusSeeOther)
		return
	}

	cid := app.churchID(r)
	if _, err := app.pledgeModel.Create(r.Context(), &dto, memberID, cid); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Pledge recorded successfully!")
	http.Redirect(w, r, "/members/"+strconv.Itoa(memberID)+"/giving", http.StatusSeeOther)
}

// POST /members/{id}/pledges/{pid}/delete
func (app *application) memberPledgeDelete(w http.ResponseWriter, r *http.Request) {
	memberID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || memberID < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	pid, err := strconv.Atoi(r.PathValue("pid"))
	if err != nil || pid < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.pledgeModel.Delete(r.Context(), pid); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Pledge deleted.")
	http.Redirect(w, r, "/members/"+strconv.Itoa(memberID)+"/giving", http.StatusSeeOther)
}
