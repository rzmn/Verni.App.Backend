package spendings

import (
	spendingsController "verni/internal/controllers/spendings"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/longpoll"
	"verni/internal/services/logging"
	"verni/internal/services/pushNotifications"
)

type AddExpenseRequest struct {
	Expense httpserver.Expense `json:"expense"`
}

type RemoveExpenseRequest struct {
	ExpenseId httpserver.ExpenseId `json:"expenseId"`
}

type GetExpensesRequest struct {
	Counterparty httpserver.UserId `json:"counterparty"`
}

type GetExpenseRequest struct {
	Id httpserver.ExpenseId `json:"id"`
}

type RequestsHandler interface {
	AddExpense(
		subject httpserver.UserId,
		request AddExpenseRequest,
		success func(httpserver.StatusCode, httpserver.Response[httpserver.IdentifiableExpense]),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
	RemoveExpense(
		subject httpserver.UserId,
		request RemoveExpenseRequest,
		success func(httpserver.StatusCode, httpserver.Response[httpserver.IdentifiableExpense]),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
	GetBalance(
		subject httpserver.UserId,
		success func(httpserver.StatusCode, httpserver.Response[[]httpserver.Balance]),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
	GetExpenses(
		subject httpserver.UserId,
		request GetExpensesRequest,
		success func(httpserver.StatusCode, httpserver.Response[[]httpserver.IdentifiableExpense]),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
	GetExpense(
		subject httpserver.UserId,
		request GetExpenseRequest,
		success func(httpserver.StatusCode, httpserver.Response[httpserver.IdentifiableExpense]),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
}

func DefaultHandler(
	controller spendingsController.Controller,
	pushService pushNotifications.Service,
	pollingService longpoll.Service,
	logger logging.Service,
) RequestsHandler {
	return &defaultRequestsHandler{
		controller:     controller,
		pushService:    pushService,
		pollingService: pollingService,
		logger:         logger,
	}
}
