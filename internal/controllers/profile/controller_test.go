package profile_test

import (
	"errors"
	"testing"
	"verni/internal/controllers/profile"
	"verni/internal/repositories"
	"verni/internal/repositories/auth"
	auth_mock "verni/internal/repositories/auth/mock"
	friends_mock "verni/internal/repositories/friends/mock"
	"verni/internal/repositories/images"
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

func TestGetInfoGetCredentialsFailed(t *testing.T) {
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

func TestUpdateDisplayNameWrongFormat(t *testing.T) {
	authRepository := auth_mock.RepositoryMock{}
	imagesRepository := images_mock.RepositoryMock{}
	usersRepository := users_mock.RepositoryMock{}
	friendsRepository := friends_mock.RepositoryMock{}
	formatValidation := formatValidation_mock.ServiceMock{
		ValidateDisplayNameFormatImpl: func(name string) error {
			return errors.New("some error")
		},
	}
	controller := profile.DefaultController(&authRepository, &imagesRepository, &usersRepository, &friendsRepository, &formatValidation)
	err := controller.UpdateDisplayName(uuid.New().String(), profile.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`UpdateDisplayName` should be failed, found no err")
	}
	if err.Code != profile.UpdateDisplayNameErrorWrongFormat {
		t.Fatalf("`UpdateDisplayName` should fail with `wrong format`, found %v", err)
	}
}

func TestUpdateDisplayNameUpdateFailed(t *testing.T) {
	authRepository := auth_mock.RepositoryMock{}
	imagesRepository := images_mock.RepositoryMock{}
	usersRepository := users_mock.RepositoryMock{
		UpdateDisplayNameImpl: func(name string, id users.UserId) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return errors.New("some error")
				},
			}
		},
	}
	friendsRepository := friends_mock.RepositoryMock{}
	formatValidation := formatValidation_mock.ServiceMock{
		ValidateDisplayNameFormatImpl: func(name string) error {
			return nil
		},
	}
	controller := profile.DefaultController(&authRepository, &imagesRepository, &usersRepository, &friendsRepository, &formatValidation)
	err := controller.UpdateDisplayName(uuid.New().String(), profile.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`UpdateDisplayName` should be failed, found no err")
	}
	if err.Code != profile.UpdateDisplayNameErrorInternal {
		t.Fatalf("`UpdateDisplayName` should fail with `internal`, found %v", err)
	}
}

func TestUpdateDisplayNameOk(t *testing.T) {
	storeCalls := 0
	authRepository := auth_mock.RepositoryMock{}
	imagesRepository := images_mock.RepositoryMock{}
	usersRepository := users_mock.RepositoryMock{
		UpdateDisplayNameImpl: func(name string, id users.UserId) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					storeCalls += 1
					return nil
				},
			}
		},
	}
	friendsRepository := friends_mock.RepositoryMock{}
	formatValidation := formatValidation_mock.ServiceMock{
		ValidateDisplayNameFormatImpl: func(name string) error {
			return nil
		},
	}
	controller := profile.DefaultController(&authRepository, &imagesRepository, &usersRepository, &friendsRepository, &formatValidation)
	err := controller.UpdateDisplayName(uuid.New().String(), profile.UserId(uuid.New().String()))
	if err != nil {
		t.Fatalf("`UpdateDisplayName` should not be failed, found err %v", err)
	}
	if storeCalls != 1 {
		t.Fatalf("store new display name should be called once found %d", storeCalls)
	}
}

func TestUpdateAvatarFailedToUploadData(t *testing.T) {
	authRepository := auth_mock.RepositoryMock{}
	imagesRepository := images_mock.RepositoryMock{
		UploadImageBase64Impl: func(base64 string) repositories.MutationWorkItemWithReturnValue[images.ImageId] {
			return repositories.MutationWorkItemWithReturnValue[images.ImageId]{
				Perform: func() (images.ImageId, error) {
					return images.ImageId(uuid.New().String()), errors.New("some error")
				},
			}
		},
	}
	usersRepository := users_mock.RepositoryMock{}
	friendsRepository := friends_mock.RepositoryMock{}
	formatValidation := formatValidation_mock.ServiceMock{}
	controller := profile.DefaultController(&authRepository, &imagesRepository, &usersRepository, &friendsRepository, &formatValidation)
	_, err := controller.UpdateAvatar(uuid.New().String(), profile.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`UpdateAvatar` should be failed, found no err")
	}
	if err.Code != profile.UpdateAvatarErrorInternal {
		t.Fatalf("`UpdateAvatar` should fail with `internal`, found %v", err)
	}
}

func TestUpdateAvatarFailedToStoreAvatar(t *testing.T) {
	avatarId := images.ImageId(uuid.New().String())
	uploadCalls := 0
	rollbackUploadCalls := 0
	authRepository := auth_mock.RepositoryMock{}
	imagesRepository := images_mock.RepositoryMock{
		UploadImageBase64Impl: func(base64 string) repositories.MutationWorkItemWithReturnValue[images.ImageId] {
			return repositories.MutationWorkItemWithReturnValue[images.ImageId]{
				Perform: func() (images.ImageId, error) {
					uploadCalls += 1
					return avatarId, nil
				},
				Rollback: func() error {
					rollbackUploadCalls += 1
					return nil
				},
			}
		},
	}
	usersRepository := users_mock.RepositoryMock{
		UpdateAvatarIdImpl: func(avatarId *users.AvatarId, id users.UserId) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return errors.New("some error")
				},
			}
		},
	}
	friendsRepository := friends_mock.RepositoryMock{}
	formatValidation := formatValidation_mock.ServiceMock{}
	controller := profile.DefaultController(&authRepository, &imagesRepository, &usersRepository, &friendsRepository, &formatValidation)
	_, err := controller.UpdateAvatar(uuid.New().String(), profile.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`UpdateAvatar` should be failed, found no err")
	}
	if err.Code != profile.UpdateAvatarErrorInternal {
		t.Fatalf("`UpdateAvatar` should fail with `internal`, found %v", err)
	}
	if uploadCalls != 1 || rollbackUploadCalls != 1 {
		t.Fatalf("update should be called once and then rolled back once, found %d %d", uploadCalls, rollbackUploadCalls)
	}
}

func TestUpdateAvatarOk(t *testing.T) {
	avatarId := images.ImageId(uuid.New().String())
	uploadCalls := 0
	updateCalls := 0
	authRepository := auth_mock.RepositoryMock{}
	imagesRepository := images_mock.RepositoryMock{
		UploadImageBase64Impl: func(base64 string) repositories.MutationWorkItemWithReturnValue[images.ImageId] {
			return repositories.MutationWorkItemWithReturnValue[images.ImageId]{
				Perform: func() (images.ImageId, error) {
					uploadCalls += 1
					return avatarId, nil
				},
			}
		},
	}
	usersRepository := users_mock.RepositoryMock{
		UpdateAvatarIdImpl: func(avatarIdToUpdate *users.AvatarId, id users.UserId) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					updateCalls += 1
					if *avatarIdToUpdate != users.AvatarId(avatarId) {
						t.Fatalf("should update same avatar is as uploaded")
					}
					return nil
				},
			}
		},
	}
	friendsRepository := friends_mock.RepositoryMock{}
	formatValidation := formatValidation_mock.ServiceMock{}
	controller := profile.DefaultController(&authRepository, &imagesRepository, &usersRepository, &friendsRepository, &formatValidation)
	_, err := controller.UpdateAvatar(uuid.New().String(), profile.UserId(uuid.New().String()))
	if err != nil {
		t.Fatalf("`UpdateAvatar` should not be failed, found err %v", err)
	}
	if uploadCalls != 1 || updateCalls != 1 {
		t.Fatalf("update should be called once and then run update user data, found %d %d", uploadCalls, updateCalls)
	}
}
