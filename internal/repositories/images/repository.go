package images

import (
	"github.com/rzmn/governi/internal/repositories"
)

type ImageId string
type Image struct {
	Id     ImageId
	Base64 string
}

type Repository interface {
	UploadImageBase64(base64 string) repositories.MutationWorkItemWithReturnValue[ImageId]
	GetImagesBase64(ids []ImageId) ([]Image, error)
}
