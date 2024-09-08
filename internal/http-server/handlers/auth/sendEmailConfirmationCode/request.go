package sendEmailConfirmationCode

import (
	"accounty/internal/http-server/responses"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RequestHandler interface {
	Handle(c *gin.Context, request Request) *Error
}

func handleError(c *gin.Context, err Error) {
	switch err.Code {
	case responses.CodeNotDelivered, responses.CodeAlreadyConfirmed:
		c.JSON(http.StatusConflict, Failure(err))
	case responses.CodeBadRequest:
		c.JSON(http.StatusBadRequest, Failure(err))
	default:
		c.JSON(http.StatusInternalServerError, Failure(err))
	}
}

func New(requestHandler RequestHandler) func(c *gin.Context) {
	return func(c *gin.Context) {
		const op = "handlers.auth.sendEmailConfirmationCode"
		var request Request
		if err := c.BindJSON(&request); err != nil {
			handleError(c, ErrBadRequest(fmt.Sprintf("%s: request failed %v", op, err)))
			return
		}
		if err := requestHandler.Handle(c, request); err != nil {
			handleError(c, *err)
			return
		}
		c.JSON(http.StatusCreated, Success())
	}
}
