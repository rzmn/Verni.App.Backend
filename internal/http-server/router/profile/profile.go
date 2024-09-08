package profile

import (
	"accounty/internal/auth/jwt"
	"accounty/internal/http-server/handlers/profile/getInfo"
	"accounty/internal/http-server/handlers/profile/setAvatar"
	"accounty/internal/http-server/handlers/profile/setDisplayName"
	"accounty/internal/http-server/helpers"
	"accounty/internal/http-server/middleware"
	"accounty/internal/storage"
	"log"
	"regexp"

	"github.com/gin-gonic/gin"
)

type getInfoRequestHandler struct {
	storage storage.Storage
}

func (h *getInfoRequestHandler) Handle(c *gin.Context) (storage.ProfileInfo, *getInfo.Error) {
	const op = "router.profile.getInfoRequestHandler.Handle"
	token := helpers.ExtractBearerToken(c)

	subject, err := jwt.GetAccessTokenSubject(token)
	if err != nil || subject == nil {
		log.Printf("%s: cannot get access token %v", op, err)
		outError := getInfo.ErrInternal()
		return storage.ProfileInfo{}, &outError
	}
	info, err := h.storage.GetAccountInfo(storage.UserId(*subject))
	if err != nil {
		log.Printf("%s: cannot get host info %v", op, err)
		outError := getInfo.ErrInternal()
		return storage.ProfileInfo{}, &outError
	}
	if info == nil {
		log.Printf("%s: profile not found", op)
		outError := getInfo.ErrInternal()
		return storage.ProfileInfo{}, &outError
	}
	return *info, nil
}

type setAvatarRequestHandler struct {
	storage storage.Storage
}

func (h *setAvatarRequestHandler) Validate(c *gin.Context, request setAvatar.Request) *setAvatar.Error {
	return nil
}

func (h *setAvatarRequestHandler) Handle(c *gin.Context, request setAvatar.Request) *setAvatar.Error {
	const op = "router.profile.setAvatarRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	token := helpers.ExtractBearerToken(c)
	subject, err := jwt.GetAccessTokenSubject(token)
	if err != nil || subject == nil {
		log.Printf("%s: cannot get access token %v", op, err)
		outError := setAvatar.ErrInternal()
		return &outError
	}
	if err := h.storage.StoreAvatarBase64(storage.UserId(*subject), request.DataBase64); err != nil {
		log.Printf("%s: cannot store avatar data %v", op, err)
		outError := setAvatar.ErrInternal()
		return &outError
	}
	return nil
}

type setDisplayNameRequestHandler struct {
	storage storage.Storage
}

func (h *setDisplayNameRequestHandler) Validate(c *gin.Context, request setDisplayName.Request) *setDisplayName.Error {
	if !regexp.MustCompile(`^[A-Za-z]+$`).MatchString(request.DisplayName) {
		outError := setDisplayName.ErrWrongFormat()
		return &outError
	}
	return nil
}

func (h *setDisplayNameRequestHandler) Handle(c *gin.Context, request setDisplayName.Request) *setDisplayName.Error {
	const op = "router.profile.setDisplayNameRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	token := helpers.ExtractBearerToken(c)
	subject, err := jwt.GetAccessTokenSubject(token)
	if err != nil || subject == nil {
		log.Printf("%s: cannot get access token %v", op, err)
		outError := setDisplayName.ErrInternal()
		return &outError
	}
	if err := h.storage.StoreDisplayName(storage.UserId(*subject), request.DisplayName); err != nil {
		log.Printf("%s: cannot store avatar data %v", op, err)
		outError := setDisplayName.ErrInternal()
		return &outError
	}
	return nil
}

func RegisterRoutes(e *gin.Engine, storage storage.Storage) {
	group := e.Group("/profile", middleware.EnsureLoggedIn(storage))
	group.GET("/getInfo", getInfo.New(&getInfoRequestHandler{storage: storage}))
	group.PUT("/setAvatar", setAvatar.New(&setAvatarRequestHandler{storage: storage}))
	group.PUT("/setDisplayName", setDisplayName.New(&setDisplayNameRequestHandler{storage: storage}))
}
