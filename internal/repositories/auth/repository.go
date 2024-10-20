package auth

import (
	"database/sql"
	"fmt"
	"log"
	"verni/internal/repositories"
	"verni/internal/storage"

	_ "github.com/lib/pq"
)

type UserId storage.UserId
type Repository interface {
	CreateUser(uid UserId, email string, password string, refreshToken string) repositories.MutationWorkItem

	CheckCredentials(email string, password string) (bool, error)
	GetUserIdByEmail(email string) (*UserId, error)

	UpdateRefreshToken(uid UserId, token string) repositories.MutationWorkItem
	UpdatePassword(uid UserId, newPassword string) repositories.MutationWorkItem
	UpdateEmail(uid UserId, newEmail string) repositories.MutationWorkItem

	GetRefreshToken(uid UserId) (string, error)
}

func PostgresRepository(config repositories.PostgresConfig) (Repository, error) {
	const op = "repositories.friends.PostgresRepository"
	psqlConnection := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DbName,
	)
	db, err := sql.Open("postgres", psqlConnection)
	if err != nil {
		log.Printf("%s: open db failed err: %v", op, err)
		return &postgresRepository{}, err
	}
	return &postgresRepository{
		db: db,
	}, nil
}
