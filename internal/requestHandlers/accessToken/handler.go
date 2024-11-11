package accessToken

import (
	authRepository "verni/internal/repositories/auth"
	"verni/internal/schema"
	"verni/internal/services/jwt"
	"verni/internal/services/logging"
)

type RequestHandler interface {
	CheckToken(
		authorizationHeaderValue string,
		success func(schema.StatusCode, schema.Response[schema.UserId]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
}

func DefaultHandler(
	repository authRepository.Repository,
	jwtService jwt.Service,
	logger logging.Service,
) RequestHandler {
	return &defaultRequestsHandler{
		repository: repository,
		jwtService: jwtService,
		logger:     logger,
	}
}
