package spendings_mock

import (
	"verni/internal/repositories"
	"verni/internal/repositories/spendings"
)

type RepositoryMock struct {
	AddExpenseImpl         func(id spendings.Expense) repositories.MutationWorkItemWithReturnValue[spendings.ExpenseId]
	RemoveExpenseImpl      func(id spendings.ExpenseId) repositories.MutationWorkItem
	GetExpenseImpl         func(id spendings.ExpenseId) (*spendings.IdentifiableExpense, error)
	GetExpensesBetweenImpl func(counterparty1 spendings.CounterpartyId, counterparty2 spendings.CounterpartyId) ([]spendings.IdentifiableExpense, error)
	GetBalanceImpl         func(counterparty spendings.CounterpartyId) ([]spendings.Balance, error)
}

func (c *RepositoryMock) AddExpense(id spendings.Expense) repositories.MutationWorkItemWithReturnValue[spendings.ExpenseId] {
	return c.AddExpenseImpl(id)
}

func (c *RepositoryMock) RemoveExpense(id spendings.ExpenseId) repositories.MutationWorkItem {
	return c.RemoveExpenseImpl(id)
}

func (c *RepositoryMock) GetExpense(id spendings.ExpenseId) (*spendings.IdentifiableExpense, error) {
	return c.GetExpenseImpl(id)
}

func (c *RepositoryMock) GetExpensesBetween(counterparty1 spendings.CounterpartyId, counterparty2 spendings.CounterpartyId) ([]spendings.IdentifiableExpense, error) {
	return c.GetExpensesBetweenImpl(counterparty1, counterparty2)
}

func (c *RepositoryMock) GetBalance(counterparty spendings.CounterpartyId) ([]spendings.Balance, error) {
	return c.GetBalanceImpl(counterparty)
}
