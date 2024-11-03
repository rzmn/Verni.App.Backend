package pushNotifications

import (
	"verni/internal/db"
	"verni/internal/repositories"
	"verni/internal/services/logging"
)

type UserId string

type Repository interface {
	StorePushToken(uid UserId, token string) repositories.MutationWorkItem
	GetPushToken(uid UserId) (*string, error)
}

func PostgresRepository(db db.DB, logger logging.Service) Repository {
	return &postgresRepository{
		db:     db,
		logger: logger,
	}
}
