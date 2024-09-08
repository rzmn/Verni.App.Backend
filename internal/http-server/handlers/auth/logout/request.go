package logout

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type RequestHandler interface {
	Handle(c *gin.Context) *Error
}

func handleError(c *gin.Context, err Error) {
	switch err.Code {
	default:
		c.JSON(http.StatusInternalServerError, Failure(err))
	}
}

func New(requestHandler RequestHandler) func(c *gin.Context) {
	return func(c *gin.Context) {
		if err := requestHandler.Handle(c); err != nil {
			handleError(c, *err)
			return
		}
		c.JSON(http.StatusCreated, Success())
	}
}
