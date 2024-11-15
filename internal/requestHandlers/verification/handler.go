package verification

import (
	verificationController "verni/internal/controllers/verification"
	"verni/internal/schema"
	"verni/internal/services/logging"
)

type RequestsHandler interface {
	ConfirmEmail(
		subject schema.UserId,
		request schema.ConfirmEmailRequest,
		success func(schema.StatusCode, schema.VoidResponse),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	SendEmailConfirmationCode(
		subject schema.UserId,
		success func(schema.StatusCode, schema.VoidResponse),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
}

func DefaultHandler(
	controller verificationController.Controller,
	logger logging.Service,
) RequestsHandler {
	return &defaultRequestsHandler{
		controller: controller,
		logger:     logger,
	}
}
