package friends

import (
	"accounty/internal/common"
	"accounty/internal/storage"
	"log"
	"slices"
)

type defaultController struct {
	storage storage.Storage
}

func (s *defaultController) AcceptFriendRequest(sender UserId, target UserId) *common.CodeBasedError[AcceptFriendRequestErrorCode] {
	const op = "friends.defaultController.AcceptFriendRequest"
	log.Printf("%s: start[sender=%s target=%s]", op, sender, target)
	hasRequest, err := s.storage.HasFriendRequest(storage.UserId(sender), storage.UserId(target))
	if err != nil {
		log.Printf("%s: cannot check friend request existence in db err: %v", op, err)
		return common.NewErrorWithDescription(AcceptFriendRequestErrorInternal, err.Error())
	}
	if !hasRequest {
		log.Printf("%s: does not have a friend request", op)
		return common.NewError(AcceptFriendRequestErrorNoSuchRequest)
	}
	if err := s.storage.RemoveFriendRequest(storage.UserId(sender), storage.UserId(target)); err != nil {
		log.Printf("%s: cannot remove friend request in db err: %v", op, err)
		return common.NewErrorWithDescription(AcceptFriendRequestErrorInternal, err.Error())
	}
	if err := s.storage.StoreFriendship(storage.UserId(sender), storage.UserId(target)); err != nil {
		log.Printf("%s: cannot store friendship to db err: %v", op, err)
		return common.NewErrorWithDescription(AcceptFriendRequestErrorInternal, err.Error())
	}
	log.Printf("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}

func (s *defaultController) GetFriends(statuses []FriendStatus, userId UserId) (map[FriendStatus][]UserId, *common.CodeBasedError[GetFriendsErrorCode]) {
	const op = "friends.defaultController.GetFriends"
	log.Printf("%s: start[statuses=%v uid=%s]", op, statuses, userId)
	friends := map[FriendStatus][]UserId{}
	if slices.Contains(statuses, FriendStatusFriends) {
		ids, err := s.storage.GetFriends(storage.UserId(userId))
		if err != nil {
			log.Printf("%s: cannot get friends from db err: %v", op, err)
			return map[FriendStatus][]UserId{}, common.NewErrorWithDescription(GetFriendsErrorInternal, err.Error())
		}
		for i := range ids {
			friends[FriendStatusFriends] = append(friends[FriendStatusFriends], UserId(ids[i]))
		}
	}
	if slices.Contains(statuses, FriendStatusSubscriber) {
		ids, err := s.storage.GetIncomingRequests(storage.UserId(userId))
		if err != nil {
			log.Printf("%s: cannot get subscribers from db err: %v", op, err)
			return map[FriendStatus][]UserId{}, common.NewErrorWithDescription(GetFriendsErrorInternal, err.Error())
		}
		for i := range ids {
			friends[FriendStatusSubscriber] = append(friends[FriendStatusSubscriber], UserId(ids[i]))
		}
	}
	if slices.Contains(statuses, FriendStatusSubscription) {
		ids, err := s.storage.GetPendingRequests(storage.UserId(userId))
		if err != nil {
			log.Printf("%s: cannot get subscriptions from db err: %v", op, err)
			return map[FriendStatus][]UserId{}, common.NewErrorWithDescription(GetFriendsErrorInternal, err.Error())
		}
		for i := range ids {
			friends[FriendStatusSubscription] = append(friends[FriendStatusSubscription], UserId(ids[i]))
		}
	}
	log.Printf("%s: success[statuses=%v uid=%s]", op, statuses, userId)
	return friends, nil
}

func (s *defaultController) RejectFriendRequest(sender UserId, target UserId) *common.CodeBasedError[RejectFriendRequestErrorCode] {
	const op = "friends.defaultController.RejectFriendRequest"
	log.Printf("%s: start[sender=%s target=%s]", op, sender, target)
	hasRequest, err := s.storage.HasFriendRequest(storage.UserId(sender), storage.UserId(target))
	if err != nil {
		log.Printf("%s: cannot check has friend request in db err: %v", op, err)
		return common.NewErrorWithDescription(RejectFriendRequestErrorInternal, err.Error())
	}
	if !hasRequest {
		log.Printf("%s: no friend request from %s to %s", op, sender, target)
		return common.NewError(RejectFriendRequestErrorNoSuchRequest)
	}
	if err := s.storage.RemoveFriendRequest(storage.UserId(sender), storage.UserId(target)); err != nil {
		log.Printf("%s: cannot remove friend request from %s to %s from db err: %v", op, sender, target, err)
		return common.NewErrorWithDescription(RejectFriendRequestErrorInternal, err.Error())
	}
	log.Printf("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}

func (s *defaultController) RollbackFriendRequest(sender UserId, target UserId) *common.CodeBasedError[RollbackFriendRequestErrorCode] {
	const op = "friends.defaultController.RollbackFriendRequest"
	log.Printf("%s: start[sender=%s target=%s]", op, sender, target)
	hasRequest, err := s.storage.HasFriendRequest(storage.UserId(sender), storage.UserId(target))
	if err != nil {
		log.Printf("%s: cannot check friendship in db err: %v", op, err)
		return common.NewErrorWithDescription(RollbackFriendRequestErrorInternal, err.Error())
	}
	if !hasRequest {
		log.Printf("%s: no friend request from %s to %s", op, sender, target)
		return common.NewError(RollbackFriendRequestErrorNoSuchRequest)
	}
	if err := s.storage.RemoveFriendRequest(storage.UserId(sender), storage.UserId(target)); err != nil {
		log.Printf("%s: cannot remove friend request from %s to %s from db err: %v", op, sender, target, err)
		return common.NewErrorWithDescription(RollbackFriendRequestErrorInternal, err.Error())
	}
	log.Printf("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}

func (s *defaultController) SendFriendRequest(sender UserId, target UserId) *common.CodeBasedError[SendFriendRequestErrorCode] {
	const op = "friends.defaultController.SendFriendRequest"
	log.Printf("%s: start[sender=%s target=%s]", op, sender, target)
	hasTarget, err := s.storage.IsUserExists(storage.UserId(target))
	if err != nil {
		log.Printf("%s: cannot check if user exists in db err: %v", op, err)
		return common.NewErrorWithDescription(SendFriendRequestErrorInternal, err.Error())
	}
	if !hasTarget {
		log.Printf("%s: user %s does not exists", op, target)
		return common.NewError(SendFriendRequestErrorNoSuchUser)
	}
	hasRequest, err := s.storage.HasFriendRequest(storage.UserId(sender), storage.UserId(target))
	if err != nil {
		log.Printf("%s: cannot check has friend request from %s to %s in db err: %v", op, sender, target, err)
		return common.NewErrorWithDescription(SendFriendRequestErrorInternal, err.Error())
	}
	if hasRequest {
		log.Printf("%s: already have friend request from %s to %s", op, sender, target)
		return common.NewError(SendFriendRequestErrorAlreadySent)
	}
	hasIncomingRequest, err := s.storage.HasFriendRequest(storage.UserId(target), storage.UserId(sender))
	if err != nil {
		log.Printf("%s: cannot check has friend request from %s to %s in db err: %v", op, target, sender, err)
		return common.NewErrorWithDescription(SendFriendRequestErrorInternal, err.Error())
	}
	if hasIncomingRequest {
		log.Printf("%s: already have friend request from %s to %s", op, target, sender)
		return common.NewError(SendFriendRequestErrorHaveIncomingRequest)
	}
	isFriends, err := s.storage.HasFriendship(storage.UserId(target), storage.UserId(sender))
	if err != nil {
		log.Printf("%s: cannot check friendship between %s and %s in db err: %v", op, target, sender, err)
		return common.NewErrorWithDescription(SendFriendRequestErrorInternal, err.Error())
	}
	if isFriends {
		log.Printf("%s: already have friendship between %s and %s", op, target, sender)
		return common.NewError(SendFriendRequestErrorAlreadyFriends)
	}
	log.Printf("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}

func (s *defaultController) Unfriend(sender UserId, target UserId) *common.CodeBasedError[UnfriendErrorCode] {
	const op = "friends.defaultController.Unfriend"
	log.Printf("%s: start[sender=%s target=%s]", op, sender, target)
	hasTarget, err := s.storage.IsUserExists(storage.UserId(target))
	if err != nil {
		log.Printf("%s: cannot check if user exists in db err: %v", op, err)
		return common.NewErrorWithDescription(UnfriendErrorInternal, err.Error())
	} else if !hasTarget {
		log.Printf("%s: user %s does not exists", op, target)
		return common.NewError(UnfriendErrorNoSuchUser)
	}
	isFriends, err := s.storage.HasFriendship(storage.UserId(target), storage.UserId(sender))
	if err != nil {
		log.Printf("%s: cannot check friendship between %s and %s in db err: %v", op, target, sender, err)
		return common.NewErrorWithDescription(UnfriendErrorInternal, err.Error())
	}
	if !isFriends {
		log.Printf("%s: no friendship between %s and %s", op, target, sender)
		return common.NewError(UnfriendErrorNotAFriend)
	}
	log.Printf("%s: success[sender=%s target=%s]", op, sender, target)
	return nil
}
