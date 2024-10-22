package images

import (
	"fmt"
	"log"
	"strings"
	"verni/internal/common"
	"verni/internal/db"
	"verni/internal/repositories"

	"github.com/google/uuid"
)

type postgresRepository struct {
	db db.DB
}

func (c *postgresRepository) UploadImageBase64(base64 string) repositories.MutationWorkItemWithReturnValue[ImageId] {
	id := ImageId(uuid.New().String())
	return repositories.MutationWorkItemWithReturnValue[ImageId]{
		Perform: func() (ImageId, error) {
			return id, c.uploadImageBase64(id, base64)
		},
		Rollback: func() error {
			return c.removeImage(id)
		},
	}
}

func (c *postgresRepository) removeImage(id ImageId) error {
	const op = "repositories.images.postgresRepository.removeImage"
	log.Printf("%s: start[id=%s]", op, id)
	query := `DELETE FROM images WHERE id = $1;`
	_, err := c.db.Exec(query, string(id))
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return err
	}
	log.Printf("%s: success[id=%s]", op, id)
	return nil
}

func (c *postgresRepository) uploadImageBase64(id ImageId, base64 string) error {
	const op = "repositories.images.postgresRepository.uploadImageBase64"
	log.Printf("%s: start[id=%s]", op, id)
	query := `INSERT INTO images(id, base64) VALUES ($1, $2);`
	_, err := c.db.Exec(query, string(id), base64)
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return err
	}
	log.Printf("%s: success[id=%s]", op, id)
	return nil
}

func (c *postgresRepository) GetImagesBase64(ids []ImageId) ([]Image, error) {
	const op = "repositories.images.postgresRepository.GetImagesBase64"
	log.Printf("%s: start", op)
	argsList := strings.Join(common.Map(ids, func(id ImageId) string {
		return fmt.Sprintf("'%s'", id)
	}), ",")
	query := fmt.Sprintf(`SELECT id, base64 FROM images WHERE id IN (%s);`, argsList)
	rows, err := c.db.Query(query)
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return []Image{}, err
	}
	defer rows.Close()
	images := []Image{}
	for rows.Next() {
		var id string
		var base64 string
		if err := rows.Scan(&id, &base64); err != nil {
			log.Printf("%s: failed to perform scan err: %v", op, err)
			return []Image{}, err
		}
		images = append(images, Image{
			Id:     ImageId(id),
			Base64: base64,
		})
	}
	if err := rows.Err(); err != nil {
		log.Printf("%s: found rows err: %v", op, err)
		return []Image{}, err
	}
	log.Printf("%s: success", op)
	return images, nil
}
