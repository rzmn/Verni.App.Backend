package httpserver

import (
	"net/http"
	"verni/internal/http-server/responses"

	"github.com/gin-gonic/gin"
)

func AnswerWithBadRequest(c *gin.Context, err error) {
	message := err.Error()
	c.JSON(
		http.StatusBadRequest,
		responses.Failure(
			responses.Error{
				Code:        responses.CodeBadRequest,
				Description: &message,
			},
		),
	)
}

func AnswerWithUnknownError(c *gin.Context, err error) {
	message := err.Error()
	c.JSON(
		http.StatusInternalServerError,
		responses.Failure(
			responses.Error{
				Code:        responses.CodeInternal,
				Description: &message,
			},
		),
	)
}
