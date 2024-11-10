package profile

import (
	profileController "verni/internal/controllers/profile"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/responses"
	"verni/internal/services/logging"
)

type SetAvatarRequest struct {
	DataBase64 string `json:"dataBase64"`
}

type SetDisplayNameRequest struct {
	DisplayName string `json:"displayName"`
}

type RequestsHandler interface {
	GetInfo(
		subject httpserver.UserId,
		success func(httpserver.StatusCode, responses.Response[httpserver.Profile]),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
	)
	SetAvatar(
		subject httpserver.UserId,
		request SetAvatarRequest,
		success func(httpserver.StatusCode, responses.Response[httpserver.ImageId]),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
	)
	SetDisplayName(
		subject httpserver.UserId,
		request SetDisplayNameRequest,
		success func(httpserver.StatusCode, responses.VoidResponse),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
	)
}

func DefaultHandler(
	controller profileController.Controller,
	logger logging.Service,
) RequestsHandler {
	return &defaultRequestsHandler{
		controller: controller,
		logger:     logger,
	}
}
