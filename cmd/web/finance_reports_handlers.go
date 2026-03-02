package main

import (
	"context"
	"net/http"
	"time"

	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/finance"
)

// priorYearResult holds totals for the prior calendar year.
type priorYearResult struct {
	TotalIncome   float64
	TotalExpenses float64
	NetBalance    float64
}

// priorYearSummary computes income/expense totals for the previous calendar year.
func (app *application) priorYearSummary(ctx context.Context, churchID int) priorYearResult {
	priorYear := time.Now().Year() - 1
	start := time.Date(priorYear, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(priorYear, 12, 31, 23, 59, 59, 0, time.UTC)

	q := app.db.Finance.Query().
		Where(
			finance.TransactionDateGTE(start),
			finance.TransactionDateLTE(end),
		)
	if churchID > 0 {
		q = q.Where(finance.HasChurchWith(church.IDEQ(churchID)))
	}

	records, err := q.All(ctx)
	if err != nil {
		return priorYearResult{}
	}

	var result priorYearResult
	for _, r := range records {
		switch r.TransactionType {
		case finance.TransactionTypeDonation, finance.TransactionTypeTithe, finance.TransactionTypeOffering:
			result.TotalIncome += r.Amount
		case finance.TransactionTypeExpense, finance.TransactionTypeSalary:
			result.TotalExpenses += r.Amount
		}
	}
	result.NetBalance = result.TotalIncome - result.TotalExpenses
	return result
}

// GET /reports/finance
func (app *application) financeReports(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cid := app.churchID(r)

	// Monthly trend for the last 12 months
	trend, _ := app.financeModel.MonthlyTrend(ctx, cid, 12)

	// Income category breakdown (all time)
	breakdown, _ := app.financeModel.IncomeCategoryBreakdown(ctx, cid)

	// Current year summary (all time totals)
	summary, _ := app.financeModel.Summary(ctx, cid)

	// Prior year summary
	prior := app.priorYearSummary(ctx, cid)

	// Build Chart.js-friendly slices from trend data
	months := make([]string, len(trend))
	incomeData := make([]float64, len(trend))
	expenseData := make([]float64, len(trend))
	for i, t := range trend {
		months[i] = t.Month
		incomeData[i] = t.Income
		expenseData[i] = t.Expenses
	}

	// Build category breakdown slices
	catLabels := make([]string, len(breakdown))
	catAmounts := make([]float64, len(breakdown))
	for i, c := range breakdown {
		catLabels[i] = c.Label
		catAmounts[i] = c.Amount
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"months":      months,
		"incomeData":  incomeData,
		"expenseData": expenseData,
		"catLabels":   catLabels,
		"catAmounts":  catAmounts,
		"summary":     summary,
		"priorYear":   prior,
		"currentYear": time.Now().Year(),
		"priorYearN":  time.Now().Year() - 1,
	}
	app.render(w, r, http.StatusOK, "finance_reports.gohtml", data)
}
