package models

import (
	"context"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/contact"
	"github.com/ntiGideon/ent/finance"
)

type FinanceModel struct {
	Db *ent.Client
}

type FinanceSummary struct {
	TotalIncome       float64
	TotalExpenses     float64
	NetBalance        float64
	ThisMonthIncome   float64
	ThisMonthExpenses float64
}

// FinanceFilter holds search/filter parameters for paginated finance listing.
type FinanceFilter struct {
	TxType   string
	Category string
	DateFrom string
	DateTo   string
	Page     int
	PageSize int
}

// Create records a new financial transaction.
func (m *FinanceModel) Create(ctx context.Context, dto *FinanceDto, churchID, userID int) (*ent.Finance, error) {
	txDate, err := time.Parse("2006-01-02", dto.TransactionDate)
	if err != nil {
		txDate = time.Now()
	}

	currency := dto.Currency
	if currency == "" {
		currency = "GHS"
	}

	fb := m.Db.Finance.Create().
		SetDescription(dto.Description).
		SetTransactionType(finance.TransactionType(dto.TransactionType)).
		SetAmount(dto.Amount).
		SetCurrency(currency).
		SetTransactionDate(txDate).
		SetCategory(dto.Category).
		SetNillablePaymentMethod(nullStr(dto.PaymentMethod)).
		SetNillableNotes(nullStr(dto.Notes)).
		SetChurchID(churchID)

	if userID > 0 {
		fb = fb.SetRecordedByID(userID)
	}
	if dto.ContactID > 0 {
		fb = fb.SetContactID(dto.ContactID)
	}

	f, err := fb.Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return f, nil
}

// List returns all finance records for a church, newest first.
// If churchID is 0, returns all records across all churches (super_admin view).
func (m *FinanceModel) List(ctx context.Context, churchID int) ([]*ent.Finance, error) {
	q := m.Db.Finance.Query()
	if churchID > 0 {
		q = q.Where(finance.HasChurchWith(church.IDEQ(churchID)))
	} else {
		q = q.Where(finance.HasChurchWith())
	}
	return q.Order(ent.Desc(finance.FieldTransactionDate)).
		WithRecordedBy(func(uq *ent.UserQuery) {
			uq.WithContact()
		}).
		WithDonor().
		All(ctx)
}

// ListFiltered returns filtered, paginated finance records and the total matching count.
// If churchID is 0, operates across all churches (super_admin view).
func (m *FinanceModel) ListFiltered(ctx context.Context, churchID int, f FinanceFilter) ([]*ent.Finance, int, error) {
	if f.PageSize <= 0 {
		f.PageSize = 10
	}
	if f.Page < 1 {
		f.Page = 1
	}

	q := m.Db.Finance.Query()
	if churchID > 0 {
		q = q.Where(finance.HasChurchWith(church.IDEQ(churchID)))
	} else {
		q = q.Where(finance.HasChurchWith())
	}

	if f.TxType != "" {
		q = q.Where(finance.TransactionTypeEQ(finance.TransactionType(f.TxType)))
	}
	if f.Category != "" {
		q = q.Where(finance.CategoryEQ(f.Category))
	}
	if f.DateFrom != "" {
		if t, err := time.Parse("2006-01-02", f.DateFrom); err == nil {
			q = q.Where(finance.TransactionDateGTE(t))
		}
	}
	if f.DateTo != "" {
		if t, err := time.Parse("2006-01-02", f.DateTo); err == nil {
			q = q.Where(finance.TransactionDateLT(t.Add(24 * time.Hour)))
		}
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	records, err := q.Order(ent.Desc(finance.FieldTransactionDate)).
		Limit(f.PageSize).
		Offset((f.Page - 1) * f.PageSize).
		WithRecordedBy(func(uq *ent.UserQuery) { uq.WithContact() }).
		WithDonor().
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// Summary returns aggregate income / expense totals for a church.
// If churchID is 0, summarises all records across all churches (super_admin view).
func (m *FinanceModel) Summary(ctx context.Context, churchID int) (*FinanceSummary, error) {
	q := m.Db.Finance.Query()
	if churchID > 0 {
		q = q.Where(finance.HasChurchWith(church.IDEQ(churchID)))
	} else {
		q = q.Where(finance.HasChurchWith())
	}
	records, err := q.All(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	s := &FinanceSummary{}
	for _, r := range records {
		switch r.TransactionType {
		case finance.TransactionTypeDonation, finance.TransactionTypeTithe, finance.TransactionTypeOffering:
			s.TotalIncome += r.Amount
			if r.TransactionDate.After(monthStart) {
				s.ThisMonthIncome += r.Amount
			}
		case finance.TransactionTypeExpense, finance.TransactionTypeSalary:
			s.TotalExpenses += r.Amount
			if r.TransactionDate.After(monthStart) {
				s.ThisMonthExpenses += r.Amount
			}
		}
	}
	s.NetBalance = s.TotalIncome - s.TotalExpenses
	return s, nil
}

// MonthlyTrendItem holds aggregated income and expense totals for one calendar month.
type MonthlyTrendItem struct {
	Month    string  `json:"month"`
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
}

// MonthlyTrend returns income and expense totals per calendar month for the last
// n months (inclusive of the current month), in chronological order.
// If churchID is 0, aggregates across all churches.
func (m *FinanceModel) MonthlyTrend(ctx context.Context, churchID, months int) ([]MonthlyTrendItem, error) {
	if months <= 0 {
		months = 6
	}

	now := time.Now()
	type slotKey struct {
		year  int
		month time.Month
	}

	result := make([]MonthlyTrendItem, months)
	slotMap := make(map[slotKey]int, months)

	for i := 0; i < months; i++ {
		t := now.AddDate(0, -(months-1-i), 0)
		k := slotKey{t.Year(), t.Month()}
		label := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC).Format("Jan '06")
		result[i] = MonthlyTrendItem{Month: label}
		slotMap[k] = i
	}

	earliestT := now.AddDate(0, -(months-1), 0)
	earliest := time.Date(earliestT.Year(), earliestT.Month(), 1, 0, 0, 0, 0, time.UTC)

	q := m.Db.Finance.Query().Where(finance.TransactionDateGTE(earliest))
	if churchID > 0 {
		q = q.Where(finance.HasChurchWith(church.IDEQ(churchID)))
	} else {
		q = q.Where(finance.HasChurchWith())
	}

	records, err := q.All(ctx)
	if err != nil {
		return result, nil // return zeroed slots on error
	}

	for _, r := range records {
		k := slotKey{r.TransactionDate.Year(), r.TransactionDate.Month()}
		idx, ok := slotMap[k]
		if !ok {
			continue
		}
		switch r.TransactionType {
		case finance.TransactionTypeDonation, finance.TransactionTypeTithe, finance.TransactionTypeOffering:
			result[idx].Income += r.Amount
		case finance.TransactionTypeExpense, finance.TransactionTypeSalary:
			result[idx].Expenses += r.Amount
		}
	}
	return result, nil
}

// CategoryTotal holds a finance type label and its total amount.
type CategoryTotal struct {
	Label  string  `json:"label"`
	Amount float64 `json:"amount"`
}

// IncomeCategoryBreakdown returns per-type income totals (tithe, offering, donation).
// If churchID is 0, aggregates across all churches.
func (m *FinanceModel) IncomeCategoryBreakdown(ctx context.Context, churchID int) ([]CategoryTotal, error) {
	q := m.Db.Finance.Query()
	if churchID > 0 {
		q = q.Where(finance.HasChurchWith(church.IDEQ(churchID)))
	} else {
		q = q.Where(finance.HasChurchWith())
	}

	records, err := q.All(ctx)
	if err != nil {
		return nil, err
	}

	totals := map[string]float64{"Tithe": 0, "Offering": 0, "Donation": 0}
	for _, r := range records {
		switch r.TransactionType {
		case finance.TransactionTypeTithe:
			totals["Tithe"] += r.Amount
		case finance.TransactionTypeOffering:
			totals["Offering"] += r.Amount
		case finance.TransactionTypeDonation:
			totals["Donation"] += r.Amount
		}
	}
	return []CategoryTotal{
		{Label: "Tithe", Amount: totals["Tithe"]},
		{Label: "Offering", Amount: totals["Offering"]},
		{Label: "Donation", Amount: totals["Donation"]},
	}, nil
}

// MemberGivingSummary holds giving stats for a single member.
type MemberGivingSummary struct {
	Total     float64
	ThisYear  float64
	ThisMonth float64
}

// ListByContact returns all income-type finance records linked to a specific contact, newest first.
func (m *FinanceModel) ListByContact(ctx context.Context, contactID int) ([]*ent.Finance, error) {
	return m.Db.Finance.Query().
		Where(
			finance.HasDonorWith(contact.IDEQ(contactID)),
		).
		Order(ent.Desc(finance.FieldTransactionDate)).
		All(ctx)
}

// SumByContact returns giving totals (lifetime, this year, this month) for a contact.
func (m *FinanceModel) SumByContact(ctx context.Context, contactID int) MemberGivingSummary {
	records, err := m.Db.Finance.Query().
		Where(
			finance.HasDonorWith(contact.IDEQ(contactID)),
			finance.TransactionTypeIn(
				finance.TransactionTypeTithe,
				finance.TransactionTypeOffering,
				finance.TransactionTypeDonation,
			),
		).
		All(ctx)
	if err != nil {
		return MemberGivingSummary{}
	}
	now := time.Now()
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	var s MemberGivingSummary
	for _, r := range records {
		s.Total += r.Amount
		if r.TransactionDate.After(yearStart) {
			s.ThisYear += r.Amount
		}
		if r.TransactionDate.After(monthStart) {
			s.ThisMonth += r.Amount
		}
	}
	return s
}

// RecentTransactions returns the most recent N transactions.
// If churchID is 0, returns across all churches (super_admin view).
func (m *FinanceModel) RecentTransactions(ctx context.Context, churchID, limit int) ([]*ent.Finance, error) {
	q := m.Db.Finance.Query()
	if churchID > 0 {
		q = q.Where(finance.HasChurchWith(church.IDEQ(churchID)))
	} else {
		q = q.Where(finance.HasChurchWith())
	}
	return q.Order(ent.Desc(finance.FieldTransactionDate)).
		Limit(limit).
		All(ctx)
}
