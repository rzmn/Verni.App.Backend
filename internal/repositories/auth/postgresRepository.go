package auth

import (
	"database/sql"
	"log"
	"verni/internal/repositories"

	"golang.org/x/crypto/bcrypt"
)

type postgresRepository struct {
	db *sql.DB
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (c *postgresRepository) CreateUser(uid UserId, email string, password string, refreshToken string) repositories.MutationWorkItem {
	return repositories.MutationWorkItem{
		Perform: func() error {
			return c.createUser(uid, email, password, refreshToken)
		},
		Rollback: func() error {
			return c.deleteUser(uid)
		},
	}
}

func (c *postgresRepository) CheckCredentials(email string, password string) (bool, error) {
	const op = "repositories.auth.postgresRepository.CheckCredentials"
	log.Printf("%s: start[email=%s]", op, email)
	query := `SELECT password FROM credentials WHERE email = $1;`
	row := c.db.QueryRow(query, email)
	var passwordHash string
	if err := row.Scan(&passwordHash); err != nil {
		log.Printf("%s: failed to perform scan err: %v", op, err)
		return false, err
	}
	log.Printf("%s: start[email=%s]", op, email)
	return checkPasswordHash(password, passwordHash), nil
}

func (c *postgresRepository) GetUserIdByEmail(email string) (*UserId, error) {
	const op = "repositories.auth.postgresRepository.GetUserIdByEmail"
	log.Printf("%s: start[email=%s]", op, email)
	query := `SELECT id FROM credentials WHERE email = $1;`
	rows, err := c.db.Query(query, email)
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Printf("%s: failed to perform scan err: %v", op, err)
			return nil, err
		}
		if err := rows.Err(); err != nil {
			log.Printf("%s: found rows err: %v", op, err)
			return nil, err
		}
		log.Printf("%s: success[email=%s]", op, email)
		return (*UserId)(&id), nil
	}
	if err := rows.Err(); err != nil {
		log.Printf("%s: found rows err: %v", op, err)
		return nil, err
	}
	log.Printf("%s: success[email=%s]", op, email)
	return nil, nil
}

func (c *postgresRepository) UpdateRefreshToken(uid UserId, token string) repositories.MutationWorkItem {
	const op = "repositories.auth.postgresRepository.UpdateRefreshToken"
	log.Printf("%s: start[uid=%s]", op, uid)
	existed, err := c.GetUserInfo(uid)
	return repositories.MutationWorkItem{
		Perform: func() error {
			if err != nil {
				log.Printf("%s: failed to get current credentals err: %v", op, err)
				return err
			}
			return c.updateRefreshToken(uid, token)
		},
		Rollback: func() error {
			if err != nil {
				log.Printf("%s: failed to get current credentals err: %v", op, err)
				return err
			}
			return c.updateRefreshToken(uid, existed.RefreshToken)
		},
	}
}

func (c *postgresRepository) updateRefreshToken(uid UserId, token string) error {
	const op = "repositories.auth.postgresRepository.updateRefreshToken"
	log.Printf("%s: start[uid=%s]", op, uid)
	query := `UPDATE credentials SET token = $2 WHERE id = $1;`
	_, err := c.db.Exec(query, string(uid), token)
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return err
	}
	log.Printf("%s: success[uid=%s]", op, uid)
	return nil
}

func (c *postgresRepository) UpdatePassword(uid UserId, password string) repositories.MutationWorkItem {
	const op = "repositories.auth.postgresRepository.UpdatePassword"
	log.Printf("%s: start[uid=%s]", op, uid)
	existed, getCredentialsErr := c.GetUserInfo(uid)
	passwordHash, hashPasswordErr := hashPassword(password)
	return repositories.MutationWorkItem{
		Perform: func() error {
			if getCredentialsErr != nil {
				log.Printf("%s: failed to get current credentals err: %v", op, getCredentialsErr)
				return getCredentialsErr
			}
			if hashPasswordErr != nil {
				log.Printf("%s: cannot hash password %v", op, hashPasswordErr)
				return hashPasswordErr
			}
			return c.updatePassword(uid, passwordHash)
		},
		Rollback: func() error {
			if getCredentialsErr != nil {
				log.Printf("%s: failed to get current credentals err: %v", op, getCredentialsErr)
				return getCredentialsErr
			}
			return c.updatePassword(uid, existed.PasswordHash)
		},
	}
}

func (c *postgresRepository) updatePassword(uid UserId, passwordHash string) error {
	const op = "repositories.auth.postgresRepository.updatePassword"
	log.Printf("%s: start[uid=%s]", op, uid)
	query := `UPDATE credentials SET password = $2 WHERE id = $1;`
	_, err := c.db.Exec(query, string(uid), passwordHash)
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return err
	}
	log.Printf("%s: success[uid=%s]", op, uid)
	return nil
}

func (c *postgresRepository) UpdateEmail(uid UserId, newEmail string) repositories.MutationWorkItem {
	const op = "repositories.auth.postgresRepository.UpdateEmail"
	log.Printf("%s: start[uid=%s]", op, uid)
	existed, err := c.GetUserInfo(uid)
	return repositories.MutationWorkItem{
		Perform: func() error {
			if err != nil {
				log.Printf("%s: failed to get current credentals err: %v", op, err)
				return err
			}
			return c.updateEmail(uid, newEmail)
		},
		Rollback: func() error {
			if err != nil {
				log.Printf("%s: failed to get current credentals err: %v", op, err)
				return err
			}
			return c.updateEmail(uid, existed.Email)
		},
	}
}

func (c *postgresRepository) updateEmail(uid UserId, newEmail string) error {
	const op = "repositories.auth.postgresRepository.updateEmail"
	log.Printf("%s: start[uid=%s]", op, uid)
	query := `UPDATE credentials SET email = $2 WHERE id = $1;`
	_, err := c.db.Exec(query, string(uid), newEmail)
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return err
	}
	log.Printf("%s: success[uid=%s]", op, uid)
	return nil
}

func (c *postgresRepository) GetRefreshToken(uid UserId) (string, error) {
	const op = "repositories.auth.postgresRepository.GetRefreshToken"
	log.Printf("%s: start[uid=%s]", op, uid)
	query := `SELECT token FROM credentials WHERE id = $1;`
	row := c.db.QueryRow(query, string(uid))
	var token string
	if err := row.Scan(&token); err != nil {
		log.Printf("%s: failed to perform scan err: %v", op, err)
		return "", err
	}
	log.Printf("%s: success[uid=%s]", op, uid)
	return token, nil
}

func (c *postgresRepository) createUser(uid UserId, email string, password string, refreshToken string) error {
	const op = "repositories.auth.postgresRepository.createUser"
	log.Printf("%s: start[uid=%s email=%s]", op, uid, email)
	passwordHash, err := hashPassword(password)
	if err != nil {
		log.Printf("%s: cannot hash password %v", op, err)
		return err
	}
	query := `INSERT INTO credentials(id, email, password, token, emailVerified) VALUES($1, $2, $3, $4, False);`
	_, err = c.db.Exec(query, string(uid), string(email), passwordHash, refreshToken)
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return err
	}
	log.Printf("%s: success[uid=%s email=%s]", op, uid, email)
	return nil
}

func (c *postgresRepository) deleteUser(uid UserId) error {
	const op = "repositories.auth.postgresRepository.deleteUser"
	log.Printf("%s: start[uid=%s]", op, uid)
	query := `DELETE FROM credentials WHERE id = $1;`
	_, err := c.db.Exec(query, string(uid))
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return err
	}
	log.Printf("%s: success[uid=%s]", op, uid)
	return nil
}

func (c *postgresRepository) GetUserInfo(uid UserId) (UserInfo, error) {
	const op = "repositories.auth.postgresRepository.GetCredentials"
	log.Printf("%s: start[uid=%s]", op, uid)
	query := `SELECT email, password, token, emailVerified FROM credentials WHERE id = $1;`
	row := c.db.QueryRow(query, string(uid))
	result := UserInfo{
		UserId: UserId(uid),
	}
	if err := row.Scan(&result.Email, &result.PasswordHash, &result.RefreshToken, &result.EmailVerified); err != nil {
		log.Printf("%s: failed to perform scan err: %v", op, err)
		return UserInfo{}, err
	}
	log.Printf("%s: success[uid=%s]", op, uid)
	return result, nil
}

func (c *postgresRepository) IsUserExists(uid UserId) (bool, error) {
	const op = "repositories.auth.postgresRepository.IsUserExists"
	log.Printf("%s: start[uid=%s]", op, uid)
	query := `SELECT EXISTS(SELECT 1 FROM credentials WHERE id = $1);`
	row := c.db.QueryRow(query, string(uid))
	var exists bool
	if err := row.Scan(&exists); err != nil {
		log.Printf("%s: failed to perform scan err: %v", op, err)
		return false, err
	}
	log.Printf("%s: success[uid=%s]", op, uid)
	return exists, nil
}
