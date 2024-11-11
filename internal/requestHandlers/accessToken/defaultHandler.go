package accessToken

import (
	"net/http"
	"strings"
	"verni/internal/common"
	authRepository "verni/internal/repositories/auth"
	"verni/internal/schema"
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
	success func(schema.StatusCode, schema.Response[schema.UserId]),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
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
				schema.Failure(
					common.NewErrorWithDescriptionValue(
						schema.CodeTokenExpired,
						err.Error(),
					),
				),
			)
		case jwt.CodeTokenInvalid:
			failure(
				http.StatusUnprocessableEntity,
				schema.Failure(
					common.NewErrorWithDescriptionValue(
						schema.CodeWrongAccessToken,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("jwt token validation failed %v", err)
			failure(
				http.StatusInternalServerError,
				schema.Failure(
					common.NewErrorWithDescriptionValue(
						schema.CodeInternal,
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
			schema.Failure(
				common.NewErrorWithDescriptionValue(
					schema.CodeInternal,
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
			schema.Failure(
				common.NewErrorWithDescriptionValue(
					schema.CodeInternal,
					err.Error(),
				),
			),
		)
		return
	}
	if !exists {
		failure(
			http.StatusUnprocessableEntity,
			schema.Failure(
				common.NewErrorWithDescriptionValue(
					schema.CodeWrongAccessToken,
					"associated user is not exists",
				),
			),
		)
		return
	}
	c.logger.LogInfo("%s: access token ok", op)
	success(http.StatusOK, schema.Success(schema.UserId(subject)))
}
