package images

import (
	"verni/internal/db"
	"verni/internal/repositories"
	"verni/internal/services/logging"
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

func PostgresRepository(db db.DB, logger logging.Service) Repository {
	return &postgresRepository{
		db:     db,
		logger: logger,
	}
}
