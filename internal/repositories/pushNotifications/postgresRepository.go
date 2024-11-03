package pushNotifications

import (
	"verni/internal/db"
	"verni/internal/repositories"
	"verni/internal/services/logging"
)

type postgresRepository struct {
	db     db.DB
	logger logging.Service
}

func (c *postgresRepository) StorePushToken(uid UserId, token string) repositories.MutationWorkItem {
	const op = "repositories.pushNotifications.postgresRepository.StorePushToken"
	currentToken, err := c.GetPushToken(uid)
	return repositories.MutationWorkItem{
		Perform: func() error {
			if err != nil {
				c.logger.Log("%s: failed to get current token info err: %v", op, err)
				return err
			}
			return c.storePushToken(uid, token)
		},
		Rollback: func() error {
			if err != nil {
				c.logger.Log("%s: failed to get current token info err: %v", op, err)
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

func (c *postgresRepository) storePushToken(uid UserId, token string) error {
	const op = "repositories.pushNotifications.postgresRepository.storePushToken"
	c.logger.Log("%s: start[uid=%v]", op, uid)
	query := `
INSERT INTO pushTokens(id, token) VALUES ($1, $2) 
ON CONFLICT (id) DO UPDATE SET token = $2;
`
	_, err := c.db.Exec(query, string(uid), token)
	if err != nil {
		c.logger.Log("%s: failed to perform query err: %v", op, err)
		return err
	}
	c.logger.Log("%s: success[uid=%v]", op, uid)
	return nil
}

func (c *postgresRepository) removePushToken(uid UserId) error {
	const op = "repositories.pushNotifications.postgresRepository.removePushToken"
	c.logger.Log("%s: start[uid=%v]", op, uid)
	query := `DELETE FROM pushTokens WHERE id = $1;`
	_, err := c.db.Exec(query, string(uid))
	if err != nil {
		c.logger.Log("%s: failed to perform query err: %v", op, err)
		return err
	}
	c.logger.Log("%s: success[uid=%v]", op, uid)
	return nil
}

func (c *postgresRepository) GetPushToken(uid UserId) (*string, error) {
	const op = "repositories.pushNotifications.postgresRepository.GetPushToken"
	c.logger.Log("%s: start[uid=%v]", op, uid)
	query := `SELECT token FROM pushTokens WHERE id = $1;`
	rows, err := c.db.Query(query, string(uid))
	if err != nil {
		c.logger.Log("%s: failed to perform query err: %v", op, err)
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			c.logger.Log("%s: failed to perform scan err: %v", op, err)
			return nil, err
		}
		if err := rows.Err(); err != nil {
			c.logger.Log("%s: found rows err: %v", op, err)
			return nil, err
		}
		c.logger.Log("%s: success[uid=%v]", op, uid)
		return &token, nil
	}
	if err := rows.Err(); err != nil {
		c.logger.Log("%s: found rows err: %v", op, err)
		return nil, err
	}
	c.logger.Log("%s: success[uid=%v]", op, uid)
	return nil, nil
}
