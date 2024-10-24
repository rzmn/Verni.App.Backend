package verification

import (
	"verni/internal/db"
	"verni/internal/repositories"
)

type Repository interface {
	StoreEmailVerificationCode(email string, code string) repositories.MutationWorkItem
	GetEmailVerificationCode(email string) (*string, error)
	RemoveEmailVerificationCode(email string) repositories.MutationWorkItem
}

func PostgresRepository(db db.DB) Repository {
	return &postgresRepository{
		db: db,
	}
}
