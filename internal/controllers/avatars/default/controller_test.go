package defaultController_test

import (
	"errors"
	"testing"

	"github.com/rzmn/Verni.App.Backend/internal/controllers/avatars"
	defaultController "github.com/rzmn/Verni.App.Backend/internal/controllers/avatars/default"
	"github.com/rzmn/Verni.App.Backend/internal/repositories/images"
	images_mock "github.com/rzmn/Verni.App.Backend/internal/repositories/images/mock"
	standartOutputLoggingService "github.com/rzmn/Verni.App.Backend/internal/services/logging/standartOutput"
)

func TestGetAvatarsCannotGetFromRepository(t *testing.T) {
	repository := images_mock.RepositoryMock{
		GetImagesBase64Impl: func(ids []images.ImageId) ([]images.Image, error) {
			return []images.Image{}, errors.New("some error")
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
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
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	_, err := controller.GetAvatars([]avatars.AvatarId{})
	if err != nil {
		t.Fatalf("`GetAvatars` should not fail with err, found %v", err)
	}
}
