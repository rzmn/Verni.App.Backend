package signup

import (
	"fmt"

	"net/http"

	"github.com/gin-gonic/gin"

	"accounty/internal/http-server/responses"
	"accounty/internal/storage"
)

type RequestHandler interface {
	Validate(request Request) *Error
	Handle(request Request) (*storage.AuthenticatedSession, *Error)
}

func handleError(c *gin.Context, err Error) {
	switch err.Code {
	case responses.CodeWrongFormat:
		c.JSON(http.StatusUnprocessableEntity, Failure(err))
	case responses.CodeAlreadyTaken:
		c.JSON(http.StatusConflict, Failure(err))
	case responses.CodeBadRequest:
		c.JSON(http.StatusBadRequest, Failure(err))
	default:
		c.JSON(http.StatusInternalServerError, Failure(err))
	}
}

func New(requestHandler RequestHandler) func(c *gin.Context) {
	return func(c *gin.Context) {
		const op = "handlers.auth.signup"
		var request Request
		if err := c.BindJSON(&request); err != nil {
			handleError(c, ErrBadRequest(fmt.Sprintf("%s: request failed %v", op, err)))
			return
		}
		if err := requestHandler.Validate(request); err != nil {
			handleError(c, *err)
			return
		}
		token, err := requestHandler.Handle(request)
		if err != nil {
			handleError(c, *err)
			return
		}
		if token == nil {
			handleError(c, ErrInternal())
			return
		}
		c.JSON(http.StatusCreated, Success(*token))
	}
}
