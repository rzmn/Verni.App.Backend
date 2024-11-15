package defaultRepository

import (
	"fmt"
	"strings"
	"verni/internal/common"
	"verni/internal/db"
	"verni/internal/repositories"
	"verni/internal/repositories/friends"
	"verni/internal/services/logging"
)

func New(db db.DB, logger logging.Service) friends.Repository {
	return &defaultRepository{
		db:     db,
		logger: logger,
	}
}

type defaultRepository struct {
	db     db.DB
	logger logging.Service
}

func (c *defaultRepository) GetFriends(userId friends.UserId) ([]friends.UserId, error) {
	const op = "repositories.friends.postgresRepository.GetFriends"
	c.logger.LogInfo("%s: start[userId=%s]", op, userId)
	query := `
SELECT r1.sender FROM friendRequests r1 WHERE r1.target = $1 AND EXISTS (
	SELECT * FROM friendRequests r2 WHERE r2.sender = $1 AND r2.target = r1.sender
);
`
	rows, err := c.db.Query(query, string(userId))
	if err != nil {
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return []friends.UserId{}, err
	}
	defer rows.Close()
	subscriptions := []friends.UserId{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			c.logger.LogInfo("%s: failed to perform scan err: %v", op, err)
			return []friends.UserId{}, err
		}
		subscriptions = append(subscriptions, friends.UserId(id))
	}
	if err := rows.Err(); err != nil {
		c.logger.LogInfo("%s: found rows err: %v", op, err)
		return []friends.UserId{}, err
	}
	c.logger.LogInfo("%s: success[userId=%s]", op, userId)
	return subscriptions, nil
}

func (c *defaultRepository) GetSubscribers(userId friends.UserId) ([]friends.UserId, error) {
	const op = "repositories.friends.postgresRepository.GetSubscribers"
	c.logger.LogInfo("%s: start[userId=%s]", op, userId)
	query := `
SELECT r1.sender FROM friendRequests r1 WHERE r1.target = $1 AND NOT EXISTS (
	SELECT * FROM friendRequests r2 WHERE r2.sender = $1 AND r2.target = r1.sender
);
`
	rows, err := c.db.Query(query, string(userId))
	if err != nil {
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return []friends.UserId{}, err
	}
	defer rows.Close()
	subscriptions := []friends.UserId{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			c.logger.LogInfo("%s: failed to perform scan err: %v", op, err)
			return []friends.UserId{}, err
		}
		subscriptions = append(subscriptions, friends.UserId(id))
	}
	if err := rows.Err(); err != nil {
		c.logger.LogInfo("%s: found rows err: %v", op, err)
		return []friends.UserId{}, err
	}
	c.logger.LogInfo("%s: success[userId=%s]", op, userId)
	return subscriptions, nil
}

func (c *defaultRepository) GetSubscriptions(userId friends.UserId) ([]friends.UserId, error) {
	const op = "repositories.friends.postgresRepository.GetSubscriptions"
	c.logger.LogInfo("%s: start[userId=%s]", op, userId)
	query := `
SELECT r1.target FROM friendRequests r1 WHERE r1.sender = $1 AND NOT EXISTS (
	SELECT * FROM friendRequests r2 WHERE r2.target = $1 AND r2.sender = r1.target
);
`
	rows, err := c.db.Query(query, string(userId))
	if err != nil {
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return []friends.UserId{}, err
	}
	defer rows.Close()
	subscriptions := []friends.UserId{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			c.logger.LogInfo("%s: failed to perform scan err: %v", op, err)
			return []friends.UserId{}, err
		}
		subscriptions = append(subscriptions, friends.UserId(id))
	}
	if err := rows.Err(); err != nil {
		c.logger.LogInfo("%s: found rows err: %v", op, err)
		return []friends.UserId{}, err
	}
	c.logger.LogInfo("%s: success[userId=%s]", op, userId)
	return subscriptions, nil
}

func (c *defaultRepository) GetStatuses(sender friends.UserId, ids []friends.UserId) (map[friends.UserId]friends.FriendStatus, error) {
	const op = "repositories.friends.postgresRepository.GetStatuses"
	c.logger.LogInfo("%s: start[sender=%s]", op, sender)

	if len(ids) == 0 {
		c.logger.LogInfo("%s: success[sender=%s]", op, sender)
		return map[friends.UserId]friends.FriendStatus{}, nil
	}
	argsList := strings.Join(common.Map(ids, func(id friends.UserId) string {
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
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return map[friends.UserId]friends.FriendStatus{}, err
	}
	defer rows.Close()
	statuses := map[friends.UserId]friends.FriendStatus{}
	for rows.Next() {
		var id string
		var isSubscriber bool
		var isSubscription bool
		var isFriend bool
		if err := rows.Scan(&id, &isSubscriber, &isSubscription, &isFriend); err != nil {
			c.logger.LogInfo("%s: failed to perform scan err: %v", op, err)
			return map[friends.UserId]friends.FriendStatus{}, err
		}
		status := friends.FriendStatusNo
		if isFriend {
			status = friends.FriendStatusFriend
		} else if isSubscriber {
			status = friends.FriendStatusSubscriber
		} else if isSubscription {
			status = friends.FriendStatusSubscription
		} else if id == string(sender) {
			status = friends.FriendStatusMe
		}
		statuses[friends.UserId(id)] = friends.FriendStatus(status)
	}
	if err := rows.Err(); err != nil {
		c.logger.LogInfo("%s: found rows err: %v", op, err)
		return map[friends.UserId]friends.FriendStatus{}, err
	}
	c.logger.LogInfo("%s: success[sender=%s]", op, sender)
	return statuses, nil
}

func (c *defaultRepository) HasFriendRequest(sender friends.UserId, target friends.UserId) (bool, error) {
	const op = "repositories.friends.postgresRepository.HasFriendRequest"
	hasRequest, err := c.hasFriendRequest(sender, target)
	if err != nil {
		c.logger.LogInfo("%s: failed to call hasFriendRequest from sender to target err: %v", op, err)
		return false, err
	}
	if !hasRequest {
		return false, nil
	}
	hasRequestFromTarget, err := c.hasFriendRequest(target, sender)
	if err != nil {
		c.logger.LogInfo("%s: failed to call hasFriendRequest from target to sender err: %v", op, err)
		return false, err
	}
	return !hasRequestFromTarget, nil
}

func (c *defaultRepository) hasFriendRequest(sender friends.UserId, target friends.UserId) (bool, error) {
	const op = "repositories.friends.postgresRepository.hasFriendRequest"
	c.logger.LogInfo("%s: start[sender=%s target=%s]", op, sender, target)
	query := `SELECT EXISTS(SELECT 1 FROM friendRequests WHERE sender = $1 AND target = $2);`
	row := c.db.QueryRow(query, string(sender), string(target))
	var has bool
	if err := row.Scan(&has); err != nil {
		c.logger.LogInfo("%s: failed to perform scan err: %v", op, err)
		return false, err
	}
	c.logger.LogInfo("%s: success[sender=%s target=%s]", op, sender, target)
	return has, nil
}

func (c *defaultRepository) StoreFriendRequest(sender friends.UserId, target friends.UserId) repositories.MutationWorkItem {
	return repositories.MutationWorkItem{
		Perform: func() error {
			return c.storeFriendRequest(sender, target)
		},
		Rollback: func() error {
			return c.removeFriendRequest(sender, target)
		},
	}
}

func (c *defaultRepository) storeFriendRequest(sender friends.UserId, target friends.UserId) error {
	const op = "repositories.friends.postgresRepository.storeFriendRequest"
	c.logger.LogInfo("%s: start[sender=%s target=%s]", op, sender, target)
	query := `INSERT INTO friendRequests(sender, target) VALUES($1, $2);`
	_, err := c.db.Exec(query, string(sender), string(target))
	if err != nil {
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return err
	}
	c.logger.LogInfo("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}

func (c *defaultRepository) RemoveFriendRequest(sender friends.UserId, target friends.UserId) repositories.MutationWorkItem {
	const op = "repositories.friends.postgresRepository.RemoveFriendRequest"
	has, err := c.HasFriendRequest(sender, target)
	return repositories.MutationWorkItem{
		Perform: func() error {
			if err != nil {
				c.logger.LogInfo("%s: failed to check is friendship exists err: %v", op, err)
				return err
			}
			if has {
				return c.removeFriendRequest(sender, target)
			} else {
				return nil
			}
		},
		Rollback: func() error {
			if err != nil {
				c.logger.LogInfo("%s: failed to check is friendship exists err: %v", op, err)
				return err
			}
			if has {
				return c.storeFriendRequest(sender, target)
			} else {
				return nil
			}
		},
	}
}

func (c *defaultRepository) removeFriendRequest(sender friends.UserId, target friends.UserId) error {
	const op = "repositories.friends.postgresRepository.removeFriendRequest"
	c.logger.LogInfo("%s: start[sender=%s target=%s]", op, sender, target)
	query := `DELETE FROM friendRequests WHERE sender = $1 and target = $2;`
	_, err := c.db.Exec(query, string(sender), string(target))
	if err != nil {
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return err
	}
	c.logger.LogInfo("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}
