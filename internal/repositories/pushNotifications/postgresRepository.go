package pushNotifications

import (
	"database/sql"
	"log"
	"verni/internal/repositories"
)

type postgresRepository struct {
	db *sql.DB
}

func (c *postgresRepository) StorePushToken(uid UserId, token string) repositories.MutationWorkItem {
	const op = "repositories.pushNotifications.postgresRepository.StorePushToken"
	currentToken, err := c.GetPushToken(uid)
	return repositories.MutationWorkItem{
		Perform: func() error {
			if err != nil {
				log.Printf("%s: failed to get current token info err: %v", op, err)
				return err
			}
			return c.storePushToken(uid, token)
		},
		Rollback: func() error {
			if err != nil {
				log.Printf("%s: failed to get current token info err: %v", op, err)
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
	log.Printf("%s: start[uid=%v]", op, uid)

	query := `
INSERT INTO pushTokens(id, token) VALUES ($1, $2) 
ON CONFLICT (id) DO UPDATE SET token = $2;
`
	_, err := c.db.Exec(query, string(uid), token)
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return err
	}
	log.Printf("%s: success[uid=%v]", op, uid)
	return nil
}

func (c *postgresRepository) removePushToken(uid UserId) error {
	const op = "repositories.pushNotifications.postgresRepository.removePushToken"
	log.Printf("%s: start[uid=%v]", op, uid)
	query := `DELETE FROM pushTokens WHERE id = $1;`
	_, err := c.db.Exec(query, string(uid))
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return err
	}
	log.Printf("%s: success[uid=%v]", op, uid)
	return nil
}

func (c *postgresRepository) GetPushToken(uid UserId) (*string, error) {
	const op = "repositories.pushNotifications.postgresRepository.GetPushToken"
	log.Printf("%s: start[uid=%v]", op, uid)
	query := `SELECT token FROM pushTokens WHERE id = $1;`
	rows, err := c.db.Query(query, string(uid))
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			log.Printf("%s: failed to perform scan err: %v", op, err)
			return nil, err
		}
		if err := rows.Err(); err != nil {
			log.Printf("%s: found rows err: %v", op, err)
			return nil, err
		}
		log.Printf("%s: success[uid=%v]", op, uid)
		return &token, nil
	}
	if err := rows.Err(); err != nil {
		log.Printf("%s: found rows err: %v", op, err)
		return nil, err
	}
	log.Printf("%s: success[uid=%v]", op, uid)
	return nil, nil
}
