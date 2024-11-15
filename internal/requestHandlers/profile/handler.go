package profile

import (
	profileController "verni/internal/controllers/profile"
	"verni/internal/schema"
	"verni/internal/services/logging"
)

type RequestsHandler interface {
	GetInfo(
		subject schema.UserId,
		success func(schema.StatusCode, schema.Response[schema.Profile]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	SetAvatar(
		subject schema.UserId,
		request schema.SetAvatarRequest,
		success func(schema.StatusCode, schema.Response[schema.ImageId]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	SetDisplayName(
		subject schema.UserId,
		request schema.SetDisplayNameRequest,
		success func(schema.StatusCode, schema.VoidResponse),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
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
