package spendings

import (
	"github.com/rzmn/Verni.App.Backend/internal/schema"
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
