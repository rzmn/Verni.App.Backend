package accessToken

import (
	"verni/internal/schema"
)

type RequestHandler interface {
	CheckToken(
		authorizationHeaderValue string,
		success func(schema.StatusCode, schema.Response[schema.UserId]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
}
