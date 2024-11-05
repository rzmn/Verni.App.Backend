package friends

import (
	"slices"
	"verni/internal/common"
	"verni/internal/repositories/friends"
	"verni/internal/services/logging"
)

type defaultController struct {
	repository Repository
	logger     logging.Service
}

func (c *defaultController) AcceptFriendRequest(sender UserId, target UserId) *common.CodeBasedError[AcceptFriendRequestErrorCode] {
	const op = "friends.defaultController.AcceptFriendRequest"
	c.logger.LogInfo("%s: start[sender=%s target=%s]", op, sender, target)
	hasRequest, err := c.repository.HasFriendRequest(friends.UserId(sender), friends.UserId(target))
	if err != nil {
		c.logger.LogInfo("%s: cannot check friend request existence in db err: %v", op, err)
		return common.NewErrorWithDescription(AcceptFriendRequestErrorInternal, err.Error())
	}
	if !hasRequest {
		c.logger.LogInfo("%s: does not have a friend request", op)
		return common.NewError(AcceptFriendRequestErrorNoSuchRequest)
	}
	transaction := c.repository.StoreFriendRequest(friends.UserId(target), friends.UserId(sender))
	if err := transaction.Perform(); err != nil {
		c.logger.LogInfo("%s: cannot store friendship to db err: %v", op, err)
		return common.NewErrorWithDescription(AcceptFriendRequestErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}

func (c *defaultController) GetFriends(statuses []FriendStatus, userId UserId) (map[FriendStatus][]UserId, *common.CodeBasedError[GetFriendsErrorCode]) {
	const op = "friends.defaultController.GetFriends"
	c.logger.LogInfo("%s: start[statuses=%v uid=%s]", op, statuses, userId)
	result := map[FriendStatus][]UserId{}
	for i := 0; i < len(statuses); i++ {
		result[statuses[i]] = []UserId{}
	}
	if slices.Contains(statuses, FriendStatusFriends) {
		ids, err := c.repository.GetFriends(friends.UserId(userId))
		if err != nil {
			c.logger.LogInfo("%s: cannot get friends from db err: %v", op, err)
			return map[FriendStatus][]UserId{}, common.NewErrorWithDescription(GetFriendsErrorInternal, err.Error())
		}
		for i := range ids {
			result[FriendStatusFriends] = append(result[FriendStatusFriends], UserId(ids[i]))
		}
	}
	if slices.Contains(statuses, FriendStatusSubscriber) {
		ids, err := c.repository.GetSubscribers(friends.UserId(userId))
		if err != nil {
			c.logger.LogInfo("%s: cannot get subscribers from db err: %v", op, err)
			return map[FriendStatus][]UserId{}, common.NewErrorWithDescription(GetFriendsErrorInternal, err.Error())
		}
		for i := range ids {
			result[FriendStatusSubscriber] = append(result[FriendStatusSubscriber], UserId(ids[i]))
		}
	}
	if slices.Contains(statuses, FriendStatusSubscription) {
		ids, err := c.repository.GetSubscriptions(friends.UserId(userId))
		if err != nil {
			c.logger.LogInfo("%s: cannot get subscriptions from db err: %v", op, err)
			return map[FriendStatus][]UserId{}, common.NewErrorWithDescription(GetFriendsErrorInternal, err.Error())
		}
		for i := range ids {
			result[FriendStatusSubscription] = append(result[FriendStatusSubscription], UserId(ids[i]))
		}
	}
	c.logger.LogInfo("%s: success[statuses=%v uid=%s]", op, statuses, userId)
	return result, nil
}

func (c *defaultController) RollbackFriendRequest(sender UserId, target UserId) *common.CodeBasedError[RollbackFriendRequestErrorCode] {
	const op = "friends.defaultController.RollbackFriendRequest"
	c.logger.LogInfo("%s: start[sender=%s target=%s]", op, sender, target)
	hasRequest, err := c.repository.HasFriendRequest(friends.UserId(sender), friends.UserId(target))
	if err != nil {
		c.logger.LogInfo("%s: cannot check has friend request in db err: %v", op, err)
		return common.NewErrorWithDescription(RollbackFriendRequestErrorInternal, err.Error())
	}
	if !hasRequest {
		c.logger.LogInfo("%s: no friend request from %s to %s", op, sender, target)
		return common.NewError(RollbackFriendRequestErrorNoSuchRequest)
	}
	transaction := c.repository.RemoveFriendRequest(friends.UserId(sender), friends.UserId(target))
	if err := transaction.Perform(); err != nil {
		c.logger.LogInfo("%s: cannot remove friend request from %s to %s from db err: %v", op, sender, target, err)
		return common.NewErrorWithDescription(RollbackFriendRequestErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}

func (c *defaultController) SendFriendRequest(sender UserId, target UserId) *common.CodeBasedError[SendFriendRequestErrorCode] {
	const op = "friends.defaultController.SendFriendRequest"
	c.logger.LogInfo("%s: start[sender=%s target=%s]", op, sender, target)
	statuses, err := c.repository.GetStatuses(friends.UserId(sender), []friends.UserId{friends.UserId(target)})
	if err != nil {
		c.logger.LogInfo("%s: cannot check target status of %s in db err: %v", op, target, err)
		return common.NewErrorWithDescription(SendFriendRequestErrorInternal, err.Error())
	}
	targetStatus, ok := statuses[friends.UserId(target)]
	if !ok {
		c.logger.LogInfo("%s: no status found for %s in db", op, target)
		return common.NewError(SendFriendRequestErrorIncorrectUserStatus)
	}
	if targetStatus != friends.FriendStatusNo {
		switch targetStatus {
		case friends.FriendStatusSubscriber:
			c.logger.LogInfo("%s: already have friend request from %s to %s", op, target, sender)
			return common.NewError(SendFriendRequestErrorHaveIncomingRequest)
		case friends.FriendStatusSubscription:
			c.logger.LogInfo("%s: already have friend request from %s to %s", op, sender, target)
			return common.NewError(SendFriendRequestErrorAlreadySent)
		case friends.FriendStatusMe:
			c.logger.LogInfo("%s: incorrect friend status %s: %d", op, target, targetStatus)
			return common.NewError(SendFriendRequestErrorIncorrectUserStatus)
		}
	}
	transaction := c.repository.StoreFriendRequest(friends.UserId(sender), friends.UserId(target))
	if err := transaction.Perform(); err != nil {
		c.logger.LogInfo("%s: cannot store friend request from %s to %s in db err: %v", op, sender, target, err)
		return common.NewErrorWithDescription(SendFriendRequestErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}

func (c *defaultController) Unfriend(sender UserId, target UserId) *common.CodeBasedError[UnfriendErrorCode] {
	const op = "friends.defaultController.Unfriend"
	c.logger.LogInfo("%s: start[sender=%s target=%s]", op, sender, target)
	statuses, err := c.repository.GetStatuses(friends.UserId(sender), []friends.UserId{friends.UserId(target)})
	if err != nil {
		c.logger.LogInfo("%s: cannot check target status of %s in db err: %v", op, target, err)
		return common.NewErrorWithDescription(UnfriendErrorInternal, err.Error())
	}
	targetStatus, ok := statuses[friends.UserId(target)]
	if !ok {
		c.logger.LogInfo("%s: no status found for %s in db", op, target)
		return common.NewError(UnfriendErrorNotAFriend)
	}
	if targetStatus != friends.FriendStatusFriend {
		c.logger.LogInfo("%s: status of %s is %d, expected %d", op, target, targetStatus, friends.FriendStatusFriend)
		return common.NewError(UnfriendErrorNotAFriend)
	}
	transaction := c.repository.RemoveFriendRequest(friends.UserId(sender), friends.UserId(target))
	if err := transaction.Perform(); err != nil {
		c.logger.LogInfo("%s: cannot remove friend request from %s to %s from db err: %v", op, sender, target, err)
		return common.NewErrorWithDescription(UnfriendErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}
