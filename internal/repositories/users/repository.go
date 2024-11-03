package users

import (
	"verni/internal/db"
	"verni/internal/repositories"
	"verni/internal/services/logging"
)

type UserId string
type AvatarId string
type User struct {
	Id          UserId
	DisplayName string
	AvatarId    *AvatarId
}

type Repository interface {
	StoreUser(user User) repositories.MutationWorkItem
	GetUsers(ids []UserId) ([]User, error)
	UpdateDisplayName(name string, id UserId) repositories.MutationWorkItem
	UpdateAvatarId(avatarId *AvatarId, id UserId) repositories.MutationWorkItem
}

func PostgresRepository(db db.DB, logger logging.Service) Repository {
	return &postgresRepository{
		db:     db,
		logger: logger,
	}
}
