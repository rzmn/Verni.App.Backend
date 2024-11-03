package spendings

import (
	"verni/internal/common"
	httpserver "verni/internal/http-server"
	"verni/internal/repositories/spendings"
	"verni/internal/services/logging"
	"verni/internal/services/pushNotifications"
)

type defaultController struct {
	repository        Repository
	pushNotifications pushNotifications.Service
	logger            logging.Service
}

func (s *defaultController) AddExpense(expense Expense, actor CounterpartyId) *common.CodeBasedError[AddExpenseErrorCode] {
	const op = "spendings.defaultController.AddExpense"
	s.logger.Log("%s: start[actor=%s]", op, actor)
	var isYourExpense bool
	for i := 0; i < len(expense.Shares); i++ {
		if expense.Shares[i].Counterparty == spendings.CounterpartyId(actor) {
			isYourExpense = true
			break
		}
	}
	if !isYourExpense {
		s.logger.Log("%s: user %s is not found in expense %v shares", op, actor, expense)
		return common.NewError(AddExpenseErrorNotYourExpense)
	}
	transaction := s.repository.AddExpense(spendings.Expense(expense))
	expenseId, err := transaction.Perform()
	if err != nil {
		s.logger.Log("%s: cannot insert expense into db err: %v", op, err)
		return common.NewErrorWithDescription(AddExpenseErrorInternal, err.Error())
	}
	for i := 0; i < len(expense.Shares); i++ {
		spending := expense.Shares[i]
		if spending.Counterparty == spendings.CounterpartyId(actor) {
			continue
		}
		s.pushNotifications.NewExpenseReceived(pushNotifications.UserId(spending.Counterparty), pushNotifications.Expense{
			Expense: httpserver.Expense{
				Timestamp:   expense.Timestamp,
				Details:     expense.Details,
				Total:       httpserver.Cost(expense.Total),
				Attachments: []httpserver.ExpenseAttachment{},
				Currency:    httpserver.Currency(expense.Currency),
				Shares: common.Map(expense.Shares, func(share spendings.ShareOfExpense) httpserver.ShareOfExpense {
					return httpserver.ShareOfExpense{
						UserId: httpserver.UserId(share.Counterparty),
						Cost:   httpserver.Cost(share.Cost),
					}
				}),
			},
			Id: httpserver.ExpenseId(expenseId),
		}, pushNotifications.UserId(actor))
	}
	s.logger.Log("%s: success[actor=%s]", op, actor)
	return nil
}

func (s *defaultController) RemoveExpense(expenseId ExpenseId, actor CounterpartyId) (IdentifiableExpense, *common.CodeBasedError[RemoveExpenseErrorCode]) {
	const op = "spendings.defaultController.RemoveExpense"
	s.logger.Log("%s: start[expenseId=%s actor=%s]", op, expenseId, actor)
	expense, err := s.repository.GetExpense(spendings.ExpenseId(expenseId))
	if err != nil {
		s.logger.Log("%s: cannot get expense from db err: %v", op, err)
		return IdentifiableExpense{}, common.NewErrorWithDescription(RemoveExpenseErrorInternal, err.Error())
	}
	if expense == nil {
		s.logger.Log("%s: expense %s does not exists", op, expenseId)
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
		s.logger.Log("%s: user %s is not found in expense %s shares", op, actor, expenseId)
		return IdentifiableExpense{}, common.NewError(RemoveExpenseErrorNotYourExpense)
	}
	transaction := s.repository.RemoveExpense(spendings.ExpenseId(expenseId))
	if err := transaction.Perform(); err != nil {
		s.logger.Log("%s: cannot remove expense from db err: %v", op, err)
		return IdentifiableExpense{}, common.NewErrorWithDescription(RemoveExpenseErrorInternal, err.Error())
	}
	s.logger.Log("%s: success[expenseId=%s actor=%s]", op, expenseId, actor)
	return IdentifiableExpense(*expense), nil
}

func (s *defaultController) GetExpense(expenseId ExpenseId, actor CounterpartyId) (IdentifiableExpense, *common.CodeBasedError[GetExpenseErrorCode]) {
	const op = "spendings.defaultController.GetExpense"
	s.logger.Log("%s: start[expenseId=%s actor=%s]", op, expenseId, actor)
	expense, err := s.repository.GetExpense(spendings.ExpenseId(expenseId))
	if err != nil {
		s.logger.Log("%s: cannot get expense from db err: %v", op, err)
		return IdentifiableExpense{}, common.NewErrorWithDescription(GetExpenseErrorInternal, err.Error())
	}
	if expense == nil {
		s.logger.Log("%s: expense %s is not found in db", op, expenseId)
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
		s.logger.Log("%s: user %s is not found in expense %s shares", op, actor, expenseId)
		return IdentifiableExpense{}, common.NewError(GetExpenseErrorNotYourExpense)
	}
	s.logger.Log("%s: success[expenseId=%s actor=%s]", op, expenseId, actor)
	return IdentifiableExpense(*expense), nil
}

func (s *defaultController) GetExpensesWith(counterparty CounterpartyId, actor CounterpartyId) ([]IdentifiableExpense, *common.CodeBasedError[GetExpensesErrorCode]) {
	const op = "spendings.defaultController.GetExpensesWith"
	s.logger.Log("%s: start[counterparty=%s actor=%s]", op, counterparty, actor)
	expenses, err := s.repository.GetExpensesBetween(spendings.CounterpartyId(counterparty), spendings.CounterpartyId(actor))
	if err != nil {
		s.logger.Log("%s: cannot get expenses from db err: %v", op, err)
		return []IdentifiableExpense{}, common.NewErrorWithDescription(GetExpensesErrorInternal, err.Error())
	}
	s.logger.Log("%s: success[counterparty=%s actor=%s]", op, counterparty, actor)
	return common.Map(expenses, func(expense spendings.IdentifiableExpense) IdentifiableExpense {
		return IdentifiableExpense(expense)
	}), nil
}

func (s *defaultController) GetBalance(actor CounterpartyId) ([]Balance, *common.CodeBasedError[GetBalanceErrorCode]) {
	const op = "spendings.defaultController.GetBalance"
	s.logger.Log("%s: start[actor=%s]", op, actor)
	balance, err := s.repository.GetBalance(spendings.CounterpartyId(actor))
	if err != nil {
		s.logger.Log("%s: cannot get balance for %s from db err: %v", op, actor, err)
		return []Balance{}, common.NewErrorWithDescription(GetBalanceErrorInternal, err.Error())
	}
	s.logger.Log("%s: success[actor=%s]", op, actor)
	return common.Map(balance, func(balance spendings.Balance) Balance {
		return Balance(balance)
	}), nil
}
