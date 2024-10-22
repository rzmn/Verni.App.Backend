package users

import (
	"verni/internal/common"
	usersRepository "verni/internal/repositories/users"
)

type UserId string
type Repository usersRepository.Repository

type User struct {
}

type Controller interface {
	Get(ids []UserId, sender UserId) ([]User, *common.CodeBasedError[GetUsersErrorCode])
	Search(query string, sender UserId) ([]User, *common.CodeBasedError[SearchUsersErrorCode])
}

func DefaultController(repository Repository) Controller {
	return &defaultController{
		repository: repository,
	}
}
