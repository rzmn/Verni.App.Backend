package profile

import (
	"accounty/internal/auth/jwt"
	"accounty/internal/http-server/handlers/profile/getInfo"
	"accounty/internal/http-server/handlers/profile/setAvatar"
	"accounty/internal/http-server/handlers/profile/setDisplayName"
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
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	info, err := h.storage.GetAccountInfo(subject)
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

func (h *setAvatarRequestHandler) Handle(c *gin.Context, request setAvatar.Request) (storage.AvatarId, *setAvatar.Error) {
	const op = "router.profile.setAvatarRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	aid, err := h.storage.StoreAvatarBase64(subject, request.DataBase64)
	if err != nil {
		log.Printf("%s: cannot store avatar data %v", op, err)
		outError := setAvatar.ErrInternal()
		return aid, &outError
	}
	return aid, nil
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
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	if err := h.storage.StoreDisplayName(subject, request.DisplayName); err != nil {
		log.Printf("%s: cannot store avatar data %v", op, err)
		outError := setDisplayName.ErrInternal()
		return &outError
	}
	return nil
}

func RegisterRoutes(e *gin.Engine, storage storage.Storage, jwtService jwt.Service) {
	group := e.Group("/profile", middleware.EnsureLoggedIn(storage, jwtService))
	group.GET("/getInfo", getInfo.New(&getInfoRequestHandler{storage: storage}))
	group.PUT("/setAvatar", setAvatar.New(&setAvatarRequestHandler{storage: storage}))
	group.PUT("/setDisplayName", setDisplayName.New(&setDisplayNameRequestHandler{storage: storage}))
}
