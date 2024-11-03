package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	httpserver "verni/internal/http-server"
	authRepository "verni/internal/repositories/auth"
	"verni/internal/services/jwt"
	"verni/internal/services/logging"

	"verni/internal/http-server/responses"
)

type UserId string
type AccessTokenChecker struct {
	Handler     gin.HandlerFunc
	AccessToken func(c *gin.Context) UserId
}

const (
	accessTokenSubjectKey = "verni-subject"
)

func JwsAccessTokenCheck(repository authRepository.Repository, jwtService jwt.Service, logger logging.Service) AccessTokenChecker {
	return AccessTokenChecker{
		Handler: func(c *gin.Context) {
			const op = "handlers.friends.ensureLoggedInMiddleware"
			logger.Log("%s: validating access token", op)
			token := jwt.AccessToken(extractBearerToken(c))
			if err := jwtService.ValidateAccessToken(token); err != nil {
				logger.Log("%s: failed to validate token %v", op, err)
				switch err.Code {
				case jwt.CodeTokenExpired:
					httpserver.Answer(c, err, http.StatusUnauthorized, responses.CodeTokenExpired)
				case jwt.CodeTokenInvalid:
					httpserver.Answer(c, err, http.StatusUnprocessableEntity, responses.CodeWrongAccessToken)
				default:
					httpserver.AnswerWithUnknownError(c, err)
				}
				return
			}
			subject, getSubjectError := jwtService.GetAccessTokenSubject(token)
			if getSubjectError != nil {
				httpserver.AnswerWithUnknownError(c, getSubjectError)
				return
			}
			exists, err := repository.IsUserExists(authRepository.UserId(subject))
			if err != nil {
				httpserver.AnswerWithUnknownError(c, err)
				return
			}
			if !exists {
				httpserver.Answer(c, err, http.StatusUnprocessableEntity, responses.CodeWrongAccessToken)
				return
			}
			logger.Log("%s: access token ok", op)
			c.Request.Header.Set(accessTokenSubjectKey, string(subject))
			c.Next()
		},
		AccessToken: func(c *gin.Context) UserId {
			return UserId(c.Request.Header.Get(accessTokenSubjectKey))
		},
	}
}

func extractBearerToken(c *gin.Context) string {
	token := c.Request.Header.Get("Authorization")
	if token == "" {
		return ""
	}
	jwtToken := strings.Split(token, " ")
	if len(jwtToken) != 2 {
		return ""
	}
	return jwtToken[1]
}
