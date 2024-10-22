package pushNotifications

import (
	"verni/internal/db"
	"verni/internal/repositories"

	_ "github.com/lib/pq"
)

type UserId string

type Repository interface {
	StorePushToken(uid UserId, token string) repositories.MutationWorkItem
	GetPushToken(uid UserId) (*string, error)
}

func PostgresRepository(db db.DB) Repository {
	return &postgresRepository{
		db: db,
	}
}
