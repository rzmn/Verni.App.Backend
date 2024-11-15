package defaultRepository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"verni/internal/common"
	"verni/internal/db"
	"verni/internal/repositories"
	"verni/internal/repositories/users"
	"verni/internal/services/logging"
)

func New(db db.DB, logger logging.Service) users.Repository {
	return &defaultRepository{
		db:     db,
		logger: logger,
	}
}

type defaultRepository struct {
	db     db.DB
	logger logging.Service
}

func (c *defaultRepository) StoreUser(user users.User) repositories.MutationWorkItem {
	return repositories.MutationWorkItem{
		Perform: func() error {
			return c.storeUser(user)
		},
		Rollback: func() error {
			return c.removeUser(user.Id)
		},
	}
}

func (c *defaultRepository) storeUser(user users.User) error {
	const op = "repositories.users.postgresRepository.storeUser"
	c.logger.LogInfo("%s: start[id=%s]", op, user.Id)

	if user.AvatarId == nil {
		query := `INSERT INTO users(id, displayName, avatarId) VALUES ($1, $2, NULL);`
		_, err := c.db.Exec(query, user.Id, user.DisplayName)
		if err != nil {
			c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
			return err
		}
	} else {
		query := `INSERT INTO users(id, displayName, avatarId) VALUES ($1, $2, $3);`
		_, err := c.db.Exec(query, user.Id, user.DisplayName, *user.AvatarId)
		if err != nil {
			c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
			return err
		}
	}
	c.logger.LogInfo("%s: success[id=%s]", op, user.Id)
	return nil
}

func (c *defaultRepository) removeUser(userId users.UserId) error {
	const op = "repositories.users.postgresRepository.removeUser"
	c.logger.LogInfo("%s: start[id=%s]", op, userId)
	query := `DELETE FROM users WHERE id = $1;`
	_, err := c.db.Exec(query, string(userId))
	if err != nil {
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return err
	}
	c.logger.LogInfo("%s: success[id=%s]", op, userId)
	return nil
}

func (c *defaultRepository) GetUsers(ids []users.UserId) ([]users.User, error) {
	const op = "repositories.users.postgresRepository.GetUsers"
	c.logger.LogInfo("%s: start", op)
	if len(ids) == 0 {
		c.logger.LogInfo("%s: success", op)
		return []users.User{}, nil
	}
	query := fmt.Sprintf(
		`SELECT id, displayName, avatarId FROM users WHERE id IN (%s);`,
		strings.Join(common.Map(ids, func(id users.UserId) string {
			return fmt.Sprintf("'%s'", id)
		}), ","),
	)
	rows, err := c.db.Query(query)
	if err != nil {
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return []users.User{}, err
	}
	defer rows.Close()
	result := []users.User{}
	for rows.Next() {
		var id string
		var displayName string
		var sqlAvatarId sql.NullString
		if err := rows.Scan(&id, &displayName, &sqlAvatarId); err != nil {
			c.logger.LogInfo("%s: failed to perform scan err: %v", op, err)
			return []users.User{}, err
		}
		var avatarId *users.AvatarId
		if sqlAvatarId.Valid {
			avatarId = (*users.AvatarId)(&sqlAvatarId.String)
		} else {
			avatarId = nil
		}
		result = append(result, users.User{
			Id:          users.UserId(id),
			DisplayName: displayName,
			AvatarId:    avatarId,
		})
	}
	if err := rows.Err(); err != nil {
		c.logger.LogInfo("%s: found rows err: %v", op, err)
		return []users.User{}, err
	}
	c.logger.LogInfo("%s: success", op)
	return result, nil
}

func (c *defaultRepository) UpdateDisplayName(name string, id users.UserId) repositories.MutationWorkItem {
	const op = "repositories.users.postgresRepository.UpdateDisplayName"
	c.logger.LogInfo("%s: start[name=%s id=%s]", op, name, id)
	usersFromDb, err := c.GetUsers([]users.UserId{id})
	return repositories.MutationWorkItem{
		Perform: func() error {
			if err != nil {
				c.logger.LogInfo("%s: cannot get user info: %v", op, err)
				return err
			}
			if len(usersFromDb) == 0 {
				err := errors.New("no such user exists")
				c.logger.LogInfo("%s: cannot get user info: %v", op, err)
				return err
			}
			return c.updateDisplayName(name, id)
		},
		Rollback: func() error {
			if err != nil {
				c.logger.LogInfo("%s: cannot get user info: %v", op, err)
				return err
			}
			if len(usersFromDb) == 0 {
				err := errors.New("no such user exists")
				c.logger.LogInfo("%s: cannot get user info: %v", op, err)
				return err
			}
			return c.updateDisplayName(usersFromDb[0].DisplayName, id)
		},
	}
}

func (c *defaultRepository) updateDisplayName(name string, id users.UserId) error {
	const op = "repositories.users.postgresRepository.updateDisplayName"
	c.logger.LogInfo("%s: start[name=%s id=%s]", op, name, id)
	query := `UPDATE users SET displayName = $2 WHERE id = $1;`
	_, err := c.db.Exec(query, string(id), name)
	if err != nil {
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return err
	}
	c.logger.LogInfo("%s: success[name=%s id=%s]", op, name, id)
	return nil
}

func (c *defaultRepository) UpdateAvatarId(avatarId *users.AvatarId, id users.UserId) repositories.MutationWorkItem {
	const op = "repositories.users.postgresRepository.UpdateAvatarId"
	c.logger.LogInfo("%s: start[avatarId=%v id=%s]", op, avatarId, id)
	usersFromDb, err := c.GetUsers([]users.UserId{id})
	return repositories.MutationWorkItem{
		Perform: func() error {
			if err != nil {
				c.logger.LogInfo("%s: cannot get user info: %v", op, err)
				return err
			}
			if len(usersFromDb) == 0 {
				err := errors.New("no such user exists")
				c.logger.LogInfo("%s: cannot get user info: %v", op, err)
				return err
			}
			return c.updateAvatarId(avatarId, id)
		},
		Rollback: func() error {
			if err != nil {
				c.logger.LogInfo("%s: cannot get user info: %v", op, err)
				return err
			}
			if len(usersFromDb) == 0 {
				err := errors.New("no such user exists")
				c.logger.LogInfo("%s: cannot get user info: %v", op, err)
				return err
			}
			return c.updateAvatarId((*users.AvatarId)(usersFromDb[0].AvatarId), id)
		},
	}
}

func (c *defaultRepository) updateAvatarId(avatarId *users.AvatarId, id users.UserId) error {
	const op = "repositories.users.postgresRepository.updateAvatarId"
	c.logger.LogInfo("%s: start[avatarId=%v id=%s]", op, avatarId, id)
	if avatarId == nil {
		query := `UPDATE users SET avatarId = NULL WHERE id = $1;`
		_, err := c.db.Exec(query, string(id))
		if err != nil {
			c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
			return err
		}
	} else {
		query := `UPDATE users SET avatarId = $2 WHERE id = $1;`
		_, err := c.db.Exec(query, string(id), string(*avatarId))
		if err != nil {
			c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
			return err
		}
	}
	c.logger.LogInfo("%s: start[success=%v id=%s]", op, avatarId, id)
	return nil
}
