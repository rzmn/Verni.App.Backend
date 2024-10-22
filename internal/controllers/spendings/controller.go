package spendings

import (
	"verni/internal/common"
	"verni/internal/pushNotifications"
	spendingsRepository "verni/internal/repositories/spendings"
)

type CounterpartyId spendingsRepository.CounterpartyId
type ExpenseId spendingsRepository.ExpenseId
type Expense spendingsRepository.Expense
type IdentifiableExpense spendingsRepository.IdentifiableExpense
type Balance spendingsRepository.Balance
type Repository spendingsRepository.Repository

type Controller interface {
	AddExpense(expense Expense, actor CounterpartyId) *common.CodeBasedError[CreateDealErrorCode]
	RemoveExpense(expenseId ExpenseId, actor CounterpartyId) (IdentifiableExpense, *common.CodeBasedError[DeleteDealErrorCode])
	GetExpense(expenseId ExpenseId, actor CounterpartyId) (IdentifiableExpense, *common.CodeBasedError[GetDealErrorCode])
	GetExpensesWith(counterparty CounterpartyId, actor CounterpartyId) ([]IdentifiableExpense, *common.CodeBasedError[GetDealsErrorCode])
	GetBalance(actor CounterpartyId) ([]Balance, *common.CodeBasedError[GetCounterpartiesErrorCode])
}

func DefaultController(repository Repository, pushNotifications pushNotifications.Service) Controller {
	return &defaultController{
		repository:        repository,
		pushNotifications: pushNotifications,
	}
}
