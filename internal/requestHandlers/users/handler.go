package users

import (
	usersController "verni/internal/controllers/users"
	"verni/internal/schema"
	"verni/internal/services/logging"
)

type RequestsHandler interface {
	GetUsers(
		subject schema.UserId,
		request schema.GetUsersRequest,
		success func(schema.StatusCode, schema.Response[[]schema.User]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
}

func DefaultHandler(
	controller usersController.Controller,
	logger logging.Service,
) RequestsHandler {
	return &defaultRequestsHandler{
		controller: controller,
		logger:     logger,
	}
}
