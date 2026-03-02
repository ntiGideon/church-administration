package models

import (
	"context"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/budget"
	"github.com/ntiGideon/ent/budgetline"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/finance"
)

type BudgetModel struct {
	Db *ent.Client
}

// BudgetVsActual holds a single line item's allocated vs actual comparison.
type BudgetVsActual struct {
	LineID    int
	Category  string
	LineType  string
	Allocated float64
	Actual    float64
	Variance  float64
	UsedPct   float64
	Currency  string
}

// BudgetSummary holds the full comparison for a budget.
type BudgetSummary struct {
	Budget                *ent.Budget
	Lines                 []BudgetVsActual
	TotalAllocatedIncome  float64
	TotalAllocatedExpense float64
	TotalActualIncome     float64
	TotalActualExpense    float64
}

// Create saves a new budget for a church.
func (m *BudgetModel) Create(ctx context.Context, dto *BudgetDto, churchID int) (*ent.Budget, error) {
	startDate, err := time.Parse("2006-01-02", dto.StartDate)
	if err != nil {
		startDate = time.Now()
	}
	endDate, err := time.Parse("2006-01-02", dto.EndDate)
	if err != nil {
		endDate = time.Now().AddDate(1, 0, 0)
	}

	status := budget.Status(dto.Status)
	if dto.Status == "" {
		status = budget.StatusDraft
	}
	period := budget.Period(dto.Period)
	if dto.Period == "" {
		period = budget.PeriodAnnual
	}

	b, err := m.Db.Budget.Create().
		SetName(dto.Name).
		SetFiscalYear(dto.FiscalYear).
		SetPeriod(period).
		SetStartDate(startDate).
		SetEndDate(endDate).
		SetStatus(status).
		SetNillableNotes(nullStr(dto.Notes)).
		SetChurchID(churchID).
		Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return b, nil
}

// ListByChurch returns all budgets for a church with their lines, newest first.
func (m *BudgetModel) ListByChurch(ctx context.Context, churchID int) ([]*ent.Budget, error) {
	q := m.Db.Budget.Query().WithLines()
	if churchID > 0 {
		q = q.Where(budget.HasChurchWith(church.IDEQ(churchID)))
	}
	return q.Order(ent.Desc(budget.FieldFiscalYear), ent.Desc(budget.FieldCreatedAt)).All(ctx)
}

// GetByID returns a budget with its lines and church loaded.
func (m *BudgetModel) GetByID(ctx context.Context, id int) (*ent.Budget, error) {
	b, err := m.Db.Budget.Query().
		Where(budget.IDEQ(id)).
		WithLines().
		WithChurch().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return b, nil
}

// Update saves changes to a budget's metadata fields.
func (m *BudgetModel) Update(ctx context.Context, id int, dto *BudgetDto) (*ent.Budget, error) {
	startDate, err := time.Parse("2006-01-02", dto.StartDate)
	if err != nil {
		startDate = time.Now()
	}
	endDate, err := time.Parse("2006-01-02", dto.EndDate)
	if err != nil {
		endDate = time.Now().AddDate(1, 0, 0)
	}

	b, err := m.Db.Budget.UpdateOneID(id).
		SetName(dto.Name).
		SetFiscalYear(dto.FiscalYear).
		SetPeriod(budget.Period(dto.Period)).
		SetStartDate(startDate).
		SetEndDate(endDate).
		SetStatus(budget.Status(dto.Status)).
		SetNillableNotes(nullStr(dto.Notes)).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Delete removes a budget and its lines.
func (m *BudgetModel) Delete(ctx context.Context, id int) error {
	return m.Db.Budget.DeleteOneID(id).Exec(ctx)
}

// AddLine adds a new line item to a budget.
func (m *BudgetModel) AddLine(ctx context.Context, budgetID int, dto *BudgetLineDto) (*ent.BudgetLine, error) {
	currency := dto.Currency
	if currency == "" {
		currency = "GHS"
	}
	lineType := budgetline.LineType(dto.LineType)
	if dto.LineType == "" {
		lineType = budgetline.LineTypeIncome
	}

	bl, err := m.Db.BudgetLine.Create().
		SetCategory(dto.Category).
		SetLineType(lineType).
		SetAllocatedAmount(dto.AllocatedAmount).
		SetCurrency(currency).
		SetNillableNotes(nullStr(dto.Notes)).
		SetBudgetID(budgetID).
		Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return bl, nil
}

// DeleteLine removes a budget line item.
func (m *BudgetModel) DeleteLine(ctx context.Context, lineID int) error {
	return m.Db.BudgetLine.DeleteOneID(lineID).Exec(ctx)
}

// GetVsActual computes allocated vs actual for each line in a budget.
func (m *BudgetModel) GetVsActual(ctx context.Context, budgetID int) (*BudgetSummary, error) {
	b, err := m.GetByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	churchID := b.ChurchID

	summary := &BudgetSummary{Budget: b}

	for _, line := range b.Edges.Lines {
		// Query Finance records matching this line's category and the budget date range.
		records, err := m.Db.Finance.Query().
			Where(
				finance.HasChurchWith(church.IDEQ(churchID)),
				finance.CategoryEQ(line.Category),
				finance.TransactionDateGTE(b.StartDate),
				finance.TransactionDateLTE(b.EndDate),
			).
			All(ctx)
		if err != nil {
			records = nil
		}

		var actual float64
		for _, r := range records {
			actual += r.Amount
		}

		variance := line.AllocatedAmount - actual
		var usedPct float64
		if line.AllocatedAmount > 0 {
			usedPct = (actual / line.AllocatedAmount) * 100
		}

		va := BudgetVsActual{
			LineID:    line.ID,
			Category:  line.Category,
			LineType:  line.LineType.String(),
			Allocated: line.AllocatedAmount,
			Actual:    actual,
			Variance:  variance,
			UsedPct:   usedPct,
			Currency:  line.Currency,
		}
		summary.Lines = append(summary.Lines, va)

		if line.LineType.String() == "income" {
			summary.TotalAllocatedIncome += line.AllocatedAmount
			summary.TotalActualIncome += actual
		} else {
			summary.TotalAllocatedExpense += line.AllocatedAmount
			summary.TotalActualExpense += actual
		}
	}

	return summary, nil
}
