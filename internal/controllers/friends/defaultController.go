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

func (s *defaultController) AcceptFriendRequest(sender UserId, target UserId) *common.CodeBasedError[AcceptFriendRequestErrorCode] {
	const op = "friends.defaultController.AcceptFriendRequest"
	s.logger.Log("%s: start[sender=%s target=%s]", op, sender, target)
	hasRequest, err := s.repository.HasFriendRequest(friends.UserId(sender), friends.UserId(target))
	if err != nil {
		s.logger.Log("%s: cannot check friend request existence in db err: %v", op, err)
		return common.NewErrorWithDescription(AcceptFriendRequestErrorInternal, err.Error())
	}
	if !hasRequest {
		s.logger.Log("%s: does not have a friend request", op)
		return common.NewError(AcceptFriendRequestErrorNoSuchRequest)
	}
	transaction := s.repository.StoreFriendRequest(friends.UserId(target), friends.UserId(sender))
	if err := transaction.Perform(); err != nil {
		s.logger.Log("%s: cannot store friendship to db err: %v", op, err)
		return common.NewErrorWithDescription(AcceptFriendRequestErrorInternal, err.Error())
	}
	s.logger.Log("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}

func (s *defaultController) GetFriends(statuses []FriendStatus, userId UserId) (map[FriendStatus][]UserId, *common.CodeBasedError[GetFriendsErrorCode]) {
	const op = "friends.defaultController.GetFriends"
	s.logger.Log("%s: start[statuses=%v uid=%s]", op, statuses, userId)
	result := map[FriendStatus][]UserId{}
	for i := 0; i < len(statuses); i++ {
		result[statuses[i]] = []UserId{}
	}
	if slices.Contains(statuses, FriendStatusFriends) {
		ids, err := s.repository.GetFriends(friends.UserId(userId))
		if err != nil {
			s.logger.Log("%s: cannot get friends from db err: %v", op, err)
			return map[FriendStatus][]UserId{}, common.NewErrorWithDescription(GetFriendsErrorInternal, err.Error())
		}
		for i := range ids {
			result[FriendStatusFriends] = append(result[FriendStatusFriends], UserId(ids[i]))
		}
	}
	if slices.Contains(statuses, FriendStatusSubscriber) {
		ids, err := s.repository.GetSubscribers(friends.UserId(userId))
		if err != nil {
			s.logger.Log("%s: cannot get subscribers from db err: %v", op, err)
			return map[FriendStatus][]UserId{}, common.NewErrorWithDescription(GetFriendsErrorInternal, err.Error())
		}
		for i := range ids {
			result[FriendStatusSubscriber] = append(result[FriendStatusSubscriber], UserId(ids[i]))
		}
	}
	if slices.Contains(statuses, FriendStatusSubscription) {
		ids, err := s.repository.GetSubscriptions(friends.UserId(userId))
		if err != nil {
			s.logger.Log("%s: cannot get subscriptions from db err: %v", op, err)
			return map[FriendStatus][]UserId{}, common.NewErrorWithDescription(GetFriendsErrorInternal, err.Error())
		}
		for i := range ids {
			result[FriendStatusSubscription] = append(result[FriendStatusSubscription], UserId(ids[i]))
		}
	}
	s.logger.Log("%s: success[statuses=%v uid=%s]", op, statuses, userId)
	return result, nil
}

func (s *defaultController) RollbackFriendRequest(sender UserId, target UserId) *common.CodeBasedError[RollbackFriendRequestErrorCode] {
	const op = "friends.defaultController.RollbackFriendRequest"
	s.logger.Log("%s: start[sender=%s target=%s]", op, sender, target)
	hasRequest, err := s.repository.HasFriendRequest(friends.UserId(sender), friends.UserId(target))
	if err != nil {
		s.logger.Log("%s: cannot check has friend request in db err: %v", op, err)
		return common.NewErrorWithDescription(RollbackFriendRequestErrorInternal, err.Error())
	}
	if !hasRequest {
		s.logger.Log("%s: no friend request from %s to %s", op, sender, target)
		return common.NewError(RollbackFriendRequestErrorNoSuchRequest)
	}
	transaction := s.repository.RemoveFriendRequest(friends.UserId(sender), friends.UserId(target))
	if err := transaction.Perform(); err != nil {
		s.logger.Log("%s: cannot remove friend request from %s to %s from db err: %v", op, sender, target, err)
		return common.NewErrorWithDescription(RollbackFriendRequestErrorInternal, err.Error())
	}
	s.logger.Log("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}

func (s *defaultController) SendFriendRequest(sender UserId, target UserId) *common.CodeBasedError[SendFriendRequestErrorCode] {
	const op = "friends.defaultController.SendFriendRequest"
	s.logger.Log("%s: start[sender=%s target=%s]", op, sender, target)
	statuses, err := s.repository.GetStatuses(friends.UserId(sender), []friends.UserId{friends.UserId(target)})
	if err != nil {
		s.logger.Log("%s: cannot check target status of %s in db err: %v", op, target, err)
		return common.NewErrorWithDescription(SendFriendRequestErrorInternal, err.Error())
	}
	targetStatus, ok := statuses[friends.UserId(target)]
	if !ok {
		s.logger.Log("%s: no status found for %s in db", op, target)
		return common.NewError(SendFriendRequestErrorIncorrectUserStatus)
	}
	if targetStatus != friends.FriendStatusNo {
		switch targetStatus {
		case friends.FriendStatusSubscriber:
			s.logger.Log("%s: already have friend request from %s to %s", op, target, sender)
			return common.NewError(SendFriendRequestErrorHaveIncomingRequest)
		case friends.FriendStatusSubscription:
			s.logger.Log("%s: already have friend request from %s to %s", op, sender, target)
			return common.NewError(SendFriendRequestErrorAlreadySent)
		case friends.FriendStatusMe:
			s.logger.Log("%s: incorrect friend status %s: %d", op, target, targetStatus)
			return common.NewError(SendFriendRequestErrorIncorrectUserStatus)
		}
	}
	transaction := s.repository.StoreFriendRequest(friends.UserId(sender), friends.UserId(target))
	if err := transaction.Perform(); err != nil {
		s.logger.Log("%s: cannot store friend request from %s to %s in db err: %v", op, sender, target, err)
		return common.NewErrorWithDescription(SendFriendRequestErrorInternal, err.Error())
	}
	s.logger.Log("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}

func (s *defaultController) Unfriend(sender UserId, target UserId) *common.CodeBasedError[UnfriendErrorCode] {
	const op = "friends.defaultController.Unfriend"
	s.logger.Log("%s: start[sender=%s target=%s]", op, sender, target)
	statuses, err := s.repository.GetStatuses(friends.UserId(sender), []friends.UserId{friends.UserId(target)})
	if err != nil {
		s.logger.Log("%s: cannot check target status of %s in db err: %v", op, target, err)
		return common.NewErrorWithDescription(UnfriendErrorInternal, err.Error())
	}
	targetStatus, ok := statuses[friends.UserId(target)]
	if !ok {
		s.logger.Log("%s: no status found for %s in db", op, target)
		return common.NewError(UnfriendErrorNotAFriend)
	}
	if targetStatus != friends.FriendStatusFriend {
		s.logger.Log("%s: status of %s is %d, expected %d", op, target, targetStatus, friends.FriendStatusFriend)
		return common.NewError(UnfriendErrorNotAFriend)
	}
	transaction := s.repository.RemoveFriendRequest(friends.UserId(sender), friends.UserId(target))
	if err := transaction.Perform(); err != nil {
		s.logger.Log("%s: cannot remove friend request from %s to %s from db err: %v", op, sender, target, err)
		return common.NewErrorWithDescription(UnfriendErrorInternal, err.Error())
	}
	s.logger.Log("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}
