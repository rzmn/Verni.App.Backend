package spendings

import (
	"github.com/rzmn/governi/internal/repositories"
)

type ExpenseId string
type CounterpartyId string
type Currency string
type Cost int64

type ShareOfExpense struct {
	Counterparty CounterpartyId
	Cost         Cost
}

type Expense struct {
	Timestamp int64
	Details   string
	Total     Cost
	Currency  Currency
	Shares    []ShareOfExpense
}

type IdentifiableExpense struct {
	Expense
	Id ExpenseId
}

type Balance struct {
	Counterparty CounterpartyId
	Currencies   map[Currency]Cost
}

type Repository interface {
	AddExpense(id Expense) repositories.MutationWorkItemWithReturnValue[ExpenseId]
	RemoveExpense(id ExpenseId) repositories.MutationWorkItem

	GetExpense(id ExpenseId) (*IdentifiableExpense, error)

	GetExpensesBetween(counterparty1 CounterpartyId, counterparty2 CounterpartyId) ([]IdentifiableExpense, error)
	GetBalance(counterparty CounterpartyId) ([]Balance, error)
}
