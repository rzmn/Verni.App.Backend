package profile

import (
	profileController "verni/internal/controllers/profile"
	httpserver "verni/internal/http-server"
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
		success func(httpserver.StatusCode, httpserver.Response[httpserver.Profile]),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
	SetAvatar(
		subject httpserver.UserId,
		request SetAvatarRequest,
		success func(httpserver.StatusCode, httpserver.Response[httpserver.ImageId]),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
	SetDisplayName(
		subject httpserver.UserId,
		request SetDisplayNameRequest,
		success func(httpserver.StatusCode, httpserver.VoidResponse),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
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
