package avatars

import (
	avatarsController "verni/internal/controllers/avatars"
	httpserver "verni/internal/http-server"
	"verni/internal/services/logging"
)

type GetAvatarsRequest struct {
	Ids []httpserver.ImageId `json:"ids"`
}

type RequestsHandler interface {
	GetAvatars(
		request GetAvatarsRequest,
		success func(httpserver.StatusCode, httpserver.Response[map[httpserver.ImageId]httpserver.Image]),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
}

func DefaultHandler(
	controller avatarsController.Controller,
	logger logging.Service,
) RequestsHandler {
	return &defaultRequestsHandler{
		controller: controller,
		logger:     logger,
	}
}
