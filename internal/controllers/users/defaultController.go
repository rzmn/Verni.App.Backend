package users

import (
	"fmt"
	"verni/internal/common"
	"verni/internal/repositories/friends"
	"verni/internal/repositories/users"
	"verni/internal/services/logging"
)

type defaultController struct {
	users   UsersRepository
	friends FriendsRepository
	logger  logging.Service
}

func (c *defaultController) Get(ids []UserId, sender UserId) ([]User, *common.CodeBasedError[GetUsersErrorCode]) {
	const op = "users.defaultController.Get"
	c.logger.LogInfo("%s: start[sender=%s, ids=%v]", op, sender, ids)
	usersFromRepository, err := c.users.GetUsers(common.Map(ids, func(id UserId) users.UserId {
		return users.UserId(id)
	}))
	if err != nil {
		c.logger.LogInfo("%s: cannot get users from db err: %v", op, err)
		return []User{}, common.NewErrorWithDescription(GetUsersErrorInternal, err.Error())
	}
	userStatuses, err := c.friends.GetStatuses(friends.UserId(sender), common.Map(ids, func(id UserId) friends.UserId {
		return friends.UserId(id)
	}))
	if err != nil {
		c.logger.LogInfo("%s: cannot get friend statuses from db err: %v", op, err)
		return []User{}, common.NewErrorWithDescription(GetUsersErrorInternal, err.Error())
	}
	result := []User{}
	for _, user := range usersFromRepository {
		status, ok := userStatuses[friends.UserId(user.Id)]
		if !ok {
			c.logger.LogInfo("%s: cannot get friend status of user %s", op, user.Id)
			return []User{}, common.NewErrorWithDescription(GetUsersUserNotFound, fmt.Sprintf("user id: %s", user.Id))
		}
		result = append(result, User{
			Id:           UserId(user.Id),
			DisplayName:  user.DisplayName,
			AvatarId:     (*AvatarId)(user.AvatarId),
			FriendStatus: FriendStatus(status),
		})
	}
	c.logger.LogInfo("%s: success[sender=%s, ids=%v]", op, sender, ids)
	return result, nil
}
