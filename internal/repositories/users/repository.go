package users

import "verni/internal/db"

type UserId string
type User struct {
	Id          UserId
	DisplayName string
}

type Repository interface {
	GetUsers(ids []UserId) ([]User, error)
	SearchUsers(query string) ([]User, error)
}

func PostgresRepository(db db.DB) Repository {
	return &postgresRepository{
		db: db,
	}
}
