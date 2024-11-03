package auth

import (
	"verni/internal/db"
	"verni/internal/repositories"
	"verni/internal/services/logging"
)

type UserId string
type UserInfo struct {
	UserId        UserId
	Email         string
	PasswordHash  string
	RefreshToken  string
	EmailVerified bool
}
type Repository interface {
	CreateUser(uid UserId, email string, password string, refreshToken string) repositories.MutationWorkItem
	MarkUserEmailValidated(uid UserId) repositories.MutationWorkItem
	IsUserExists(uid UserId) (bool, error)

	CheckCredentials(email string, password string) (bool, error)
	GetUserIdByEmail(email string) (*UserId, error)

	UpdateRefreshToken(uid UserId, token string) repositories.MutationWorkItem
	UpdatePassword(uid UserId, newPassword string) repositories.MutationWorkItem
	UpdateEmail(uid UserId, newEmail string) repositories.MutationWorkItem

	GetUserInfo(uid UserId) (UserInfo, error)
}

func PostgresRepository(db db.DB, logger logging.Service) Repository {
	return &postgresRepository{
		db:     db,
		logger: logger,
	}
}
