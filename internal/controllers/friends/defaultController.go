package friends

import (
	"log"
	"slices"
	"verni/internal/common"
	"verni/internal/repositories/friends"
)

type defaultController struct {
	repository Repository
}

func (s *defaultController) AcceptFriendRequest(sender UserId, target UserId) *common.CodeBasedError[AcceptFriendRequestErrorCode] {
	const op = "friends.defaultController.AcceptFriendRequest"
	log.Printf("%s: start[sender=%s target=%s]", op, sender, target)
	hasRequest, err := s.repository.HasFriendRequest(friends.UserId(sender), friends.UserId(target))
	if err != nil {
		log.Printf("%s: cannot check friend request existence in db err: %v", op, err)
		return common.NewErrorWithDescription(AcceptFriendRequestErrorInternal, err.Error())
	}
	if !hasRequest {
		log.Printf("%s: does not have a friend request", op)
		return common.NewError(AcceptFriendRequestErrorNoSuchRequest)
	}
	transaction := s.repository.StoreFriendRequest(friends.UserId(target), friends.UserId(sender))
	if err := transaction.Perform(); err != nil {
		log.Printf("%s: cannot store friendship to db err: %v", op, err)
		return common.NewErrorWithDescription(AcceptFriendRequestErrorInternal, err.Error())
	}
	log.Printf("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}

func (s *defaultController) GetFriends(statuses []FriendStatus, userId UserId) (map[FriendStatus][]UserId, *common.CodeBasedError[GetFriendsErrorCode]) {
	const op = "friends.defaultController.GetFriends"
	log.Printf("%s: start[statuses=%v uid=%s]", op, statuses, userId)
	result := map[FriendStatus][]UserId{}
	for i := 0; i < len(statuses); i++ {
		result[statuses[i]] = []UserId{}
	}
	if slices.Contains(statuses, FriendStatusFriends) {
		ids, err := s.repository.GetFriends(friends.UserId(userId))
		if err != nil {
			log.Printf("%s: cannot get friends from db err: %v", op, err)
			return map[FriendStatus][]UserId{}, common.NewErrorWithDescription(GetFriendsErrorInternal, err.Error())
		}
		for i := range ids {
			result[FriendStatusFriends] = append(result[FriendStatusFriends], UserId(ids[i]))
		}
	}
	if slices.Contains(statuses, FriendStatusSubscriber) {
		ids, err := s.repository.GetSubscribers(friends.UserId(userId))
		if err != nil {
			log.Printf("%s: cannot get subscribers from db err: %v", op, err)
			return map[FriendStatus][]UserId{}, common.NewErrorWithDescription(GetFriendsErrorInternal, err.Error())
		}
		for i := range ids {
			result[FriendStatusSubscriber] = append(result[FriendStatusSubscriber], UserId(ids[i]))
		}
	}
	if slices.Contains(statuses, FriendStatusSubscription) {
		ids, err := s.repository.GetSubscriptions(friends.UserId(userId))
		if err != nil {
			log.Printf("%s: cannot get subscriptions from db err: %v", op, err)
			return map[FriendStatus][]UserId{}, common.NewErrorWithDescription(GetFriendsErrorInternal, err.Error())
		}
		for i := range ids {
			result[FriendStatusSubscription] = append(result[FriendStatusSubscription], UserId(ids[i]))
		}
	}
	log.Printf("%s: success[statuses=%v uid=%s]", op, statuses, userId)
	return result, nil
}

func (s *defaultController) RollbackFriendRequest(sender UserId, target UserId) *common.CodeBasedError[RollbackFriendRequestErrorCode] {
	const op = "friends.defaultController.RollbackFriendRequest"
	log.Printf("%s: start[sender=%s target=%s]", op, sender, target)
	hasRequest, err := s.repository.HasFriendRequest(friends.UserId(sender), friends.UserId(target))
	if err != nil {
		log.Printf("%s: cannot check has friend request in db err: %v", op, err)
		return common.NewErrorWithDescription(RollbackFriendRequestErrorInternal, err.Error())
	}
	if !hasRequest {
		log.Printf("%s: no friend request from %s to %s", op, sender, target)
		return common.NewError(RollbackFriendRequestErrorNoSuchRequest)
	}
	transaction := s.repository.RemoveFriendRequest(friends.UserId(sender), friends.UserId(target))
	if err := transaction.Perform(); err != nil {
		log.Printf("%s: cannot remove friend request from %s to %s from db err: %v", op, sender, target, err)
		return common.NewErrorWithDescription(RollbackFriendRequestErrorInternal, err.Error())
	}
	log.Printf("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}

func (s *defaultController) SendFriendRequest(sender UserId, target UserId) *common.CodeBasedError[SendFriendRequestErrorCode] {
	const op = "friends.defaultController.SendFriendRequest"
	log.Printf("%s: start[sender=%s target=%s]", op, sender, target)
	statuses, err := s.repository.GetStatuses(friends.UserId(sender), []friends.UserId{friends.UserId(target)})
	if err != nil {
		log.Printf("%s: cannot check target status of %s in db err: %v", op, target, err)
		return common.NewErrorWithDescription(SendFriendRequestErrorInternal, err.Error())
	}
	targetStatus, ok := statuses[friends.UserId(target)]
	if !ok {
		log.Printf("%s: no status found for %s in db", op, target)
		return common.NewError(SendFriendRequestErrorIncorrectUserStatus)
	}
	if targetStatus != friends.FriendStatusNo {
		switch targetStatus {
		case friends.FriendStatusSubscriber:
			log.Printf("%s: already have friend request from %s to %s", op, target, sender)
			return common.NewError(SendFriendRequestErrorHaveIncomingRequest)
		case friends.FriendStatusSubscription:
			log.Printf("%s: already have friend request from %s to %s", op, sender, target)
			return common.NewError(SendFriendRequestErrorAlreadySent)
		case friends.FriendStatusMe:
			log.Printf("%s: incorrect friend status %s: %d", op, target, targetStatus)
			return common.NewError(SendFriendRequestErrorIncorrectUserStatus)
		}
	}
	transaction := s.repository.StoreFriendRequest(friends.UserId(sender), friends.UserId(target))
	if err := transaction.Perform(); err != nil {
		log.Printf("%s: cannot store friend request from %s to %s in db err: %v", op, sender, target, err)
		return common.NewErrorWithDescription(SendFriendRequestErrorInternal, err.Error())
	}
	log.Printf("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}

func (s *defaultController) Unfriend(sender UserId, target UserId) *common.CodeBasedError[UnfriendErrorCode] {
	const op = "friends.defaultController.Unfriend"
	log.Printf("%s: start[sender=%s target=%s]", op, sender, target)
	statuses, err := s.repository.GetStatuses(friends.UserId(sender), []friends.UserId{friends.UserId(target)})
	if err != nil {
		log.Printf("%s: cannot check target status of %s in db err: %v", op, target, err)
		return common.NewErrorWithDescription(UnfriendErrorInternal, err.Error())
	}
	targetStatus, ok := statuses[friends.UserId(target)]
	if !ok {
		log.Printf("%s: no status found for %s in db", op, target)
		return common.NewError(UnfriendErrorNotAFriend)
	}
	if targetStatus != friends.FriendStatusFriend {
		log.Printf("%s: status of %s is %d, expected %d", op, target, targetStatus, friends.FriendStatusFriend)
		return common.NewError(UnfriendErrorNotAFriend)
	}
	transaction := s.repository.RemoveFriendRequest(friends.UserId(sender), friends.UserId(target))
	if err := transaction.Perform(); err != nil {
		log.Printf("%s: cannot remove friend request from %s to %s from db err: %v", op, sender, target, err)
		return common.NewErrorWithDescription(UnfriendErrorInternal, err.Error())
	}
	log.Printf("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}
