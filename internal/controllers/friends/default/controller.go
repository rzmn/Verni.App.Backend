package defaultController

import (
	"slices"

	"github.com/rzmn/Verni.App.Backend/internal/common"
	"github.com/rzmn/Verni.App.Backend/internal/controllers/friends"
	friendsRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/friends"
	"github.com/rzmn/Verni.App.Backend/internal/services/logging"
)

type Repository friendsRepository.Repository

func New(repository Repository, logger logging.Service) friends.Controller {
	return &defaultController{
		repository: repository,
		logger:     logger,
	}
}

type defaultController struct {
	repository Repository
	logger     logging.Service
}

func (c *defaultController) AcceptFriendRequest(sender friends.UserId, target friends.UserId) *common.CodeBasedError[friends.AcceptFriendRequestErrorCode] {
	const op = "friends.defaultController.AcceptFriendRequest"
	c.logger.LogInfo("%s: start[sender=%s target=%s]", op, sender, target)
	hasRequest, err := c.repository.HasFriendRequest(friendsRepository.UserId(sender), friendsRepository.UserId(target))
	if err != nil {
		c.logger.LogInfo("%s: cannot check friend request existence in db err: %v", op, err)
		return common.NewErrorWithDescription(friends.AcceptFriendRequestErrorInternal, err.Error())
	}
	if !hasRequest {
		c.logger.LogInfo("%s: does not have a friend request", op)
		return common.NewError(friends.AcceptFriendRequestErrorNoSuchRequest)
	}
	transaction := c.repository.StoreFriendRequest(friendsRepository.UserId(target), friendsRepository.UserId(sender))
	if err := transaction.Perform(); err != nil {
		c.logger.LogInfo("%s: cannot store friendship to db err: %v", op, err)
		return common.NewErrorWithDescription(friends.AcceptFriendRequestErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}

func (c *defaultController) GetFriends(statuses []friends.FriendStatus, userId friends.UserId) (map[friends.FriendStatus][]friends.UserId, *common.CodeBasedError[friends.GetFriendsErrorCode]) {
	const op = "friends.defaultController.GetFriends"
	c.logger.LogInfo("%s: start[statuses=%v uid=%s]", op, statuses, userId)
	result := map[friends.FriendStatus][]friends.UserId{}
	for i := 0; i < len(statuses); i++ {
		result[statuses[i]] = []friends.UserId{}
	}
	if slices.Contains(statuses, friends.FriendStatusFriends) {
		ids, err := c.repository.GetFriends(friendsRepository.UserId(userId))
		if err != nil {
			c.logger.LogInfo("%s: cannot get friends from db err: %v", op, err)
			return map[friends.FriendStatus][]friends.UserId{}, common.NewErrorWithDescription(friends.GetFriendsErrorInternal, err.Error())
		}
		for i := range ids {
			result[friends.FriendStatusFriends] = append(result[friends.FriendStatusFriends], friends.UserId(ids[i]))
		}
	}
	if slices.Contains(statuses, friends.FriendStatusSubscriber) {
		ids, err := c.repository.GetSubscribers(friendsRepository.UserId(userId))
		if err != nil {
			c.logger.LogInfo("%s: cannot get subscribers from db err: %v", op, err)
			return map[friends.FriendStatus][]friends.UserId{}, common.NewErrorWithDescription(friends.GetFriendsErrorInternal, err.Error())
		}
		for i := range ids {
			result[friends.FriendStatusSubscriber] = append(result[friends.FriendStatusSubscriber], friends.UserId(ids[i]))
		}
	}
	if slices.Contains(statuses, friends.FriendStatusSubscription) {
		ids, err := c.repository.GetSubscriptions(friendsRepository.UserId(userId))
		if err != nil {
			c.logger.LogInfo("%s: cannot get subscriptions from db err: %v", op, err)
			return map[friends.FriendStatus][]friends.UserId{}, common.NewErrorWithDescription(friends.GetFriendsErrorInternal, err.Error())
		}
		for i := range ids {
			result[friends.FriendStatusSubscription] = append(result[friends.FriendStatusSubscription], friends.UserId(ids[i]))
		}
	}
	c.logger.LogInfo("%s: success[statuses=%v uid=%s]", op, statuses, userId)
	return result, nil
}

func (c *defaultController) RollbackFriendRequest(sender friends.UserId, target friends.UserId) *common.CodeBasedError[friends.RollbackFriendRequestErrorCode] {
	const op = "friends.defaultController.RollbackFriendRequest"
	c.logger.LogInfo("%s: start[sender=%s target=%s]", op, sender, target)
	hasRequest, err := c.repository.HasFriendRequest(friendsRepository.UserId(sender), friendsRepository.UserId(target))
	if err != nil {
		c.logger.LogInfo("%s: cannot check has friend request in db err: %v", op, err)
		return common.NewErrorWithDescription(friends.RollbackFriendRequestErrorInternal, err.Error())
	}
	if !hasRequest {
		c.logger.LogInfo("%s: no friend request from %s to %s", op, sender, target)
		return common.NewError(friends.RollbackFriendRequestErrorNoSuchRequest)
	}
	transaction := c.repository.RemoveFriendRequest(friendsRepository.UserId(sender), friendsRepository.UserId(target))
	if err := transaction.Perform(); err != nil {
		c.logger.LogInfo("%s: cannot remove friend request from %s to %s from db err: %v", op, sender, target, err)
		return common.NewErrorWithDescription(friends.RollbackFriendRequestErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}

func (c *defaultController) SendFriendRequest(sender friends.UserId, target friends.UserId) *common.CodeBasedError[friends.SendFriendRequestErrorCode] {
	const op = "friends.defaultController.SendFriendRequest"
	c.logger.LogInfo("%s: start[sender=%s target=%s]", op, sender, target)
	statuses, err := c.repository.GetStatuses(friendsRepository.UserId(sender), []friendsRepository.UserId{friendsRepository.UserId(target)})
	if err != nil {
		c.logger.LogInfo("%s: cannot check target status of %s in db err: %v", op, target, err)
		return common.NewErrorWithDescription(friends.SendFriendRequestErrorInternal, err.Error())
	}
	targetStatus, ok := statuses[friendsRepository.UserId(target)]
	if !ok {
		c.logger.LogInfo("%s: no status found for %s in db", op, target)
		return common.NewError(friends.SendFriendRequestErrorIncorrectUserStatus)
	}
	if targetStatus != friendsRepository.FriendStatusNo {
		switch targetStatus {
		case friendsRepository.FriendStatusSubscriber:
			c.logger.LogInfo("%s: already have friend request from %s to %s", op, target, sender)
			return common.NewError(friends.SendFriendRequestErrorHaveIncomingRequest)
		case friendsRepository.FriendStatusSubscription:
			c.logger.LogInfo("%s: already have friend request from %s to %s", op, sender, target)
			return common.NewError(friends.SendFriendRequestErrorAlreadySent)
		case friendsRepository.FriendStatusMe:
			c.logger.LogInfo("%s: incorrect friend status %s: %d", op, target, targetStatus)
			return common.NewError(friends.SendFriendRequestErrorIncorrectUserStatus)
		}
	}
	transaction := c.repository.StoreFriendRequest(friendsRepository.UserId(sender), friendsRepository.UserId(target))
	if err := transaction.Perform(); err != nil {
		c.logger.LogInfo("%s: cannot store friend request from %s to %s in db err: %v", op, sender, target, err)
		return common.NewErrorWithDescription(friends.SendFriendRequestErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}

func (c *defaultController) Unfriend(sender friends.UserId, target friends.UserId) *common.CodeBasedError[friends.UnfriendErrorCode] {
	const op = "friends.defaultController.Unfriend"
	c.logger.LogInfo("%s: start[sender=%s target=%s]", op, sender, target)
	statuses, err := c.repository.GetStatuses(friendsRepository.UserId(sender), []friendsRepository.UserId{friendsRepository.UserId(target)})
	if err != nil {
		c.logger.LogInfo("%s: cannot check target status of %s in db err: %v", op, target, err)
		return common.NewErrorWithDescription(friends.UnfriendErrorInternal, err.Error())
	}
	targetStatus, ok := statuses[friendsRepository.UserId(target)]
	if !ok {
		c.logger.LogInfo("%s: no status found for %s in db", op, target)
		return common.NewError(friends.UnfriendErrorNotAFriend)
	}
	if targetStatus != friendsRepository.FriendStatusFriend {
		c.logger.LogInfo("%s: status of %s is %d, expected %d", op, target, targetStatus, friendsRepository.FriendStatusFriend)
		return common.NewError(friends.UnfriendErrorNotAFriend)
	}
	transaction := c.repository.RemoveFriendRequest(friendsRepository.UserId(sender), friendsRepository.UserId(target))
	if err := transaction.Perform(); err != nil {
		c.logger.LogInfo("%s: cannot remove friend request from %s to %s from db err: %v", op, sender, target, err)
		return common.NewErrorWithDescription(friends.UnfriendErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}
