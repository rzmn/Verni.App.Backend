package postgresDb

import (
	"database/sql"
	"fmt"
	"verni/internal/db"
	"verni/internal/services/logging"

	_ "github.com/lib/pq"
)

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DbName   string `json:"dbName"`
}

func Postgres(config PostgresConfig, logger logging.Service) (db.DB, error) {
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
		logger.LogInfo("%s: open db failed err: %v", op, err)
		return nil, err
	}
	return db, nil
}
