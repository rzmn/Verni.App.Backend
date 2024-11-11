package accessToken

import (
	"net/http"
	"strings"
	"verni/internal/common"
	httpserver "verni/internal/http-server"
	authRepository "verni/internal/repositories/auth"
	"verni/internal/services/jwt"
	"verni/internal/services/logging"
)

type defaultRequestsHandler struct {
	repository authRepository.Repository
	jwtService jwt.Service
	logger     logging.Service
}

func (c *defaultRequestsHandler) CheckToken(
	authorizationHeaderValue string,
	success func(httpserver.StatusCode, httpserver.Response[httpserver.UserId]),
	failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
) {
	const op = "requestHandlers.accessToken.defaultRequestsHandler.CheckToken"
	c.logger.LogInfo("%s: validating access token", op)
	token := func() jwt.AccessToken {
		if authorizationHeaderValue == "" {
			return ""
		}
		jwtToken := strings.Split(authorizationHeaderValue, " ")
		if len(jwtToken) != 2 {
			return ""
		}
		return jwt.AccessToken(jwtToken[1])
	}()
	if err := c.jwtService.ValidateAccessToken(token); err != nil {
		c.logger.LogInfo("%s: failed to validate token %v", op, err)
		switch err.Code {
		case jwt.CodeTokenExpired:
			failure(
				http.StatusUnauthorized,
				httpserver.Failure(
					common.NewErrorWithDescriptionValue(
						httpserver.CodeTokenExpired,
						err.Error(),
					),
				),
			)
		case jwt.CodeTokenInvalid:
			failure(
				http.StatusUnprocessableEntity,
				httpserver.Failure(
					common.NewErrorWithDescriptionValue(
						httpserver.CodeWrongAccessToken,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("jwt token validation failed %v", err)
			failure(
				http.StatusInternalServerError,
				httpserver.Failure(
					common.NewErrorWithDescriptionValue(
						httpserver.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	subject, getSubjectError := c.jwtService.GetAccessTokenSubject(token)
	if getSubjectError != nil {
		c.logger.LogError("jwt token get subject failed %v", getSubjectError)
		failure(
			http.StatusInternalServerError,
			httpserver.Failure(
				common.NewErrorWithDescriptionValue(
					httpserver.CodeInternal,
					getSubjectError.Error(),
				),
			),
		)
		return
	}
	exists, err := c.repository.IsUserExists(authRepository.UserId(subject))
	if err != nil {
		c.logger.LogError("valid token with invalid subject - %v", err)
		failure(
			http.StatusInternalServerError,
			httpserver.Failure(
				common.NewErrorWithDescriptionValue(
					httpserver.CodeInternal,
					err.Error(),
				),
			),
		)
		return
	}
	if !exists {
		failure(
			http.StatusUnprocessableEntity,
			httpserver.Failure(
				common.NewErrorWithDescriptionValue(
					httpserver.CodeWrongAccessToken,
					"associated user is not exists",
				),
			),
		)
		return
	}
	c.logger.LogInfo("%s: access token ok", op)
	success(http.StatusOK, httpserver.Success(httpserver.UserId(subject)))
}
