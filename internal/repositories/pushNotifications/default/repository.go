package defaultRepository

import (
	"github.com/rzmn/governi/internal/db"
	"github.com/rzmn/governi/internal/repositories"
	"github.com/rzmn/governi/internal/repositories/pushNotifications"
	"github.com/rzmn/governi/internal/services/logging"
)

func New(db db.DB, logger logging.Service) pushNotifications.Repository {
	return &defaultRepository{
		db:     db,
		logger: logger,
	}
}

type defaultRepository struct {
	db     db.DB
	logger logging.Service
}

func (c *defaultRepository) StorePushToken(uid pushNotifications.UserId, token string) repositories.MutationWorkItem {
	const op = "repositories.pushNotifications.postgresRepository.StorePushToken"
	currentToken, err := c.GetPushToken(uid)
	return repositories.MutationWorkItem{
		Perform: func() error {
			if err != nil {
				c.logger.LogInfo("%s: failed to get current token info err: %v", op, err)
				return err
			}
			return c.storePushToken(uid, token)
		},
		Rollback: func() error {
			if err != nil {
				c.logger.LogInfo("%s: failed to get current token info err: %v", op, err)
				return err
			}
			if currentToken == nil {
				return c.removePushToken(uid)
			} else {
				return c.storePushToken(uid, *currentToken)
			}
		},
	}
}

func (c *defaultRepository) storePushToken(uid pushNotifications.UserId, token string) error {
	const op = "repositories.pushNotifications.postgresRepository.storePushToken"
	c.logger.LogInfo("%s: start[uid=%v]", op, uid)
	query := `
INSERT INTO pushTokens(id, token) VALUES ($1, $2) 
ON CONFLICT (id) DO UPDATE SET token = $2;
`
	_, err := c.db.Exec(query, string(uid), token)
	if err != nil {
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return err
	}
	c.logger.LogInfo("%s: success[uid=%v]", op, uid)
	return nil
}

func (c *defaultRepository) removePushToken(uid pushNotifications.UserId) error {
	const op = "repositories.pushNotifications.postgresRepository.removePushToken"
	c.logger.LogInfo("%s: start[uid=%v]", op, uid)
	query := `DELETE FROM pushTokens WHERE id = $1;`
	_, err := c.db.Exec(query, string(uid))
	if err != nil {
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return err
	}
	c.logger.LogInfo("%s: success[uid=%v]", op, uid)
	return nil
}

func (c *defaultRepository) GetPushToken(uid pushNotifications.UserId) (*string, error) {
	const op = "repositories.pushNotifications.postgresRepository.GetPushToken"
	c.logger.LogInfo("%s: start[uid=%v]", op, uid)
	query := `SELECT token FROM pushTokens WHERE id = $1;`
	rows, err := c.db.Query(query, string(uid))
	if err != nil {
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			c.logger.LogInfo("%s: failed to perform scan err: %v", op, err)
			return nil, err
		}
		if err := rows.Err(); err != nil {
			c.logger.LogInfo("%s: found rows err: %v", op, err)
			return nil, err
		}
		c.logger.LogInfo("%s: success[uid=%v]", op, uid)
		return &token, nil
	}
	if err := rows.Err(); err != nil {
		c.logger.LogInfo("%s: found rows err: %v", op, err)
		return nil, err
	}
	c.logger.LogInfo("%s: success[uid=%v]", op, uid)
	return nil, nil
}
