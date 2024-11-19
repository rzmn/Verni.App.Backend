package accessToken

import (
	"github.com/rzmn/governi/internal/schema"
)

type RequestHandler interface {
	CheckToken(
		authorizationHeaderValue string,
		success func(schema.StatusCode, schema.Response[schema.UserId]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
}
