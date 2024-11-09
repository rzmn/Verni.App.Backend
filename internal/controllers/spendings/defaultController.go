package spendings

import (
	"verni/internal/common"
	"verni/internal/repositories/spendings"
	"verni/internal/services/logging"
)

type defaultController struct {
	repository Repository
	logger     logging.Service
}

func (c *defaultController) AddExpense(expense Expense, actor CounterpartyId) (IdentifiableExpense, *common.CodeBasedError[AddExpenseErrorCode]) {
	const op = "spendings.defaultController.AddExpense"
	c.logger.LogInfo("%s: start[actor=%s]", op, actor)
	var isYourExpense bool
	for i := 0; i < len(expense.Shares); i++ {
		if expense.Shares[i].Counterparty == spendings.CounterpartyId(actor) {
			isYourExpense = true
			break
		}
	}
	if !isYourExpense {
		c.logger.LogInfo("%s: user %s is not found in expense %v shares", op, actor, expense)
		return IdentifiableExpense{}, common.NewError(AddExpenseErrorNotYourExpense)
	}
	transaction := c.repository.AddExpense(spendings.Expense(expense))
	expenseId, err := transaction.Perform()
	if err != nil {
		c.logger.LogInfo("%s: cannot insert expense into db err: %v", op, err)
		return IdentifiableExpense{}, common.NewErrorWithDescription(AddExpenseErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[actor=%s]", op, actor)
	return IdentifiableExpense{
		Expense: spendings.Expense(expense),
		Id:      expenseId,
	}, nil
}

func (c *defaultController) RemoveExpense(expenseId ExpenseId, actor CounterpartyId) (IdentifiableExpense, *common.CodeBasedError[RemoveExpenseErrorCode]) {
	const op = "spendings.defaultController.RemoveExpense"
	c.logger.LogInfo("%s: start[expenseId=%s actor=%s]", op, expenseId, actor)
	expense, err := c.repository.GetExpense(spendings.ExpenseId(expenseId))
	if err != nil {
		c.logger.LogInfo("%s: cannot get expense from db err: %v", op, err)
		return IdentifiableExpense{}, common.NewErrorWithDescription(RemoveExpenseErrorInternal, err.Error())
	}
	if expense == nil {
		c.logger.LogInfo("%s: expense %s does not exists", op, expenseId)
		return IdentifiableExpense{}, common.NewError(RemoveExpenseErrorExpenseNotFound)
	}
	var isYourExpense bool
	for i := 0; i < len(expense.Shares); i++ {
		if expense.Shares[i].Counterparty == spendings.CounterpartyId(actor) {
			isYourExpense = true
			break
		}
	}
	if !isYourExpense {
		c.logger.LogInfo("%s: user %s is not found in expense %s shares", op, actor, expenseId)
		return IdentifiableExpense{}, common.NewError(RemoveExpenseErrorNotYourExpense)
	}
	transaction := c.repository.RemoveExpense(spendings.ExpenseId(expenseId))
	if err := transaction.Perform(); err != nil {
		c.logger.LogInfo("%s: cannot remove expense from db err: %v", op, err)
		return IdentifiableExpense{}, common.NewErrorWithDescription(RemoveExpenseErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[expenseId=%s actor=%s]", op, expenseId, actor)
	return IdentifiableExpense(*expense), nil
}

func (c *defaultController) GetExpense(expenseId ExpenseId, actor CounterpartyId) (IdentifiableExpense, *common.CodeBasedError[GetExpenseErrorCode]) {
	const op = "spendings.defaultController.GetExpense"
	c.logger.LogInfo("%s: start[expenseId=%s actor=%s]", op, expenseId, actor)
	expense, err := c.repository.GetExpense(spendings.ExpenseId(expenseId))
	if err != nil {
		c.logger.LogInfo("%s: cannot get expense from db err: %v", op, err)
		return IdentifiableExpense{}, common.NewErrorWithDescription(GetExpenseErrorInternal, err.Error())
	}
	if expense == nil {
		c.logger.LogInfo("%s: expense %s is not found in db", op, expenseId)
		return IdentifiableExpense{}, common.NewError(GetExpenseErrorExpenseNotFound)
	}
	var isYourExpense bool
	for i := 0; i < len(expense.Shares); i++ {
		if expense.Shares[i].Counterparty == spendings.CounterpartyId(actor) {
			isYourExpense = true
			break
		}
	}
	if !isYourExpense {
		c.logger.LogInfo("%s: user %s is not found in expense %s shares", op, actor, expenseId)
		return IdentifiableExpense{}, common.NewError(GetExpenseErrorNotYourExpense)
	}
	c.logger.LogInfo("%s: success[expenseId=%s actor=%s]", op, expenseId, actor)
	return IdentifiableExpense(*expense), nil
}

func (c *defaultController) GetExpensesWith(counterparty CounterpartyId, actor CounterpartyId) ([]IdentifiableExpense, *common.CodeBasedError[GetExpensesErrorCode]) {
	const op = "spendings.defaultController.GetExpensesWith"
	c.logger.LogInfo("%s: start[counterparty=%s actor=%s]", op, counterparty, actor)
	expenses, err := c.repository.GetExpensesBetween(spendings.CounterpartyId(counterparty), spendings.CounterpartyId(actor))
	if err != nil {
		c.logger.LogInfo("%s: cannot get expenses from db err: %v", op, err)
		return []IdentifiableExpense{}, common.NewErrorWithDescription(GetExpensesErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[counterparty=%s actor=%s]", op, counterparty, actor)
	return common.Map(expenses, func(expense spendings.IdentifiableExpense) IdentifiableExpense {
		return IdentifiableExpense(expense)
	}), nil
}

func (c *defaultController) GetBalance(actor CounterpartyId) ([]Balance, *common.CodeBasedError[GetBalanceErrorCode]) {
	const op = "spendings.defaultController.GetBalance"
	c.logger.LogInfo("%s: start[actor=%s]", op, actor)
	balance, err := c.repository.GetBalance(spendings.CounterpartyId(actor))
	if err != nil {
		c.logger.LogInfo("%s: cannot get balance for %s from db err: %v", op, actor, err)
		return []Balance{}, common.NewErrorWithDescription(GetBalanceErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[actor=%s]", op, actor)
	return common.Map(balance, func(balance spendings.Balance) Balance {
		return Balance(balance)
	}), nil
}
