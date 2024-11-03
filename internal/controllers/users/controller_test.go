package users_test

import (
	"errors"
	"testing"
	"verni/internal/controllers/users"
	friendsRepository "verni/internal/repositories/friends"
	friends_mock "verni/internal/repositories/friends/mock"
	usersRepository "verni/internal/repositories/users"
	users_mock "verni/internal/repositories/users/mock"
	"verni/internal/services/logging"

	"github.com/google/uuid"
)

func TestGetUsersFailed(t *testing.T) {
	usersRepository := users_mock.RepositoryMock{
		GetUsersImpl: func(ids []usersRepository.UserId) ([]usersRepository.User, error) {
			return []usersRepository.User{}, errors.New("some error")
		},
	}
	friendsRepository := friends_mock.RepositoryMock{
		GetStatusesImpl: func(sender friendsRepository.UserId, ids []friendsRepository.UserId) (map[friendsRepository.UserId]friendsRepository.FriendStatus, error) {
			return map[friendsRepository.UserId]friendsRepository.FriendStatus{}, nil
		},
	}
	controller := users.DefaultController(&usersRepository, &friendsRepository, logging.TestService())
	_, err := controller.Get([]users.UserId{users.UserId(uuid.New().String())}, users.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`Get` should be failed, found no err")
	}
	if err.Code != users.GetUsersErrorInternal {
		t.Fatalf("`Get` should be failed with `internal`, found err %v", err)
	}
}

func TestGetFriendStatusesFailed(t *testing.T) {
	usersRepository := users_mock.RepositoryMock{
		GetUsersImpl: func(ids []usersRepository.UserId) ([]usersRepository.User, error) {
			return []usersRepository.User{}, nil
		},
	}
	friendsRepository := friends_mock.RepositoryMock{
		GetStatusesImpl: func(sender friendsRepository.UserId, ids []friendsRepository.UserId) (map[friendsRepository.UserId]friendsRepository.FriendStatus, error) {
			return map[friendsRepository.UserId]friendsRepository.FriendStatus{}, errors.New("some error")
		},
	}
	controller := users.DefaultController(&usersRepository, &friendsRepository, logging.TestService())
	_, err := controller.Get([]users.UserId{users.UserId(uuid.New().String())}, users.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`Get` should be failed, found no err")
	}
	if err.Code != users.GetUsersErrorInternal {
		t.Fatalf("`Get` should be failed with `internal`, found err %v", err)
	}
}
