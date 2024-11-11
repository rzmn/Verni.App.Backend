package accessToken

import (
	httpserver "verni/internal/http-server"
	authRepository "verni/internal/repositories/auth"
	"verni/internal/services/jwt"
	"verni/internal/services/logging"
)

type RequestHandler interface {
	CheckToken(
		authorizationHeaderValue string,
		success func(httpserver.StatusCode, httpserver.Response[httpserver.UserId]),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
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
