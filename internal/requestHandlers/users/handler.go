package users

import (
	usersController "verni/internal/controllers/users"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/responses"
	"verni/internal/services/logging"
)

type GetUsersRequest struct {
	Ids []httpserver.UserId `json:"ids"`
}

type RequestsHandler interface {
	GetUsers(
		subject httpserver.UserId,
		request GetUsersRequest,
		success func(httpserver.StatusCode, responses.Response[[]httpserver.User]),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
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
