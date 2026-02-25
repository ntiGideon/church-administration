package models

import (
	"context"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/finance"
)

type FinanceModel struct {
	Db *ent.Client
}

type FinanceSummary struct {
	TotalIncome   float64
	TotalExpenses float64
	NetBalance    float64
	ThisMonthIncome   float64
	ThisMonthExpenses float64
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

	f, err := fb.Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return f, nil
}

// List returns all finance records for a church, newest first.
func (m *FinanceModel) List(ctx context.Context, churchID int) ([]*ent.Finance, error) {
	return m.Db.Finance.Query().
		Where(finance.HasChurchWith()).
		Order(ent.Desc(finance.FieldTransactionDate)).
		WithRecordedBy(func(uq *ent.UserQuery) {
			uq.WithContact()
		}).
		All(ctx)
}

// Summary returns aggregate income / expense totals for a church.
func (m *FinanceModel) Summary(ctx context.Context, churchID int) (*FinanceSummary, error) {
	records, err := m.Db.Finance.Query().
		Where(finance.HasChurchWith()).
		All(ctx)
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

// RecentTransactions returns the most recent N transactions.
func (m *FinanceModel) RecentTransactions(ctx context.Context, churchID, limit int) ([]*ent.Finance, error) {
	return m.Db.Finance.Query().
		Where(finance.HasChurchWith()).
		Order(ent.Desc(finance.FieldTransactionDate)).
		Limit(limit).
		All(ctx)
}
