package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"verni/internal/auth/jwt"
	"verni/internal/storage"

	"verni/internal/http-server/responses"
)

const (
	LoggedInSubjectKey = "verni-subject"
)

func EnsureLoggedIn(s storage.Storage, jwtService jwt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.friends.ensureLoggedInMiddleware"

		log.Printf("%s: validating access token", op)
		token := jwt.AccessToken(extractBearerToken(c))

		if err := jwtService.ValidateAccessToken(token); err != nil {
			log.Printf("%s: failed to validate token %v", op, err)
			switch err.Code {
			case jwt.CodeTokenExpired:
				c.AbortWithStatusJSON(http.StatusUnauthorized, responses.Failure(responses.Error{
					Code: responses.CodeTokenExpired,
				}))
			case jwt.CodeTokenInvalid:
				c.AbortWithStatusJSON(http.StatusUnprocessableEntity, responses.Failure(responses.Error{
					Code: responses.CodeWrongAccessToken,
				}))
			default:
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.Failure(responses.Error{
					Code: responses.CodeInternal,
				}))
			}
			return
		}
		subject, getSubjectError := jwtService.GetAccessTokenSubject(token)
		if getSubjectError != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.Failure(responses.Error{
				Code: responses.CodeInternal,
			}))
			return
		}
		exists, err := s.IsUserExists(storage.UserId(subject))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.Failure(responses.Error{
				Code: responses.CodeInternal,
			}))
			return
		}
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, responses.Failure(responses.Error{
				Code: responses.CodeWrongAccessToken,
			}))
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
