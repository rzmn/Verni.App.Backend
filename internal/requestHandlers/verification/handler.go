package verification

import (
	"github.com/rzmn/governi/internal/schema"
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
