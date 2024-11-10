package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"verni/internal/common"
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
			logger.LogInfo("%s: validating access token", op)
			token := jwt.AccessToken(extractBearerToken(c))
			if err := jwtService.ValidateAccessToken(token); err != nil {
				logger.LogInfo("%s: failed to validate token %v", op, err)
				switch err.Code {
				case jwt.CodeTokenExpired:
					c.AbortWithStatusJSON(
						http.StatusUnauthorized,
						responses.Failure(
							common.NewErrorWithDescriptionValue(
								responses.CodeTokenExpired,
								err.Error(),
							),
						),
					)
				case jwt.CodeTokenInvalid:
					c.AbortWithStatusJSON(
						http.StatusUnprocessableEntity,
						responses.Failure(
							common.NewErrorWithDescriptionValue(
								responses.CodeWrongAccessToken,
								err.Error(),
							),
						),
					)
				default:
					logger.LogError("jwt token validation failed %v", err)
					c.AbortWithStatusJSON(
						http.StatusInternalServerError,
						responses.Failure(
							common.NewErrorWithDescriptionValue(
								responses.CodeInternal,
								err.Error(),
							),
						),
					)
				}
				return
			}
			subject, getSubjectError := jwtService.GetAccessTokenSubject(token)
			if getSubjectError != nil {
				logger.LogError("jwt token get subject failed %v", getSubjectError)
				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					responses.Failure(
						common.NewErrorWithDescriptionValue(
							responses.CodeInternal,
							getSubjectError.Error(),
						),
					),
				)
				return
			}
			exists, err := repository.IsUserExists(authRepository.UserId(subject))
			if err != nil {
				logger.LogError("valid token with invalid subject - %v", err)
				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					responses.Failure(
						common.NewErrorWithDescriptionValue(
							responses.CodeInternal,
							err.Error(),
						),
					),
				)
				return
			}
			if !exists {
				c.AbortWithStatusJSON(
					http.StatusUnprocessableEntity,
					responses.Failure(
						common.NewErrorWithDescriptionValue(
							responses.CodeWrongAccessToken,
							"associated user is not exists",
						),
					),
				)
				return
			}
			logger.LogInfo("%s: access token ok", op)
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
