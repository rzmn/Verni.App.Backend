package spendings

import (
	spendingsController "verni/internal/controllers/spendings"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/longpoll"
	"verni/internal/http-server/responses"
	"verni/internal/services/logging"
	"verni/internal/services/pushNotifications"
)

type HttpCode int

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
		subject spendingsController.CounterpartyId,
		request AddExpenseRequest,
		success func(HttpCode, responses.Response[httpserver.IdentifiableExpense]),
		failure func(HttpCode, responses.Response[responses.Error]),
	)
	RemoveExpense(
		subject spendingsController.CounterpartyId,
		request RemoveExpenseRequest,
		success func(HttpCode, responses.Response[httpserver.IdentifiableExpense]),
		failure func(HttpCode, responses.Response[responses.Error]),
	)
	GetBalance(
		subject spendingsController.CounterpartyId,
		success func(HttpCode, responses.Response[[]httpserver.Balance]),
		failure func(HttpCode, responses.Response[responses.Error]),
	)
	GetExpenses(
		subject spendingsController.CounterpartyId,
		request GetExpensesRequest,
		success func(HttpCode, responses.Response[[]httpserver.IdentifiableExpense]),
		failure func(HttpCode, responses.Response[responses.Error]),
	)
	GetExpense(
		subject spendingsController.CounterpartyId,
		request GetExpenseRequest,
		success func(HttpCode, responses.Response[httpserver.IdentifiableExpense]),
		failure func(HttpCode, responses.Response[responses.Error]),
	)
}

func DefaultHandler(
	controller spendingsController.Controller,
	pushService pushNotifications.Service,
	pollingService longpoll.Service,
	logger logging.Service,
) RequestsHandler {
	return &defaultRequestsHandler{}
}
