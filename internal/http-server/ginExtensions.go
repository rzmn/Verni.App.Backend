package httpserver

import (
	"net/http"
	"verni/internal/common"
	"verni/internal/http-server/responses"

	"github.com/gin-gonic/gin"
)

func Answer(c *gin.Context, err error, httpCode int, errorCode responses.Code) {
	c.AbortWithStatusJSON(
		httpCode,
		responses.Failure(common.NewErrorWithDescriptionValue(errorCode, err.Error())),
	)
}

func AnswerWithBadRequest(c *gin.Context, err error) {
	Answer(c, err, http.StatusBadRequest, responses.CodeBadRequest)
}

func AnswerWithUnknownError(c *gin.Context, err error) {
	Answer(c, err, http.StatusInternalServerError, responses.CodeInternal)
}
