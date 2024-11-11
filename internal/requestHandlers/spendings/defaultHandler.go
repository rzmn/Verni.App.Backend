package spendings

import (
	"net/http"
	"verni/internal/common"
	spendingsController "verni/internal/controllers/spendings"
	httpserver "verni/internal/http-server"
	spendingsRepository "verni/internal/repositories/spendings"
	"verni/internal/services/logging"
	"verni/internal/services/longpoll"
	"verni/internal/services/pushNotifications"
)

type defaultRequestsHandler struct {
	controller     spendingsController.Controller
	pushService    pushNotifications.Service
	pollingService longpoll.Service
	logger         logging.Service
}

func (c *defaultRequestsHandler) AddExpense(
	subject httpserver.UserId,
	request AddExpenseRequest,
	success func(httpserver.StatusCode, httpserver.Response[httpserver.IdentifiableExpense]),
	failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
) {
	expense, err := c.controller.AddExpense(mapHttpServerExpense(request.Expense), spendingsController.CounterpartyId(subject))
	if err != nil {
		switch err.Code {
		case spendingsController.AddExpenseErrorNoSuchUser:
			failure(
				http.StatusConflict,
				httpserver.Failure(
					common.NewErrorWithDescriptionValue(
						httpserver.CodeNoSuchUser,
						err.Error(),
					),
				),
			)
		case spendingsController.AddExpenseErrorNotYourExpense:
			failure(
				http.StatusConflict,
				httpserver.Failure(
					common.NewErrorWithDescriptionValue(
						httpserver.CodeIsNotYourExpense,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("addExpense request %v failed with unknown err: %v", request, err)
			failure(
				http.StatusInternalServerError,
				httpserver.Failure(
					common.NewErrorWithDescriptionValue(
						httpserver.CodeInternal,
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
	success(http.StatusOK, httpserver.Success(mapIdentifiableExpense(expense)))
}

func (c *defaultRequestsHandler) RemoveExpense(
	subject httpserver.UserId,
	request RemoveExpenseRequest,
	success func(httpserver.StatusCode, httpserver.Response[httpserver.IdentifiableExpense]),
	failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
) {
	expense, err := c.controller.RemoveExpense(spendingsController.ExpenseId(request.ExpenseId), spendingsController.CounterpartyId(subject))
	if err != nil {
		switch err.Code {
		case spendingsController.RemoveExpenseErrorExpenseNotFound:
			failure(
				http.StatusConflict,
				httpserver.Failure(
					common.NewErrorWithDescriptionValue(
						httpserver.CodeExpenseNotFound,
						err.Error(),
					),
				),
			)
		case spendingsController.RemoveExpenseErrorNotAFriend:
			failure(
				http.StatusConflict,
				httpserver.Failure(
					common.NewErrorWithDescriptionValue(
						httpserver.CodeNotAFriend,
						err.Error(),
					),
				),
			)
		case spendingsController.RemoveExpenseErrorNotYourExpense:
			failure(
				http.StatusConflict,
				httpserver.Failure(
					common.NewErrorWithDescriptionValue(
						httpserver.CodeIsNotYourExpense,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("removeExpense request %v failed with unknown err: %v", request, err)
			failure(
				http.StatusInternalServerError,
				httpserver.Failure(
					common.NewErrorWithDescriptionValue(
						httpserver.CodeInternal,
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
	success(http.StatusOK, httpserver.Success(mapIdentifiableExpense(expense)))
}

func (c *defaultRequestsHandler) GetBalance(
	subject httpserver.UserId,
	success func(httpserver.StatusCode, httpserver.Response[[]httpserver.Balance]),
	failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
) {
	balance, err := c.controller.GetBalance(spendingsController.CounterpartyId(subject))
	if err != nil {
		switch err.Code {
		default:
			c.logger.LogError("getBalance request failed with unknown err: %v", err)
			failure(
				http.StatusInternalServerError,
				httpserver.Failure(
					common.NewErrorWithDescriptionValue(
						httpserver.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	success(http.StatusOK, httpserver.Success(common.Map(balance, mapBalance)))
}

func (c *defaultRequestsHandler) GetExpenses(
	subject httpserver.UserId,
	request GetExpensesRequest,
	success func(httpserver.StatusCode, httpserver.Response[[]httpserver.IdentifiableExpense]),
	failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
) {
	expenses, err := c.controller.GetExpensesWith(spendingsController.CounterpartyId(request.Counterparty), spendingsController.CounterpartyId(subject))
	if err != nil {
		switch err.Code {
		default:
			c.logger.LogError("getExpenses request %v failed with unknown err: %v", request, err)
			failure(
				http.StatusInternalServerError,
				httpserver.Failure(
					common.NewErrorWithDescriptionValue(
						httpserver.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	success(http.StatusOK, httpserver.Success(common.Map(expenses, mapIdentifiableExpense)))
}

func (c *defaultRequestsHandler) GetExpense(
	subject httpserver.UserId,
	request GetExpenseRequest,
	success func(httpserver.StatusCode, httpserver.Response[httpserver.IdentifiableExpense]),
	failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
) {
	expense, err := c.controller.GetExpense(spendingsController.ExpenseId(request.Id), spendingsController.CounterpartyId(subject))
	if err != nil {
		switch err.Code {
		case spendingsController.GetExpenseErrorExpenseNotFound:
			failure(
				http.StatusConflict,
				httpserver.Failure(
					common.NewErrorWithDescriptionValue(
						httpserver.CodeExpenseNotFound,
						err.Error(),
					),
				),
			)
		case spendingsController.GetExpenseErrorNotYourExpense:
			failure(
				http.StatusConflict,
				httpserver.Failure(
					common.NewErrorWithDescriptionValue(
						httpserver.CodeIsNotYourExpense,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("getExpense request %v failed with unknown err: %v", request, err)
			failure(
				http.StatusInternalServerError,
				httpserver.Failure(
					common.NewErrorWithDescriptionValue(
						httpserver.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	success(http.StatusOK, httpserver.Success(mapIdentifiableExpense(expense)))
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
