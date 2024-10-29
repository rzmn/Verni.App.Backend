package spendings

import (
	"log"
	"verni/internal/common"
	httpserver "verni/internal/http-server"
	"verni/internal/repositories/spendings"
	"verni/internal/services/pushNotifications"
)

type defaultController struct {
	repository        Repository
	pushNotifications pushNotifications.Service
}

func (s *defaultController) AddExpense(expense Expense, actor CounterpartyId) *common.CodeBasedError[CreateDealErrorCode] {
	const op = "spendings.defaultController.AddExpense"
	log.Printf("%s: start[actor=%s]", op, actor)
	transaction := s.repository.AddExpense(spendings.Expense(expense))
	expenseId, err := transaction.Perform()
	if err != nil {
		log.Printf("%s: cannot insert expense into db err: %v", op, err)
		return common.NewErrorWithDescription(CreateDealErrorInternal, err.Error())
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
	log.Printf("%s: success[actor=%s]", op, actor)
	return nil
}

func (s *defaultController) RemoveExpense(expenseId ExpenseId, actor CounterpartyId) (IdentifiableExpense, *common.CodeBasedError[DeleteDealErrorCode]) {
	const op = "spendings.defaultController.RemoveExpense"
	log.Printf("%s: start[expenseId=%s actor=%s]", op, expenseId, actor)
	expense, err := s.repository.GetExpense(spendings.ExpenseId(expenseId))
	if err != nil {
		log.Printf("%s: cannot get expense from db err: %v", op, err)
		return IdentifiableExpense{}, common.NewErrorWithDescription(DeleteDealErrorInternal, err.Error())
	}
	if expense == nil {
		log.Printf("%s: expense %s does not exists", op, expenseId)
		return IdentifiableExpense{}, common.NewError(DeleteDealErrorDealNotFound)
	}
	var isYourExpense bool
	for i := 0; i < len(expense.Shares); i++ {
		if expense.Shares[i].Counterparty == spendings.CounterpartyId(actor) {
			isYourExpense = true
			break
		}
	}
	if !isYourExpense {
		log.Printf("%s: user %s is not found in expense %s shares", op, actor, expenseId)
		return IdentifiableExpense{}, common.NewError(DeleteDealErrorNotYourDeal)
	}
	transaction := s.repository.RemoveExpense(spendings.ExpenseId(expenseId))
	if err := transaction.Perform(); err != nil {
		log.Printf("%s: cannot remove expense from db err: %v", op, err)
		return IdentifiableExpense{}, common.NewErrorWithDescription(DeleteDealErrorInternal, err.Error())
	}
	log.Printf("%s: success[expenseId=%s actor=%s]", op, expenseId, actor)
	return IdentifiableExpense(*expense), nil
}

func (s *defaultController) GetExpense(expenseId ExpenseId, actor CounterpartyId) (IdentifiableExpense, *common.CodeBasedError[GetDealErrorCode]) {
	const op = "spendings.defaultController.GetExpense"
	log.Printf("%s: start[expenseId=%s actor=%s]", op, expenseId, actor)
	expense, err := s.repository.GetExpense(spendings.ExpenseId(expenseId))
	if err != nil {
		log.Printf("%s: cannot get expense from db err: %v", op, err)
		return IdentifiableExpense{}, common.NewErrorWithDescription(GetDealErrorInternal, err.Error())
	}
	if expense == nil {
		log.Printf("%s: expense %s is not found in db", op, expenseId)
		return IdentifiableExpense{}, common.NewError(GetDealErrorDealNotFound)
	}
	log.Printf("%s: success[expenseId=%s actor=%s]", op, expenseId, actor)
	return IdentifiableExpense(*expense), nil
}

func (s *defaultController) GetExpensesWith(counterparty CounterpartyId, actor CounterpartyId) ([]IdentifiableExpense, *common.CodeBasedError[GetDealsErrorCode]) {
	const op = "spendings.defaultController.GetExpensesWith"
	log.Printf("%s: start[counterparty=%s actor=%s]", op, counterparty, actor)
	expenses, err := s.repository.GetExpensesBetween(spendings.CounterpartyId(counterparty), spendings.CounterpartyId(actor))
	if err != nil {
		log.Printf("%s: cannot get expenses from db err: %v", op, err)
		return []IdentifiableExpense{}, common.NewErrorWithDescription(GetDealsErrorInternal, err.Error())
	}
	log.Printf("%s: success[counterparty=%s actor=%s]", op, counterparty, actor)
	return common.Map(expenses, func(expense spendings.IdentifiableExpense) IdentifiableExpense {
		return IdentifiableExpense(expense)
	}), nil
}

func (s *defaultController) GetBalance(actor CounterpartyId) ([]Balance, *common.CodeBasedError[GetCounterpartiesErrorCode]) {
	const op = "spendings.defaultController.GetBalance"
	log.Printf("%s: start[actor=%s]", op, actor)
	balance, err := s.repository.GetBalance(spendings.CounterpartyId(actor))
	if err != nil {
		log.Printf("%s: cannot get balance for %s from db err: %v", op, actor, err)
		return []Balance{}, common.NewErrorWithDescription(GetCounterpartiesErrorInternal, err.Error())
	}
	log.Printf("%s: success[actor=%s]", op, actor)
	return common.Map(balance, func(balance spendings.Balance) Balance {
		return Balance(balance)
	}), nil
}
