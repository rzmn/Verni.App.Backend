package defaultController

import (
	"github.com/rzmn/Verni.App.Backend/internal/common"
	"github.com/rzmn/Verni.App.Backend/internal/controllers/spendings"
	spendingsRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/spendings"
	"github.com/rzmn/Verni.App.Backend/internal/services/logging"
)

type Repository spendingsRepository.Repository

func New(repository Repository, logger logging.Service) spendings.Controller {
	return &defaultController{
		repository: repository,
		logger:     logger,
	}
}

type defaultController struct {
	repository Repository
	logger     logging.Service
}

func (c *defaultController) AddExpense(expense spendings.Expense, actor spendings.CounterpartyId) (spendings.IdentifiableExpense, *common.CodeBasedError[spendings.AddExpenseErrorCode]) {
	const op = "spendings.defaultController.AddExpense"
	c.logger.LogInfo("%s: start[actor=%s]", op, actor)
	var isYourExpense bool
	for i := 0; i < len(expense.Shares); i++ {
		if expense.Shares[i].Counterparty == spendingsRepository.CounterpartyId(actor) {
			isYourExpense = true
			break
		}
	}
	if !isYourExpense {
		c.logger.LogInfo("%s: user %s is not found in expense %v shares", op, actor, expense)
		return spendings.IdentifiableExpense{}, common.NewError(spendings.AddExpenseErrorNotYourExpense)
	}
	transaction := c.repository.AddExpense(spendingsRepository.Expense(expense))
	expenseId, err := transaction.Perform()
	if err != nil {
		c.logger.LogInfo("%s: cannot insert expense into db err: %v", op, err)
		return spendings.IdentifiableExpense{}, common.NewErrorWithDescription(spendings.AddExpenseErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[actor=%s]", op, actor)
	return spendings.IdentifiableExpense{
		Expense: spendingsRepository.Expense(expense),
		Id:      expenseId,
	}, nil
}

func (c *defaultController) RemoveExpense(expenseId spendings.ExpenseId, actor spendings.CounterpartyId) (spendings.IdentifiableExpense, *common.CodeBasedError[spendings.RemoveExpenseErrorCode]) {
	const op = "spendings.defaultController.RemoveExpense"
	c.logger.LogInfo("%s: start[expenseId=%s actor=%s]", op, expenseId, actor)
	expense, err := c.repository.GetExpense(spendingsRepository.ExpenseId(expenseId))
	if err != nil {
		c.logger.LogInfo("%s: cannot get expense from db err: %v", op, err)
		return spendings.IdentifiableExpense{}, common.NewErrorWithDescription(spendings.RemoveExpenseErrorInternal, err.Error())
	}
	if expense == nil {
		c.logger.LogInfo("%s: expense %s does not exists", op, expenseId)
		return spendings.IdentifiableExpense{}, common.NewError(spendings.RemoveExpenseErrorExpenseNotFound)
	}
	var isYourExpense bool
	for i := 0; i < len(expense.Shares); i++ {
		if expense.Shares[i].Counterparty == spendingsRepository.CounterpartyId(actor) {
			isYourExpense = true
			break
		}
	}
	if !isYourExpense {
		c.logger.LogInfo("%s: user %s is not found in expense %s shares", op, actor, expenseId)
		return spendings.IdentifiableExpense{}, common.NewError(spendings.RemoveExpenseErrorNotYourExpense)
	}
	transaction := c.repository.RemoveExpense(spendingsRepository.ExpenseId(expenseId))
	if err := transaction.Perform(); err != nil {
		c.logger.LogInfo("%s: cannot remove expense from db err: %v", op, err)
		return spendings.IdentifiableExpense{}, common.NewErrorWithDescription(spendings.RemoveExpenseErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[expenseId=%s actor=%s]", op, expenseId, actor)
	return spendings.IdentifiableExpense(*expense), nil
}

func (c *defaultController) GetExpense(expenseId spendings.ExpenseId, actor spendings.CounterpartyId) (spendings.IdentifiableExpense, *common.CodeBasedError[spendings.GetExpenseErrorCode]) {
	const op = "spendings.defaultController.GetExpense"
	c.logger.LogInfo("%s: start[expenseId=%s actor=%s]", op, expenseId, actor)
	expense, err := c.repository.GetExpense(spendingsRepository.ExpenseId(expenseId))
	if err != nil {
		c.logger.LogInfo("%s: cannot get expense from db err: %v", op, err)
		return spendings.IdentifiableExpense{}, common.NewErrorWithDescription(spendings.GetExpenseErrorInternal, err.Error())
	}
	if expense == nil {
		c.logger.LogInfo("%s: expense %s is not found in db", op, expenseId)
		return spendings.IdentifiableExpense{}, common.NewError(spendings.GetExpenseErrorExpenseNotFound)
	}
	var isYourExpense bool
	for i := 0; i < len(expense.Shares); i++ {
		if expense.Shares[i].Counterparty == spendingsRepository.CounterpartyId(actor) {
			isYourExpense = true
			break
		}
	}
	if !isYourExpense {
		c.logger.LogInfo("%s: user %s is not found in expense %s shares", op, actor, expenseId)
		return spendings.IdentifiableExpense{}, common.NewError(spendings.GetExpenseErrorNotYourExpense)
	}
	c.logger.LogInfo("%s: success[expenseId=%s actor=%s]", op, expenseId, actor)
	return spendings.IdentifiableExpense(*expense), nil
}

func (c *defaultController) GetExpensesWith(counterparty spendings.CounterpartyId, actor spendings.CounterpartyId) ([]spendings.IdentifiableExpense, *common.CodeBasedError[spendings.GetExpensesErrorCode]) {
	const op = "spendings.defaultController.GetExpensesWith"
	c.logger.LogInfo("%s: start[counterparty=%s actor=%s]", op, counterparty, actor)
	expenses, err := c.repository.GetExpensesBetween(spendingsRepository.CounterpartyId(counterparty), spendingsRepository.CounterpartyId(actor))
	if err != nil {
		c.logger.LogInfo("%s: cannot get expenses from db err: %v", op, err)
		return []spendings.IdentifiableExpense{}, common.NewErrorWithDescription(spendings.GetExpensesErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[counterparty=%s actor=%s]", op, counterparty, actor)
	return common.Map(expenses, func(expense spendingsRepository.IdentifiableExpense) spendings.IdentifiableExpense {
		return spendings.IdentifiableExpense(expense)
	}), nil
}

func (c *defaultController) GetBalance(actor spendings.CounterpartyId) ([]spendings.Balance, *common.CodeBasedError[spendings.GetBalanceErrorCode]) {
	const op = "spendings.defaultController.GetBalance"
	c.logger.LogInfo("%s: start[actor=%s]", op, actor)
	balance, err := c.repository.GetBalance(spendingsRepository.CounterpartyId(actor))
	if err != nil {
		c.logger.LogInfo("%s: cannot get balance for %s from db err: %v", op, actor, err)
		return []spendings.Balance{}, common.NewErrorWithDescription(spendings.GetBalanceErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[actor=%s]", op, actor)
	return common.Map(balance, func(balance spendingsRepository.Balance) spendings.Balance {
		return spendings.Balance(balance)
	}), nil
}
