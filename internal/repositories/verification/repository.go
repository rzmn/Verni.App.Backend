package verification

import (
	"verni/internal/db"
	"verni/internal/repositories"
	"verni/internal/services/logging"
)

type Repository interface {
	StoreEmailVerificationCode(email string, code string) repositories.MutationWorkItem
	GetEmailVerificationCode(email string) (*string, error)
	RemoveEmailVerificationCode(email string) repositories.MutationWorkItem
}

func PostgresRepository(db db.DB, logger logging.Service) Repository {
	return &postgresRepository{
		db:     db,
		logger: logger,
	}
}
