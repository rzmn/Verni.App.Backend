package spendings

import (
	spendingsController "verni/internal/controllers/spendings"
	"verni/internal/schema"
	"verni/internal/services/logging"
	"verni/internal/services/pushNotifications"
	"verni/internal/services/realtimeEvents"
)

type RequestsHandler interface {
	AddExpense(
		subject schema.UserId,
		request schema.AddExpenseRequest,
		success func(schema.StatusCode, schema.Response[schema.IdentifiableExpense]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	RemoveExpense(
		subject schema.UserId,
		request schema.RemoveExpenseRequest,
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
		request schema.GetExpensesRequest,
		success func(schema.StatusCode, schema.Response[[]schema.IdentifiableExpense]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	GetExpense(
		subject schema.UserId,
		request schema.GetExpenseRequest,
		success func(schema.StatusCode, schema.Response[schema.IdentifiableExpense]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
}

func DefaultHandler(
	controller spendingsController.Controller,
	pushService pushNotifications.Service,
	realtimeEvents realtimeEvents.Service,
	logger logging.Service,
) RequestsHandler {
	return &defaultRequestsHandler{
		controller:     controller,
		pushService:    pushService,
		realtimeEvents: realtimeEvents,
		logger:         logger,
	}
}
