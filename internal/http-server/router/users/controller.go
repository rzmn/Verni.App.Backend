package users

import (
	"accounty/internal/common"
	"accounty/internal/storage"
)

type UserId storage.UserId
type User storage.User

type Controller interface {
	Get(ids []UserId, sender UserId) ([]User, *common.CodeBasedError[GetUsersErrorCode])
	Search(query string, sender UserId) ([]User, *common.CodeBasedError[SearchUsersErrorCode])
}

func DefaultController(storage storage.Storage) Controller {
	return &defaultController{
		storage: storage,
	}
}
