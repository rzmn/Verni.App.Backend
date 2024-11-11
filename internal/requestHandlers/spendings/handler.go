package spendings

import (
	spendingsController "verni/internal/controllers/spendings"
	"verni/internal/schema"
	"verni/internal/services/logging"
	"verni/internal/services/longpoll"
	"verni/internal/services/pushNotifications"
)

type AddExpenseRequest struct {
	Expense schema.Expense `json:"expense"`
}

type RemoveExpenseRequest struct {
	ExpenseId schema.ExpenseId `json:"expenseId"`
}

type GetExpensesRequest struct {
	Counterparty schema.UserId `json:"counterparty"`
}

type GetExpenseRequest struct {
	Id schema.ExpenseId `json:"id"`
}

type RequestsHandler interface {
	AddExpense(
		subject schema.UserId,
		request AddExpenseRequest,
		success func(schema.StatusCode, schema.Response[schema.IdentifiableExpense]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	RemoveExpense(
		subject schema.UserId,
		request RemoveExpenseRequest,
		success func(schema.StatusCode, schema.Response[schema.IdentifiableExpense]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	GetBalance(
		subject schema.UserId,
		success func(schema.StatusCode, schema.Response[[]schema.Balance]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	GetExpenses(
		subject schema.UserId,
		request GetExpensesRequest,
		success func(schema.StatusCode, schema.Response[[]schema.IdentifiableExpense]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	GetExpense(
		subject schema.UserId,
		request GetExpenseRequest,
		success func(schema.StatusCode, schema.Response[schema.IdentifiableExpense]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
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
