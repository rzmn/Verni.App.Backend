package users

import (
	"log"

	"github.com/gin-gonic/gin"

	"accounty/internal/auth/jwt"
	"accounty/internal/storage"

	"accounty/internal/http-server/handlers/users/get"
	"accounty/internal/http-server/handlers/users/search"
	"accounty/internal/http-server/helpers"
	"accounty/internal/http-server/middleware"
)

type getRequestHandler struct {
	storage storage.Storage
}

func (h *getRequestHandler) Handle(c *gin.Context, request get.Request) ([]storage.User, *get.Error) {
	const op = "router.users.getRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	token := helpers.ExtractBearerToken(c)
	subject, err := jwt.GetAccessTokenSubject(token)
	if err != nil || subject == nil {
		log.Printf("%s: cannot get access token %v", op, err)
		outError := get.ErrInternal()
		return nil, &outError
	}
	users, err := h.storage.GetUsers(storage.UserId(*subject), request.Ids)
	if err != nil {
		log.Printf("%s: cannot read from db %v", op, err)
		outError := get.ErrInternal()
		return nil, &outError
	}
	return users, nil
}

type searchRequestHandler struct {
	storage storage.Storage
}

func (h *searchRequestHandler) Handle(c *gin.Context, request search.Request) ([]storage.User, *search.Error) {
	const op = "router.users.searchRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	token := helpers.ExtractBearerToken(c)
	subject, err := jwt.GetAccessTokenSubject(token)
	if err != nil || subject == nil {
		log.Printf("%s: cannot get access token %v", op, err)
		outError := search.ErrInternal()
		return nil, &outError
	}
	users, err := h.storage.SearchUsers(storage.UserId(*subject), request.Query)
	if err != nil {
		log.Printf("%s: cannot read from db %v", op, err)
		outError := search.ErrInternal()
		return nil, &outError
	}
	return users, nil
}

func RegisterRoutes(e *gin.Engine, storage storage.Storage) {
	group := e.Group("/users", middleware.EnsureLoggedIn(storage))
	group.GET("/get", get.New(&getRequestHandler{storage: storage}))
	group.GET("/search", search.New(&searchRequestHandler{storage: storage}))
}
