package users

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"verni/internal/common"
	"verni/internal/db"
	"verni/internal/repositories"
)

type postgresRepository struct {
	db db.DB
}

func (c *postgresRepository) GetUsers(ids []UserId) ([]User, error) {
	const op = "repositories.users.postgresRepository.GetUsers"
	log.Printf("%s: start", op)
	argsList := strings.Join(common.Map(ids, func(id UserId) string {
		return fmt.Sprintf("'%s'", id)
	}), ",")
	query := fmt.Sprintf(`SELECT id, displayName, avatarId FROM users WHERE id IN (%s);`, argsList)
	rows, err := c.db.Query(query)
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return []User{}, err
	}
	defer rows.Close()
	users := []User{}
	for rows.Next() {
		var id string
		var displayName string
		var avatarId *string
		if err := rows.Scan(&id, &displayName, &avatarId); err != nil {
			log.Printf("%s: failed to perform scan err: %v", op, err)
			return []User{}, err
		}
		users = append(users, User{
			Id:          UserId(id),
			DisplayName: displayName,
			AvatarId:    (*AvatarId)(avatarId),
		})
	}
	if err := rows.Err(); err != nil {
		log.Printf("%s: found rows err: %v", op, err)
		return []User{}, err
	}
	log.Printf("%s: success", op)
	return users, nil
}

func (c *postgresRepository) SearchUsers(query string) ([]User, error) {
	const op = "repositories.users.postgresRepository.SearchUsers"
	log.Printf("%s: start[q=%s]", op, query)
	sqlQuery := fmt.Sprintf(`
SELECT 
	id, 
	displayName,
	avatarId
FROM 
	users 
WHERE 
	displayName LIKE '%s%%' or displayName = $1 
ORDER BY
	id;
`, query)
	rows, err := c.db.Query(sqlQuery, query)
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return []User{}, err
	}
	defer rows.Close()
	users := []User{}
	for rows.Next() {
		var id string
		var displayName string
		var avatarId *string
		if err := rows.Scan(&id, &displayName, &avatarId); err != nil {
			log.Printf("%s: failed to perform scan err: %v", op, err)
			return []User{}, err
		}
		users = append(users, User{
			Id:          UserId(id),
			DisplayName: displayName,
			AvatarId:    (*AvatarId)(avatarId),
		})
	}
	if err := rows.Err(); err != nil {
		log.Printf("%s: found rows err: %v", op, err)
		return []User{}, err
	}
	log.Printf("%s: success[q=%s]", op, query)
	return users, nil
}

func (c *postgresRepository) UpdateDisplayName(name string, id UserId) repositories.MutationWorkItem {
	const op = "repositories.users.postgresRepository.UpdateDisplayName"
	log.Printf("%s: start[name=%s id=%s]", op, name, id)
	users, err := c.GetUsers([]UserId{id})
	return repositories.MutationWorkItem{
		Perform: func() error {
			if err != nil {
				log.Printf("%s: cannot get user info: %v", op, err)
				return err
			}
			if len(users) == 0 {
				err := errors.New("no such user exists")
				log.Printf("%s: cannot get user info: %v", op, err)
				return err
			}
			return c.updateDisplayName(name, id)
		},
		Rollback: func() error {
			if err != nil {
				log.Printf("%s: cannot get user info: %v", op, err)
				return err
			}
			if len(users) == 0 {
				err := errors.New("no such user exists")
				log.Printf("%s: cannot get user info: %v", op, err)
				return err
			}
			return c.updateDisplayName(users[0].DisplayName, id)
		},
	}
}

func (c *postgresRepository) updateDisplayName(name string, id UserId) error {
	const op = "repositories.users.postgresRepository.updateDisplayName"
	log.Printf("%s: start[name=%s id=%s]", op, name, id)
	query := `UPDATE users SET displayName = $2 WHERE id = $1;`
	_, err := c.db.Exec(query, string(id), name)
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return err
	}
	log.Printf("%s: success[name=%s id=%s]", op, name, id)
	return nil
}

func (c *postgresRepository) UpdateAvatarId(avatarId *AvatarId, id UserId) repositories.MutationWorkItem {
	const op = "repositories.users.postgresRepository.UpdateAvatarId"
	log.Printf("%s: start[avatarId=%v id=%s]", op, avatarId, id)
	users, err := c.GetUsers([]UserId{id})
	return repositories.MutationWorkItem{
		Perform: func() error {
			if err != nil {
				log.Printf("%s: cannot get user info: %v", op, err)
				return err
			}
			if len(users) == 0 {
				err := errors.New("no such user exists")
				log.Printf("%s: cannot get user info: %v", op, err)
				return err
			}
			return c.updateAvatarId(avatarId, id)
		},
		Rollback: func() error {
			if err != nil {
				log.Printf("%s: cannot get user info: %v", op, err)
				return err
			}
			if len(users) == 0 {
				err := errors.New("no such user exists")
				log.Printf("%s: cannot get user info: %v", op, err)
				return err
			}
			return c.updateAvatarId((*AvatarId)(users[0].AvatarId), id)
		},
	}
}

func (c *postgresRepository) updateAvatarId(avatarId *AvatarId, id UserId) error {
	const op = "repositories.users.postgresRepository.updateAvatarId"
	log.Printf("%s: start[avatarId=%v id=%s]", op, avatarId, id)
	if avatarId == nil {
		query := `UPDATE users SET avatarId = NULL WHERE id = $1;`
		_, err := c.db.Exec(query, string(id), string(*avatarId))
		if err != nil {
			log.Printf("%s: failed to perform query err: %v", op, err)
			return err
		}
	} else {
		query := `UPDATE users SET avatarId = $2 WHERE id = $1;`
		_, err := c.db.Exec(query, string(id), string(*avatarId))
		if err != nil {
			log.Printf("%s: failed to perform query err: %v", op, err)
			return err
		}
	}
	log.Printf("%s: start[success=%v id=%s]", op, avatarId, id)
	return nil
}
