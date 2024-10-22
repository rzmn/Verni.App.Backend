package users

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"verni/internal/common"
)

type postgresRepository struct {
	db *sql.DB
}

func (c *postgresRepository) GetUsers(ids []UserId) ([]User, error) {
	const op = "repositories.users.postgresRepository.GetUsers"
	log.Printf("%s: start", op)
	argsList := strings.Join(common.Map(ids, func(id UserId) string {
		return fmt.Sprintf("'%s'", id)
	}), ",")
	query := fmt.Sprintf(`SELECT id, displayName FROM users WHERE u.id IN (%s);`, argsList)
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
		if err := rows.Scan(&id, &displayName); err != nil {
			log.Printf("%s: failed to perform scan err: %v", op, err)
			return []User{}, err
		}
		users = append(users, User{
			Id:          UserId(id),
			DisplayName: displayName,
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
	displayName 
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
		if err := rows.Scan(&id, &displayName); err != nil {
			log.Printf("%s: failed to perform scan err: %v", op, err)
			return []User{}, err
		}
		users = append(users, User{
			Id:          UserId(id),
			DisplayName: displayName,
		})
	}
	if err := rows.Err(); err != nil {
		log.Printf("%s: found rows err: %v", op, err)
		return []User{}, err
	}
	log.Printf("%s: success[q=%s]", op, query)
	return users, nil
}
