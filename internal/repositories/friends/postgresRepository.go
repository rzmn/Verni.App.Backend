package friends

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"verni/internal/common"
	"verni/internal/repositories"
)

type postgresRepository struct {
	db *sql.DB
}

func (c *postgresRepository) GetFriends(userId UserId) ([]UserId, error) {
	const op = "repositories.friends.postgresRepository.GetFriends"
	log.Printf("%s: start[userId=%s]", op, userId)
	query := `
SELECT r1.sender FROM friendRequests r1 WHERE r1.target = $1 AND EXISTS (
	SELECT * FROM friendRequests r2 WHERE r2.sender = $1 AND r2.target = r1.sender
);
`
	rows, err := c.db.Query(query, string(userId))
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return []UserId{}, err
	}
	defer rows.Close()
	subscriptions := []UserId{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Printf("%s: failed to perform scan err: %v", op, err)
			return []UserId{}, err
		}
		subscriptions = append(subscriptions, UserId(id))
	}
	if err := rows.Err(); err != nil {
		log.Printf("%s: found rows err: %v", op, err)
		return []UserId{}, err
	}
	log.Printf("%s: success[userId=%s]", op, userId)
	return subscriptions, nil
}

func (c *postgresRepository) GetSubscribers(userId UserId) ([]UserId, error) {
	const op = "repositories.friends.postgresRepository.GetSubscribers"
	log.Printf("%s: start[userId=%s]", op, userId)
	query := `
SELECT r1.sender FROM friendRequests r1 WHERE r1.target = $1 AND NOT EXISTS (
	SELECT * FROM friendRequests r2 WHERE r2.sender = $1 AND r2.target = r1.sender
);
`
	rows, err := c.db.Query(query, string(userId))
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return []UserId{}, err
	}
	defer rows.Close()
	subscriptions := []UserId{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Printf("%s: failed to perform scan err: %v", op, err)
			return []UserId{}, err
		}
		subscriptions = append(subscriptions, UserId(id))
	}
	if err := rows.Err(); err != nil {
		log.Printf("%s: found rows err: %v", op, err)
		return []UserId{}, err
	}
	log.Printf("%s: success[userId=%s]", op, userId)
	return subscriptions, nil
}

func (c *postgresRepository) GetSubscriptions(userId UserId) ([]UserId, error) {
	const op = "repositories.friends.postgresRepository.GetSubscriptions"
	log.Printf("%s: start[userId=%s]", op, userId)
	query := `
SELECT r1.target FROM friendRequests r1 WHERE r1.sender = $1 AND NOT EXISTS (
	SELECT * FROM friendRequests r2 WHERE r2.target = $1 AND r2.sender = r1.target
);
`
	rows, err := c.db.Query(query, string(userId))
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return []UserId{}, err
	}
	defer rows.Close()
	subscriptions := []UserId{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Printf("%s: failed to perform scan err: %v", op, err)
			return []UserId{}, err
		}
		subscriptions = append(subscriptions, UserId(id))
	}
	if err := rows.Err(); err != nil {
		log.Printf("%s: found rows err: %v", op, err)
		return []UserId{}, err
	}
	log.Printf("%s: success[userId=%s]", op, userId)
	return subscriptions, nil
}

func (c *postgresRepository) GetStatuses(sender UserId, ids []UserId) (map[UserId]FriendStatus, error) {
	const op = "repositories.friends.postgresRepository.GetStatuses"
	log.Printf("%s: start[sender=%s]", op, sender)

	if len(ids) == 0 {
		log.Printf("%s: success[sender=%s]", op, sender)
		return map[UserId]FriendStatus{}, nil
	}
	argsList := strings.Join(common.Map(ids, func(id UserId) string {
		return fmt.Sprintf("'%s'", id)
	}), ",")
	query := fmt.Sprintf(`
SELECT 
    id,
    CASE 
        WHEN EXISTS (SELECT 1 FROM friendRequests WHERE sender = $1 AND target = id) 
        THEN TRUE 
        ELSE FALSE 
    END AS isSubscriber,
    CASE 
        WHEN EXISTS (SELECT 1 FROM friendRequests WHERE target = $1 AND sender = id) 
        THEN TRUE 
        ELSE FALSE 
    END AS isSubscription,
    CASE 
        WHEN EXISTS (
        	SELECT 1 FROM friendRequests r1 WHERE r1.target = $1 AND r1.sender = id AND EXISTS (
				SELECT * FROM friendRequests r2 WHERE r2.sender = $1 AND r2.target = id
			)
        ) 
        THEN TRUE 
        ELSE FALSE 
    END AS isFriend
FROM 
    unnest( ARRAY[(%s)] ) AS id;
`, argsList)
	rows, err := c.db.Query(query, string(sender))
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return map[UserId]FriendStatus{}, err
	}
	defer rows.Close()
	statuses := map[UserId]FriendStatus{}
	for rows.Next() {
		var id string
		var isSubscriber bool
		var isSubscription bool
		var isFriend bool
		if err := rows.Scan(&id, &isSubscriber, &isSubscription, &isFriend); err != nil {
			log.Printf("%s: failed to perform scan err: %v", op, err)
			return map[UserId]FriendStatus{}, err
		}
		status := FriendStatusNo
		if isFriend {
			status = FriendStatusFriend
		} else if isSubscriber {
			status = FriendStatusSubscriber
		} else if isSubscription {
			status = FriendStatusSubscription
		} else if id == string(sender) {
			status = FriendStatusMe
		}
		statuses[UserId(id)] = FriendStatus(status)
	}
	if err := rows.Err(); err != nil {
		log.Printf("%s: found rows err: %v", op, err)
		return map[UserId]FriendStatus{}, err
	}
	log.Printf("%s: success[sender=%s]", op, sender)
	return statuses, nil
}

func (c *postgresRepository) HasFriendRequest(sender UserId, target UserId) (bool, error) {
	const op = "repositories.friends.postgresRepository.HasFriendRequest"
	log.Printf("%s: start[sender=%s target=%s]", op, sender, target)
	query := `SELECT EXISTS(SELECT 1 FROM friendRequests WHERE sender = $1 AND target = $2);`
	row := c.db.QueryRow(query, string(sender), string(target))
	var has bool
	if err := row.Scan(&has); err != nil {
		log.Printf("%s: failed to perform scan err: %v", op, err)
		return false, err
	}
	log.Printf("%s: success[sender=%s target=%s]", op, sender, target)
	return has, nil
}

func (c *postgresRepository) StoreFriendRequest(sender UserId, target UserId) repositories.MutationWorkItem {
	return repositories.MutationWorkItem{
		Perform: func() error {
			return c.storeFriendRequest(sender, target)
		},
		Rollback: func() error {
			return c.removeFriendRequest(sender, target)
		},
	}
}

func (c *postgresRepository) storeFriendRequest(sender UserId, target UserId) error {
	const op = "repositories.friends.postgresRepository.storeFriendRequest"
	log.Printf("%s: start[sender=%s target=%s]", op, sender, target)
	query := `INSERT INTO friendRequests(sender, target) VALUES($1, $2);`
	_, err := c.db.Exec(query, string(sender), string(target))
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return err
	}
	log.Printf("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}

func (c *postgresRepository) RemoveFriendRequest(sender UserId, target UserId) repositories.MutationWorkItem {
	return repositories.MutationWorkItem{
		Perform: func() error {
			return c.removeFriendRequest(sender, target)
		},
		Rollback: func() error {
			return c.storeFriendRequest(sender, target)
		},
	}
}

func (c *postgresRepository) removeFriendRequest(sender UserId, target UserId) error {
	const op = "repositories.friends.postgresRepository.removeFriendRequest"
	log.Printf("%s: start[sender=%s target=%s]", op, sender, target)
	query := `DELETE FROM friendRequests WHERE sender = $1 and target = $2;`
	_, err := c.db.Exec(query, string(sender), string(target))
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return err
	}
	log.Printf("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}
