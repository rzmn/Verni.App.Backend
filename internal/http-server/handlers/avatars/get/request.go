package get

import (
	"accounty/internal/http-server/responses"
	"accounty/internal/storage"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RequestHandler interface {
	Handle(c *gin.Context, request Request) (map[storage.AvatarId]storage.AvatarData, *Error)
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
		const op = "handlers.avatars.get"

		requestQueryString := c.Query("data")
		log.Printf("%s: got data param: %s", op, requestQueryString)
		request := Request{}

		if err := json.Unmarshal([]byte(requestQueryString), &request); err != nil {
			handleError(c, ErrBadRequest(fmt.Sprintf("%s: request failed %v", op, err)))
			return
		}
		users, err := requestHandler.Handle(c, request)
		if err != nil {
			handleError(c, *err)
			return
		}
		c.JSON(http.StatusOK, Success(users))
	}
}
