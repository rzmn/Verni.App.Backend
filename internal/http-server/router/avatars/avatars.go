package avatars

import (
	"accounty/internal/http-server/handlers/avatars/get"
	"accounty/internal/storage"
	"log"

	"github.com/gin-gonic/gin"
)

type getRequestHandler struct {
	storage storage.Storage
}

func (h *getRequestHandler) Handle(c *gin.Context, request get.Request) (map[storage.AvatarId]storage.AvatarData, *get.Error) {
	const op = "router.avatars.getRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	avatars, err := h.storage.GetAvatarsBase64(request.Ids)
	if err != nil {
		log.Printf("%s: cannot read from db %v", op, err)
		outError := get.ErrInternal()
		return nil, &outError
	}
	return avatars, nil
}

func RegisterRoutes(e *gin.Engine, storage storage.Storage) {
	group := e.Group("/avatars")
	group.GET("/get", get.New(&getRequestHandler{storage: storage}))
}
