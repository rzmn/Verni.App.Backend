package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"verni/internal/auth/jwt"
	httpserver "verni/internal/http-server"
	authRepository "verni/internal/repositories/auth"

	"verni/internal/http-server/responses"
)

const (
	LoggedInSubjectKey = "verni-subject"
)

func EnsureLoggedIn(repository authRepository.Repository, jwtService jwt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.friends.ensureLoggedInMiddleware"
		log.Printf("%s: validating access token", op)
		token := jwt.AccessToken(extractBearerToken(c))
		if err := jwtService.ValidateAccessToken(token); err != nil {
			log.Printf("%s: failed to validate token %v", op, err)
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
		log.Printf("%s: access token ok", op)
		c.Request.Header.Set(LoggedInSubjectKey, string(subject))
		c.Next()
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
