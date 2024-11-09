package spendings

import (
	"net/http"
	"verni/internal/common"
	spendingsController "verni/internal/controllers/spendings"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/longpoll"
	"verni/internal/http-server/responses"
	spendingsRepository "verni/internal/repositories/spendings"
	"verni/internal/services/logging"
	"verni/internal/services/pushNotifications"
)

type defaultRequestsHandler struct {
	controller     spendingsController.Controller
	pushService    pushNotifications.Service
	pollingService longpoll.Service
	logger         logging.Service
}

func (c *defaultRequestsHandler) AddExpense(
	subject spendingsController.CounterpartyId,
	request AddExpenseRequest,
	success func(HttpCode, responses.Response[httpserver.IdentifiableExpense]),
	failure func(HttpCode, responses.Response[responses.Error]),
) {
	expense, err := c.controller.AddExpense(mapHttpServerExpense(request.Expense), subject)
	if err != nil {
		switch err.Code {
		case spendingsController.AddExpenseErrorNoSuchUser:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeNoSuchUser,
						err.Error(),
					),
				),
			)
		case spendingsController.AddExpenseErrorNotYourExpense:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeIsNotYourExpense,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("addExpense request %v failed with unknown err: %v", request, err)
			failure(
				http.StatusInternalServerError,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	for _, share := range expense.Shares {
		if share.Counterparty == spendingsRepository.CounterpartyId(subject) {
			continue
		}
		c.pushService.NewExpenseReceived(
			pushNotifications.UserId(share.Counterparty),
			pushNotifications.Expense(mapIdentifiableExpense(expense)),
			pushNotifications.UserId(subject),
		)
		c.pollingService.ExpensesUpdated(longpoll.UserId(share.Counterparty), longpoll.UserId(subject))
		c.pollingService.CounterpartiesUpdated(longpoll.UserId(share.Counterparty))
	}
	success(http.StatusOK, responses.Success(mapIdentifiableExpense(expense)))
}

func (c *defaultRequestsHandler) RemoveExpense(
	subject spendingsController.CounterpartyId,
	request RemoveExpenseRequest,
	success func(HttpCode, responses.Response[httpserver.IdentifiableExpense]),
	failure func(HttpCode, responses.Response[responses.Error]),
) {
	expense, err := c.controller.RemoveExpense(spendingsController.ExpenseId(request.ExpenseId), subject)
	if err != nil {
		switch err.Code {
		case spendingsController.RemoveExpenseErrorExpenseNotFound:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeExpenseNotFound,
						err.Error(),
					),
				),
			)
		case spendingsController.RemoveExpenseErrorNotAFriend:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeNotAFriend,
						err.Error(),
					),
				),
			)
		case spendingsController.RemoveExpenseErrorNotYourExpense:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeIsNotYourExpense,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("removeExpense request %v failed with unknown err: %v", request, err)
			failure(
				http.StatusInternalServerError,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	for _, share := range expense.Shares {
		if share.Counterparty == spendingsRepository.CounterpartyId(subject) {
			continue
		}
		c.pushService.NewExpenseReceived(
			pushNotifications.UserId(share.Counterparty),
			pushNotifications.Expense(mapIdentifiableExpense(expense)),
			pushNotifications.UserId(subject),
		)
		c.pollingService.ExpensesUpdated(longpoll.UserId(share.Counterparty), longpoll.UserId(subject))
		c.pollingService.CounterpartiesUpdated(longpoll.UserId(share.Counterparty))
	}
	success(http.StatusOK, responses.Success(mapIdentifiableExpense(expense)))
}

func (c *defaultRequestsHandler) GetBalance(
	subject spendingsController.CounterpartyId,
	success func(HttpCode, responses.Response[[]httpserver.Balance]),
	failure func(HttpCode, responses.Response[responses.Error]),
) {
	balance, err := c.controller.GetBalance(subject)
	if err != nil {
		switch err.Code {
		default:
			c.logger.LogError("getBalance request failed with unknown err: %v", err)
			failure(
				http.StatusInternalServerError,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	success(http.StatusOK, responses.Success(common.Map(balance, mapBalance)))
}

func (c *defaultRequestsHandler) GetExpenses(
	subject spendingsController.CounterpartyId,
	request GetExpensesRequest,
	success func(HttpCode, responses.Response[[]httpserver.IdentifiableExpense]),
	failure func(HttpCode, responses.Response[responses.Error]),
) {
	expenses, err := c.controller.GetExpensesWith(spendingsController.CounterpartyId(request.Counterparty), subject)
	if err != nil {
		switch err.Code {
		default:
			c.logger.LogError("getExpenses request %v failed with unknown err: %v", request, err)
			failure(
				http.StatusInternalServerError,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	success(http.StatusOK, responses.Success(common.Map(expenses, mapIdentifiableExpense)))
}

func (c *defaultRequestsHandler) GetExpense(
	subject spendingsController.CounterpartyId,
	request GetExpenseRequest,
	success func(HttpCode, responses.Response[httpserver.IdentifiableExpense]),
	failure func(HttpCode, responses.Response[responses.Error]),
) {
	expense, err := c.controller.GetExpense(spendingsController.ExpenseId(request.Id), subject)
	if err != nil {
		switch err.Code {
		case spendingsController.GetExpenseErrorExpenseNotFound:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeExpenseNotFound,
						err.Error(),
					),
				),
			)
		case spendingsController.GetExpenseErrorNotYourExpense:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeIsNotYourExpense,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("getExpense request %v failed with unknown err: %v", request, err)
			failure(
				http.StatusInternalServerError,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	success(http.StatusOK, responses.Success(mapIdentifiableExpense(expense)))
}

func mapHttpServerExpense(expense httpserver.Expense) spendingsController.Expense {
	return spendingsController.Expense{
		Timestamp: expense.Timestamp,
		Details:   expense.Details,
		Total:     spendingsRepository.Cost(expense.Total),
		Currency:  spendingsRepository.Currency(expense.Currency),
		Shares: common.Map(expense.Shares, func(share httpserver.ShareOfExpense) spendingsRepository.ShareOfExpense {
			return spendingsRepository.ShareOfExpense{
				Counterparty: spendingsRepository.CounterpartyId(share.UserId),
				Cost:         spendingsRepository.Cost(share.Cost),
			}
		}),
	}
}

func mapIdentifiableExpense(expense spendingsController.IdentifiableExpense) httpserver.IdentifiableExpense {
	return httpserver.IdentifiableExpense{
		Id:      httpserver.ExpenseId(expense.Id),
		Expense: mapExpense(spendingsController.Expense(expense.Expense)),
	}
}

func mapExpense(expense spendingsController.Expense) httpserver.Expense {
	return httpserver.Expense{
		Timestamp:   expense.Timestamp,
		Details:     expense.Details,
		Total:       httpserver.Cost(expense.Total),
		Attachments: []httpserver.ExpenseAttachment{},
		Currency:    httpserver.Currency(expense.Currency),
		Shares:      common.Map(expense.Shares, mapShareOfExpense),
	}
}

func mapShareOfExpense(share spendingsRepository.ShareOfExpense) httpserver.ShareOfExpense {
	return httpserver.ShareOfExpense{
		UserId: httpserver.UserId(share.Counterparty),
		Cost:   httpserver.Cost(share.Cost),
	}
}

func mapBalance(balance spendingsController.Balance) httpserver.Balance {
	currencies := map[httpserver.Currency]httpserver.Cost{}
	for currency, cost := range balance.Currencies {
		currencies[httpserver.Currency(currency)] = httpserver.Cost(cost)
	}
	return httpserver.Balance{
		Counterparty: string(balance.Counterparty),
		Currencies:   currencies,
	}
}
