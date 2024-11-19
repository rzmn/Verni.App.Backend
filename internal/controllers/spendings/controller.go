package spendings

import (
	"github.com/rzmn/governi/internal/common"
	spendingsRepository "github.com/rzmn/governi/internal/repositories/spendings"
)

type CounterpartyId spendingsRepository.CounterpartyId
type ExpenseId spendingsRepository.ExpenseId
type Expense spendingsRepository.Expense
type IdentifiableExpense spendingsRepository.IdentifiableExpense
type Balance spendingsRepository.Balance

type Controller interface {
	AddExpense(expense Expense, actor CounterpartyId) (IdentifiableExpense, *common.CodeBasedError[AddExpenseErrorCode])
	RemoveExpense(expenseId ExpenseId, actor CounterpartyId) (IdentifiableExpense, *common.CodeBasedError[RemoveExpenseErrorCode])
	GetExpense(expenseId ExpenseId, actor CounterpartyId) (IdentifiableExpense, *common.CodeBasedError[GetExpenseErrorCode])
	GetExpensesWith(counterparty CounterpartyId, actor CounterpartyId) ([]IdentifiableExpense, *common.CodeBasedError[GetExpensesErrorCode])
	GetBalance(actor CounterpartyId) ([]Balance, *common.CodeBasedError[GetBalanceErrorCode])
}
