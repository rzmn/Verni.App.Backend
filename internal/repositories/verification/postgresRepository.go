package verification

import (
	"log"
	"verni/internal/db"
	"verni/internal/repositories"
)

type postgresRepository struct {
	db db.DB
}

func (c *postgresRepository) StoreEmailVerificationCode(email string, code string) repositories.MutationWorkItem {
	const op = "repositories.verification.postgresRepository.StoreEmailVerificationCode"
	currentCode, err := c.GetEmailVerificationCode(email)
	return repositories.MutationWorkItem{
		Perform: func() error {
			if err != nil {
				log.Printf("%s: failed to get current code err: %v", op, err)
				return err
			}
			return c.storeEmailVerificationCode(email, code)
		},
		Rollback: func() error {
			if err != nil {
				log.Printf("%s: failed to get current code err: %v", op, err)
				return err
			}
			if currentCode == nil {
				return c.removeEmailVerificationCode(email)
			} else {
				return c.storeEmailVerificationCode(email, *currentCode)
			}
		},
	}
}

func (c *postgresRepository) storeEmailVerificationCode(email string, code string) error {
	const op = "repositories.verification.postgresRepository.storeEmailVerificationCode"
	log.Printf("%s: start[email=%s]", op, email)
	query := `
INSERT INTO emailVerification(email, code) VALUES ($1, $2) 
ON CONFLICT (email) DO UPDATE SET code = $2;
`
	_, err := c.db.Exec(query, email, code)
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return err
	}
	log.Printf("%s: success[email=%s]", op, email)
	return nil
}

func (c *postgresRepository) GetEmailVerificationCode(email string) (*string, error) {
	const op = "repositories.verification.postgresRepository.GetEmailVerificationCode"
	log.Printf("%s: start[email=%s]", op, email)
	query := `SELECT code FROM emailVerification WHERE email = $1;`
	rows, err := c.db.Query(query, email)
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			log.Printf("%s: failed to perform scan err: %v", op, err)
			return nil, err
		}
		if err := rows.Err(); err != nil {
			log.Printf("%s: found rows err: %v", op, err)
			return nil, err
		}
		log.Printf("%s: success[email=%s]", op, email)
		return &code, nil
	}
	if err := rows.Err(); err != nil {
		log.Printf("%s: found rows err: %v", op, err)
		return nil, err
	}
	log.Printf("%s: success[email=%s]", op, email)
	return nil, nil
}

func (c *postgresRepository) RemoveEmailVerificationCode(email string) repositories.MutationWorkItem {
	const op = "repositories.verification.postgresRepository.RemoveEmailVerificationCode"
	code, err := c.GetEmailVerificationCode(email)
	return repositories.MutationWorkItem{
		Perform: func() error {
			if err != nil {
				log.Printf("%s: failed to get current code err: %v", op, err)
				return err
			}
			if code == nil {
				return nil
			} else {
				return c.removeEmailVerificationCode(email)
			}
		},
		Rollback: func() error {
			if err != nil {
				log.Printf("%s: failed to get current code err: %v", op, err)
				return err
			}
			if code == nil {
				return nil
			} else {
				return c.storeEmailVerificationCode(email, *code)
			}
		},
	}
}

func (c *postgresRepository) removeEmailVerificationCode(email string) error {
	const op = "repositories.verification.postgresRepository.removeEmailVerificationCode"
	log.Printf("%s: start[email=%s]", op, email)
	query := `DELETE FROM emailVerification WHERE email = $1;`
	_, err := c.db.Exec(query, email)
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return err
	}
	log.Printf("%s: success[email=%s]", op, email)
	return nil
}
