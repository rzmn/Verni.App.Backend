package pushNotifications

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
	StorePushToken(uid UserId, token string) repositories.MutationWorkItem
	GetPushToken(uid UserId) (*string, error)
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
