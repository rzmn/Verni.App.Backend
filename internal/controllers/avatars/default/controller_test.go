package defaultController_test

import (
	"errors"
	"testing"
	"verni/internal/controllers/avatars"
	defaultController "verni/internal/controllers/avatars/default"
	"verni/internal/repositories/images"
	images_mock "verni/internal/repositories/images/mock"
	"verni/internal/services/logging"
)

func TestGetAvatarsCannotGetFromRepository(t *testing.T) {
	repository := images_mock.RepositoryMock{
		GetImagesBase64Impl: func(ids []images.ImageId) ([]images.Image, error) {
			return []images.Image{}, errors.New("some error")
		},
	}
	controller := defaultController.New(&repository, logging.TestService())
	_, err := controller.GetAvatars([]avatars.AvatarId{})
	if err == nil {
		t.Fatalf("`GetAvatars` should fail with err, found nil")
	}
	if err.Code != avatars.GetAvatarsErrorInternal {
		t.Fatalf("`GetAvatars` should fail with code `internal`, found %v", err)
	}
}

func TestGetAvatarsOk(t *testing.T) {
	repository := images_mock.RepositoryMock{
		GetImagesBase64Impl: func(ids []images.ImageId) ([]images.Image, error) {
			return []images.Image{}, nil
		},
	}
	controller := defaultController.New(&repository, logging.TestService())
	_, err := controller.GetAvatars([]avatars.AvatarId{})
	if err != nil {
		t.Fatalf("`GetAvatars` should not fail with err, found %v", err)
	}
}
