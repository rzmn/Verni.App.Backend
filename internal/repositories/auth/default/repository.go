package defaultRepository

import (
	"github.com/rzmn/governi/internal/db"
	"github.com/rzmn/governi/internal/repositories"
	"github.com/rzmn/governi/internal/repositories/auth"
	"github.com/rzmn/governi/internal/services/logging"

	"golang.org/x/crypto/bcrypt"
)

func New(db db.DB, logger logging.Service) auth.Repository {
	return &defaultRepository{
		db:     db,
		logger: logger,
	}
}

type defaultRepository struct {
	db     db.DB
	logger logging.Service
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (c *defaultRepository) CreateUser(uid auth.UserId, email string, password string, refreshToken string) repositories.MutationWorkItem {
	return repositories.MutationWorkItem{
		Perform: func() error {
			return c.createUser(uid, email, password, refreshToken)
		},
		Rollback: func() error {
			return c.deleteUser(uid)
		},
	}
}

func (c *defaultRepository) CheckCredentials(email string, password string) (bool, error) {
	const op = "repositories.auth.postgresRepository.CheckCredentials"
	c.logger.LogInfo("%s: start[email=%s]", op, email)
	query := `SELECT password FROM credentials WHERE email = $1;`
	row := c.db.QueryRow(query, email)
	var passwordHash string
	if err := row.Scan(&passwordHash); err != nil {
		c.logger.LogInfo("%s: failed to perform scan err: %v", op, err)
		return false, err
	}
	c.logger.LogInfo("%s: start[email=%s]", op, email)
	return checkPasswordHash(password, passwordHash), nil
}

func (c *defaultRepository) GetUserIdByEmail(email string) (*auth.UserId, error) {
	const op = "repositories.auth.postgresRepository.GetUserIdByEmail"
	c.logger.LogInfo("%s: start[email=%s]", op, email)
	query := `SELECT id FROM credentials WHERE email = $1;`
	rows, err := c.db.Query(query, email)
	if err != nil {
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			c.logger.LogInfo("%s: failed to perform scan err: %v", op, err)
			return nil, err
		}
		if err := rows.Err(); err != nil {
			c.logger.LogInfo("%s: found rows err: %v", op, err)
			return nil, err
		}
		c.logger.LogInfo("%s: success[email=%s]", op, email)
		return (*auth.UserId)(&id), nil
	}
	if err := rows.Err(); err != nil {
		c.logger.LogInfo("%s: found rows err: %v", op, err)
		return nil, err
	}
	c.logger.LogInfo("%s: success[email=%s]", op, email)
	return nil, nil
}

func (c *defaultRepository) UpdateRefreshToken(uid auth.UserId, token string) repositories.MutationWorkItem {
	const op = "repositories.auth.postgresRepository.UpdateRefreshToken"
	c.logger.LogInfo("%s: start[uid=%s]", op, uid)
	existed, err := c.GetUserInfo(uid)
	return repositories.MutationWorkItem{
		Perform: func() error {
			if err != nil {
				c.logger.LogInfo("%s: failed to get current credentals err: %v", op, err)
				return err
			}
			return c.updateRefreshToken(uid, token)
		},
		Rollback: func() error {
			if err != nil {
				c.logger.LogInfo("%s: failed to get current credentals err: %v", op, err)
				return err
			}
			return c.updateRefreshToken(uid, existed.RefreshToken)
		},
	}
}

func (c *defaultRepository) updateRefreshToken(uid auth.UserId, token string) error {
	const op = "repositories.auth.postgresRepository.updateRefreshToken"
	c.logger.LogInfo("%s: start[uid=%s]", op, uid)
	query := `UPDATE credentials SET token = $2 WHERE id = $1;`
	_, err := c.db.Exec(query, string(uid), token)
	if err != nil {
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return err
	}
	c.logger.LogInfo("%s: success[uid=%s]", op, uid)
	return nil
}

func (c *defaultRepository) UpdatePassword(uid auth.UserId, password string) repositories.MutationWorkItem {
	const op = "repositories.auth.postgresRepository.UpdatePassword"
	c.logger.LogInfo("%s: start[uid=%s]", op, uid)
	existed, getCredentialsErr := c.GetUserInfo(uid)
	passwordHash, hashPasswordErr := hashPassword(password)
	return repositories.MutationWorkItem{
		Perform: func() error {
			if getCredentialsErr != nil {
				c.logger.LogInfo("%s: failed to get current credentals err: %v", op, getCredentialsErr)
				return getCredentialsErr
			}
			if hashPasswordErr != nil {
				c.logger.LogInfo("%s: cannot hash password %v", op, hashPasswordErr)
				return hashPasswordErr
			}
			return c.updatePassword(uid, passwordHash)
		},
		Rollback: func() error {
			if getCredentialsErr != nil {
				c.logger.LogInfo("%s: failed to get current credentals err: %v", op, getCredentialsErr)
				return getCredentialsErr
			}
			return c.updatePassword(uid, existed.PasswordHash)
		},
	}
}

func (c *defaultRepository) updatePassword(uid auth.UserId, passwordHash string) error {
	const op = "repositories.auth.postgresRepository.updatePassword"
	c.logger.LogInfo("%s: start[uid=%s]", op, uid)
	query := `UPDATE credentials SET password = $2 WHERE id = $1;`
	_, err := c.db.Exec(query, string(uid), passwordHash)
	if err != nil {
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return err
	}
	c.logger.LogInfo("%s: success[uid=%s]", op, uid)
	return nil
}

func (c *defaultRepository) UpdateEmail(uid auth.UserId, newEmail string) repositories.MutationWorkItem {
	const op = "repositories.auth.postgresRepository.UpdateEmail"
	c.logger.LogInfo("%s: start[uid=%s]", op, uid)
	existed, err := c.GetUserInfo(uid)
	return repositories.MutationWorkItem{
		Perform: func() error {
			if err != nil {
				c.logger.LogInfo("%s: failed to get current credentals err: %v", op, err)
				return err
			}
			return c.updateEmail(uid, newEmail)
		},
		Rollback: func() error {
			if err != nil {
				c.logger.LogInfo("%s: failed to get current credentals err: %v", op, err)
				return err
			}
			if err := c.updateEmail(uid, existed.Email); err != nil {
				return err
			}
			if existed.EmailVerified {
				if err := c.markUserEmailValidated(uid, true); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func (c *defaultRepository) updateEmail(uid auth.UserId, newEmail string) error {
	const op = "repositories.auth.postgresRepository.updateEmail"
	c.logger.LogInfo("%s: start[uid=%s]", op, uid)
	query := `UPDATE credentials SET email = $2, emailVerified = False WHERE id = $1;`
	_, err := c.db.Exec(query, string(uid), newEmail)
	if err != nil {
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return err
	}
	c.logger.LogInfo("%s: success[uid=%s]", op, uid)
	return nil
}

func (c *defaultRepository) GetRefreshToken(uid auth.UserId) (string, error) {
	const op = "repositories.auth.postgresRepository.GetRefreshToken"
	c.logger.LogInfo("%s: start[uid=%s]", op, uid)
	query := `SELECT token FROM credentials WHERE id = $1;`
	row := c.db.QueryRow(query, string(uid))
	var token string
	if err := row.Scan(&token); err != nil {
		c.logger.LogInfo("%s: failed to perform scan err: %v", op, err)
		return "", err
	}
	c.logger.LogInfo("%s: success[uid=%s]", op, uid)
	return token, nil
}

func (c *defaultRepository) MarkUserEmailValidated(uid auth.UserId) repositories.MutationWorkItem {
	const op = "repositories.auth.postgresRepository.MarkUserEmailValidated"
	existed, err := c.GetUserInfo(uid)
	return repositories.MutationWorkItem{
		Perform: func() error {
			if err != nil {
				c.logger.LogInfo("%s: failed to get current credentals err: %v", op, err)
				return err
			}
			if existed.EmailVerified {
				return nil
			} else {
				return c.markUserEmailValidated(uid, true)
			}
		},
		Rollback: func() error {
			if err != nil {
				c.logger.LogInfo("%s: failed to get current credentals err: %v", op, err)
				return err
			}
			if existed.EmailVerified {
				return nil
			} else {
				return c.markUserEmailValidated(uid, false)
			}
		},
	}
}

func (c *defaultRepository) markUserEmailValidated(uid auth.UserId, validated bool) error {
	const op = "repositories.auth.postgresRepository.markUserEmailValidated"
	c.logger.LogInfo("%s: start[uid=%s validated=%t]", op, uid, validated)
	query := `UPDATE credentials SET emailVerified = $2 WHERE id = $1;`
	_, err := c.db.Exec(query, string(uid), validated)
	if err != nil {
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return err
	}
	c.logger.LogInfo("%s: success[uid=%s validated=%t]", op, uid, validated)
	return nil
}

func (c *defaultRepository) createUser(uid auth.UserId, email string, password string, refreshToken string) error {
	const op = "repositories.auth.postgresRepository.createUser"
	c.logger.LogInfo("%s: start[uid=%s email=%s]", op, uid, email)
	passwordHash, err := hashPassword(password)
	if err != nil {
		c.logger.LogInfo("%s: cannot hash password %v", op, err)
		return err
	}
	query := `INSERT INTO credentials(id, email, password, token, emailVerified) VALUES($1, $2, $3, $4, False);`
	_, err = c.db.Exec(query, string(uid), string(email), passwordHash, refreshToken)
	if err != nil {
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return err
	}
	c.logger.LogInfo("%s: success[uid=%s email=%s]", op, uid, email)
	return nil
}

func (c *defaultRepository) deleteUser(uid auth.UserId) error {
	const op = "repositories.auth.postgresRepository.deleteUser"
	c.logger.LogInfo("%s: start[uid=%s]", op, uid)
	query := `DELETE FROM credentials WHERE id = $1;`
	_, err := c.db.Exec(query, string(uid))
	if err != nil {
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return err
	}
	c.logger.LogInfo("%s: success[uid=%s]", op, uid)
	return nil
}

func (c *defaultRepository) GetUserInfo(uid auth.UserId) (auth.UserInfo, error) {
	const op = "repositories.auth.postgresRepository.GetCredentials"
	c.logger.LogInfo("%s: start[uid=%s]", op, uid)
	query := `SELECT email, password, token, emailVerified FROM credentials WHERE id = $1;`
	row := c.db.QueryRow(query, string(uid))
	result := auth.UserInfo{
		UserId: auth.UserId(uid),
	}
	if err := row.Scan(&result.Email, &result.PasswordHash, &result.RefreshToken, &result.EmailVerified); err != nil {
		c.logger.LogInfo("%s: failed to perform scan err: %v", op, err)
		return auth.UserInfo{}, err
	}
	c.logger.LogInfo("%s: success[uid=%s]", op, uid)
	return result, nil
}

func (c *defaultRepository) IsUserExists(uid auth.UserId) (bool, error) {
	const op = "repositories.auth.postgresRepository.IsUserExists"
	c.logger.LogInfo("%s: start[uid=%s]", op, uid)
	query := `SELECT EXISTS(SELECT 1 FROM credentials WHERE id = $1);`
	row := c.db.QueryRow(query, string(uid))
	var exists bool
	if err := row.Scan(&exists); err != nil {
		c.logger.LogInfo("%s: failed to perform scan err: %v", op, err)
		return false, err
	}
	c.logger.LogInfo("%s: success[uid=%s]", op, uid)
	return exists, nil
}
