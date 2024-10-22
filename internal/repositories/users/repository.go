package users

import (
	"database/sql"
	"fmt"
	"log"
	"verni/internal/repositories"
)

type UserId string
type User struct {
	Id          UserId
	DisplayName string
}

type Repository interface {
	GetUsers(ids []UserId) ([]User, error)
	SearchUsers(query string) ([]User, error)
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
