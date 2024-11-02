package spendings

import (
	"verni/internal/common"
	spendingsRepository "verni/internal/repositories/spendings"
	"verni/internal/services/pushNotifications"
)

type CounterpartyId spendingsRepository.CounterpartyId
type ExpenseId spendingsRepository.ExpenseId
type Expense spendingsRepository.Expense
type IdentifiableExpense spendingsRepository.IdentifiableExpense
type Balance spendingsRepository.Balance
type Repository spendingsRepository.Repository

type Controller interface {
	AddExpense(expense Expense, actor CounterpartyId) *common.CodeBasedError[AddExpenseErrorCode]
	RemoveExpense(expenseId ExpenseId, actor CounterpartyId) (IdentifiableExpense, *common.CodeBasedError[RemoveExpenseErrorCode])
	GetExpense(expenseId ExpenseId, actor CounterpartyId) (IdentifiableExpense, *common.CodeBasedError[GetExpenseErrorCode])
	GetExpensesWith(counterparty CounterpartyId, actor CounterpartyId) ([]IdentifiableExpense, *common.CodeBasedError[GetExpensesErrorCode])
	GetBalance(actor CounterpartyId) ([]Balance, *common.CodeBasedError[GetBalanceErrorCode])
}

func DefaultController(repository Repository, pushNotifications pushNotifications.Service) Controller {
	return &defaultController{
		repository:        repository,
		pushNotifications: pushNotifications,
	}
}
