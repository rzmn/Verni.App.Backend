package friends

import (
	"database/sql"
	"fmt"
	"log"
	"verni/internal/repositories"

	_ "github.com/lib/pq"
)

type UserId string

type FriendStatus int

const (
	FriendStatusNo = iota
	FriendStatusSubscriber
	FriendStatusSubscription
	FriendStatusFriend
	FriendStatusMe
)

type Repository interface {
	GetFriends(userId UserId) ([]UserId, error)
	GetSubscribers(userId UserId) ([]UserId, error)
	GetSubscriptions(userId UserId) ([]UserId, error)
	GetStatuses(sender UserId, ids []UserId) (map[UserId]FriendStatus, error)

	HasFriendRequest(sender UserId, target UserId) (bool, error)
	StoreFriendRequest(sender UserId, target UserId) repositories.MutationWorkItem
	RemoveFriendRequest(sender UserId, target UserId) repositories.MutationWorkItem
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
