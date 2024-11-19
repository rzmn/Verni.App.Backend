package defaultController

import (
	"fmt"

	"github.com/rzmn/governi/internal/common"
	"github.com/rzmn/governi/internal/controllers/users"
	friendsRepository "github.com/rzmn/governi/internal/repositories/friends"
	usersRepository "github.com/rzmn/governi/internal/repositories/users"
	"github.com/rzmn/governi/internal/services/logging"
)

type UsersRepository usersRepository.Repository
type FriendsRepository friendsRepository.Repository

func New(users UsersRepository, friends FriendsRepository, logger logging.Service) users.Controller {
	return &defaultController{
		users:   users,
		friends: friends,
		logger:  logger,
	}
}

type defaultController struct {
	users   UsersRepository
	friends FriendsRepository
	logger  logging.Service
}

func (c *defaultController) Get(ids []users.UserId, sender users.UserId) ([]users.User, *common.CodeBasedError[users.GetUsersErrorCode]) {
	const op = "users.defaultController.Get"
	c.logger.LogInfo("%s: start[sender=%s, ids=%v]", op, sender, ids)
	usersFromRepository, err := c.users.GetUsers(common.Map(ids, func(id users.UserId) usersRepository.UserId {
		return usersRepository.UserId(id)
	}))
	if err != nil {
		c.logger.LogInfo("%s: cannot get users from db err: %v", op, err)
		return []users.User{}, common.NewErrorWithDescription(users.GetUsersErrorInternal, err.Error())
	}
	userStatuses, err := c.friends.GetStatuses(friendsRepository.UserId(sender), common.Map(ids, func(id users.UserId) friendsRepository.UserId {
		return friendsRepository.UserId(id)
	}))
	if err != nil {
		c.logger.LogInfo("%s: cannot get friend statuses from db err: %v", op, err)
		return []users.User{}, common.NewErrorWithDescription(users.GetUsersErrorInternal, err.Error())
	}
	result := []users.User{}
	for _, user := range usersFromRepository {
		status, ok := userStatuses[friendsRepository.UserId(user.Id)]
		if !ok {
			c.logger.LogInfo("%s: cannot get friend status of user %s", op, user.Id)
			return []users.User{}, common.NewErrorWithDescription(users.GetUsersUserNotFound, fmt.Sprintf("user id: %s", user.Id))
		}
		result = append(result, users.User{
			Id:           users.UserId(user.Id),
			DisplayName:  user.DisplayName,
			AvatarId:     (*users.AvatarId)(user.AvatarId),
			FriendStatus: users.FriendStatus(status),
		})
	}
	c.logger.LogInfo("%s: success[sender=%s, ids=%v]", op, sender, ids)
	return result, nil
}
