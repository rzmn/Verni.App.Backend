package profile_test

import (
	"errors"
	"testing"
	"verni/internal/controllers/profile"
	"verni/internal/repositories/auth"
	auth_mock "verni/internal/repositories/auth/mock"
	friends_mock "verni/internal/repositories/friends/mock"
	images_mock "verni/internal/repositories/images/mock"
	"verni/internal/repositories/users"
	users_mock "verni/internal/repositories/users/mock"
	formatValidation_mock "verni/internal/services/formatValidation/mock"

	"github.com/google/uuid"
)

func TestGetInfoGetUsersFailed(t *testing.T) {
	authRepository := auth_mock.RepositoryMock{}
	imagesRepository := images_mock.RepositoryMock{}
	usersRepository := users_mock.RepositoryMock{
		GetUsersImpl: func(ids []users.UserId) ([]users.User, error) {
			return []users.User{}, errors.New("some error")
		},
	}
	friendsRepository := friends_mock.RepositoryMock{}
	formatValidation := formatValidation_mock.ServiceMock{}

	controller := profile.DefaultController(&authRepository, &imagesRepository, &usersRepository, &friendsRepository, &formatValidation)
	_, err := controller.GetProfileInfo(profile.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`GetProfileInfo` should be failed, found no err")
	}
	if err.Code != profile.GetInfoErrorInternal {
		t.Fatalf("`GetProfileInfo` should be failed with `internal`, found %v", err)
	}
}

func TestGetInfoNoUsersFound(t *testing.T) {
	authRepository := auth_mock.RepositoryMock{}
	imagesRepository := images_mock.RepositoryMock{}
	usersRepository := users_mock.RepositoryMock{
		GetUsersImpl: func(ids []users.UserId) ([]users.User, error) {
			return []users.User{}, nil
		},
	}
	friendsRepository := friends_mock.RepositoryMock{}
	formatValidation := formatValidation_mock.ServiceMock{}

	controller := profile.DefaultController(&authRepository, &imagesRepository, &usersRepository, &friendsRepository, &formatValidation)
	_, err := controller.GetProfileInfo(profile.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`GetProfileInfo` should be failed, found no err")
	}
	if err.Code != profile.GetInfoErrorInternal {
		t.Fatalf("`GetProfileInfo` should be failed with `internal`, found %v", err)
	}
}

func TestGetInfoGetCredentialsFound(t *testing.T) {
	authRepository := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid auth.UserId) (auth.UserInfo, error) {
			return auth.UserInfo{}, errors.New("some error")
		},
	}
	imagesRepository := images_mock.RepositoryMock{}
	usersRepository := users_mock.RepositoryMock{
		GetUsersImpl: func(ids []users.UserId) ([]users.User, error) {
			return []users.User{{}}, nil
		},
	}
	friendsRepository := friends_mock.RepositoryMock{}
	formatValidation := formatValidation_mock.ServiceMock{}

	controller := profile.DefaultController(&authRepository, &imagesRepository, &usersRepository, &friendsRepository, &formatValidation)
	_, err := controller.GetProfileInfo(profile.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`GetProfileInfo` should be failed, found no err")
	}
	if err.Code != profile.GetInfoErrorInternal {
		t.Fatalf("`GetProfileInfo` should be failed with `internal`, found %v", err)
	}
}

func TestGetInfoOk(t *testing.T) {
	authRepository := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid auth.UserId) (auth.UserInfo, error) {
			return auth.UserInfo{}, nil
		},
	}
	imagesRepository := images_mock.RepositoryMock{}
	usersRepository := users_mock.RepositoryMock{
		GetUsersImpl: func(ids []users.UserId) ([]users.User, error) {
			return []users.User{{}}, nil
		},
	}
	friendsRepository := friends_mock.RepositoryMock{}
	formatValidation := formatValidation_mock.ServiceMock{}

	controller := profile.DefaultController(&authRepository, &imagesRepository, &usersRepository, &friendsRepository, &formatValidation)
	_, err := controller.GetProfileInfo(profile.UserId(uuid.New().String()))
	if err != nil {
		t.Fatalf("`GetProfileInfo` should not be failed, found err %v", err)
	}
}
