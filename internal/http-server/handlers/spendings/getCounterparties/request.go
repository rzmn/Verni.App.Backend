package getCounterparties

import (
	"accounty/internal/http-server/responses"
	"accounty/internal/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RequestHandler interface {
	Handle(c *gin.Context, request Request) ([]storage.SpendingsPreview, *Error)
}

func handleError(c *gin.Context, err Error) {
	switch err.Code {
	case responses.CodeBadRequest:
		c.JSON(http.StatusBadRequest, Failure(err))
	default:
		c.JSON(http.StatusInternalServerError, Failure(err))
	}
}

func New(requestHandler RequestHandler) func(c *gin.Context) {
	return func(c *gin.Context) {
		previews, err := requestHandler.Handle(c, Request{})
		if err != nil {
			handleError(c, *err)
			return
		}
		c.JSON(http.StatusOK, Success(previews))
	}
}
