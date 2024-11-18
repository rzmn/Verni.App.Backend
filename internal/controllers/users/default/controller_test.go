package defaultController_test

import (
	"errors"
	"testing"
	"verni/internal/controllers/users"
	defaultController "verni/internal/controllers/users/default"
	friendsRepository "verni/internal/repositories/friends"
	friends_mock "verni/internal/repositories/friends/mock"
	usersRepository "verni/internal/repositories/users"
	users_mock "verni/internal/repositories/users/mock"
	standartOutputLoggingService "verni/internal/services/logging/standartOutput"

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
	controller := defaultController.New(&usersRepository, &friendsRepository, standartOutputLoggingService.New())
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
	controller := defaultController.New(&usersRepository, &friendsRepository, standartOutputLoggingService.New())
	_, err := controller.Get([]users.UserId{users.UserId(uuid.New().String())}, users.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`Get` should be failed, found no err")
	}
	if err.Code != users.GetUsersErrorInternal {
		t.Fatalf("`Get` should be failed with `internal`, found err %v", err)
	}
}

func TestGetUsersMissingUserStatus(t *testing.T) {
	userA := usersRepository.User{
		Id: usersRepository.UserId(uuid.New().String()),
	}
	userB := usersRepository.User{
		Id: usersRepository.UserId(uuid.New().String()),
	}
	usersRepository := users_mock.RepositoryMock{
		GetUsersImpl: func(ids []usersRepository.UserId) ([]usersRepository.User, error) {
			return []usersRepository.User{userA, userB}, nil
		},
	}
	friendsRepository := friends_mock.RepositoryMock{
		GetStatusesImpl: func(sender friendsRepository.UserId, ids []friendsRepository.UserId) (map[friendsRepository.UserId]friendsRepository.FriendStatus, error) {
			return map[friendsRepository.UserId]friendsRepository.FriendStatus{
				friendsRepository.UserId(userA.Id): friendsRepository.FriendStatusSubscriber,
			}, nil
		},
	}
	controller := defaultController.New(&usersRepository, &friendsRepository, standartOutputLoggingService.New())
	_, err := controller.Get([]users.UserId{users.UserId(uuid.New().String())}, users.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`Get` should be failed, found no err")
	}
	if err.Code != users.GetUsersUserNotFound {
		t.Fatalf("`Get` should be failed with `not found`, found err %v", err)
	}
}

func TestGetUsersOk(t *testing.T) {
	userA := usersRepository.User{
		Id: usersRepository.UserId(uuid.New().String()),
	}
	userB := usersRepository.User{
		Id: usersRepository.UserId(uuid.New().String()),
	}
	usersRepository := users_mock.RepositoryMock{
		GetUsersImpl: func(ids []usersRepository.UserId) ([]usersRepository.User, error) {
			return []usersRepository.User{userA, userB}, nil
		},
	}
	friendsRepository := friends_mock.RepositoryMock{
		GetStatusesImpl: func(sender friendsRepository.UserId, ids []friendsRepository.UserId) (map[friendsRepository.UserId]friendsRepository.FriendStatus, error) {
			return map[friendsRepository.UserId]friendsRepository.FriendStatus{
				friendsRepository.UserId(userA.Id): friendsRepository.FriendStatusSubscriber,
				friendsRepository.UserId(userB.Id): friendsRepository.FriendStatusSubscription,
			}, nil
		},
	}
	controller := defaultController.New(&usersRepository, &friendsRepository, standartOutputLoggingService.New())
	_, err := controller.Get([]users.UserId{users.UserId(uuid.New().String())}, users.UserId(uuid.New().String()))
	if err != nil {
		t.Fatalf("`Get` should not be failed, found err: %v", err)
	}
}
