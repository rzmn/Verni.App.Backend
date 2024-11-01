package friends_mock

import (
	"verni/internal/repositories"
	"verni/internal/repositories/friends"
)

type RepositoryMock struct {
	GetFriendsImpl          func(userId friends.UserId) ([]friends.UserId, error)
	GetSubscribersImpl      func(userId friends.UserId) ([]friends.UserId, error)
	GetSubscriptionsImpl    func(userId friends.UserId) ([]friends.UserId, error)
	GetStatusesImpl         func(sender friends.UserId, ids []friends.UserId) (map[friends.UserId]friends.FriendStatus, error)
	HasFriendRequestImpl    func(sender friends.UserId, target friends.UserId) (bool, error)
	StoreFriendRequestImpl  func(sender friends.UserId, target friends.UserId) repositories.MutationWorkItem
	RemoveFriendRequestImpl func(sender friends.UserId, target friends.UserId) repositories.MutationWorkItem
}

func (c *RepositoryMock) GetFriends(userId friends.UserId) ([]friends.UserId, error) {
	return c.GetFriendsImpl(userId)
}

func (c *RepositoryMock) GetSubscribers(userId friends.UserId) ([]friends.UserId, error) {
	return c.GetSubscribersImpl(userId)
}

func (c *RepositoryMock) GetSubscriptions(userId friends.UserId) ([]friends.UserId, error) {
	return c.GetSubscriptionsImpl(userId)
}

func (c *RepositoryMock) GetStatuses(sender friends.UserId, ids []friends.UserId) (map[friends.UserId]friends.FriendStatus, error) {
	return c.GetStatusesImpl(sender, ids)
}

func (c *RepositoryMock) HasFriendRequest(sender friends.UserId, target friends.UserId) (bool, error) {
	return c.HasFriendRequestImpl(sender, target)
}

func (c *RepositoryMock) StoreFriendRequest(sender friends.UserId, target friends.UserId) repositories.MutationWorkItem {
	return c.StoreFriendRequestImpl(sender, target)
}

func (c *RepositoryMock) RemoveFriendRequest(sender friends.UserId, target friends.UserId) repositories.MutationWorkItem {
	return c.RemoveFriendRequestImpl(sender, target)
}
