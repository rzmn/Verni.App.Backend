package accessToken

import (
	"errors"
	"net/http"
	"strings"
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
			failure(http.StatusUnauthorized, schema.Failure(err, schema.CodeTokenExpired))
		case jwt.CodeTokenInvalid:
			failure(http.StatusUnprocessableEntity, schema.Failure(err, schema.CodeWrongAccessToken))
		default:
			c.logger.LogError("%s: jwt token validation failed %v", op, err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
		return
	}
	subject, getSubjectError := c.jwtService.GetAccessTokenSubject(token)
	if getSubjectError != nil {
		c.logger.LogError("%s: jwt token get subject failed %v", op, getSubjectError)
		failure(http.StatusInternalServerError, schema.Failure(getSubjectError, schema.CodeInternal))
		return
	}
	exists, err := c.repository.IsUserExists(authRepository.UserId(subject))
	if err != nil {
		c.logger.LogError("%s: valid token with invalid subject - %v", op, err)
		failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		return
	}
	if !exists {
		c.logger.LogError("%s: associated user is not exists", op)
		failure(http.StatusUnprocessableEntity, schema.Failure(errors.New("associated user is not exists"), schema.CodeWrongAccessToken))
		return
	}
	c.logger.LogInfo("%s: access token ok", op)
	success(http.StatusOK, schema.Success(schema.UserId(subject)))
}
