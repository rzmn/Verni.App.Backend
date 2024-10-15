package rollbackRequest

import (
	"fmt"

	"net/http"

	"github.com/gin-gonic/gin"

	"verni/internal/http-server/responses"
)

type RequestHandler interface {
	Validate(c *gin.Context, request Request) *Error
	Handle(c *gin.Context, request Request) *Error
}

func handleError(c *gin.Context, err Error) {
	switch err.Code {
	case responses.CodeNoSuchRequest, responses.CodeNoSuchUser:
		c.JSON(http.StatusConflict, Failure(err))
	case responses.CodeBadRequest:
		c.JSON(http.StatusBadRequest, Failure(err))
	default:
		c.JSON(http.StatusInternalServerError, Failure(err))
	}
}

func New(requestHandler RequestHandler) func(c *gin.Context) {
	return func(c *gin.Context) {
		const op = "handlers.friends.rollbackRequest"
		var request Request
		if err := c.BindJSON(&request); err != nil {
			handleError(c, ErrBadRequest(fmt.Sprintf("%s: request failed %v", op, err)))
			return
		}
		if err := requestHandler.Validate(c, request); err != nil {
			handleError(c, *err)
			return
		}
		if err := requestHandler.Handle(c, request); err != nil {
			handleError(c, *err)
			return
		}
		c.JSON(http.StatusCreated, Success())
	}
}
