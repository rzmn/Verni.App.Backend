package friends_test

import (
	"testing"
	"verni/internal/repositories"
	"verni/internal/repositories/friends"

	"github.com/google/uuid"
)

var (
	_s *friends.Repository
)

func getRepository(t *testing.T) friends.Repository {
	if _s != nil {
		return *_s
	}
	repository, err := friends.PostgresRepository(
		repositories.PostgresConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "tester",
			Password: "test_password",
			DbName:   "mydb",
		},
	)
	if err != nil {
		t.Fatalf("failed to init repository err: %v", err)
	}
	_s = &repository
	return repository
}

func randomUid() friends.UserId {
	return friends.UserId(uuid.New().String())
}

func TestStoreAndRemoveFriendRequest(t *testing.T) {
	s := getRepository(t)
	sender := randomUid()
	target := randomUid()

	storeTransaction := s.StoreFriendRequest(sender, target)
	if err := storeTransaction.Perform(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	exists, err := s.HasFriendRequest(sender, target)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !exists {
		t.Fatalf("exists should be true")
	}
	exists, err = s.HasFriendRequest(target, sender)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if exists {
		t.Fatalf("exists should be false")
	}
	removeTransaction := s.RemoveFriendRequest(sender, target)
	if err := removeTransaction.Perform(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	exists, err = s.HasFriendRequest(sender, target)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if exists {
		t.Fatalf("exists should be false")
	}
}

func TestHasFriendRequestEmpty(t *testing.T) {
	s := getRepository(t)
	sender := randomUid()
	target := randomUid()
	exists, err := s.HasFriendRequest(sender, target)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if exists {
		t.Fatalf("exists should be false")
	}
}

func TestGetSubscribers(t *testing.T) {
	s := getRepository(t)
	sender := randomUid()
	target := randomUid()

	storeTransaction := s.StoreFriendRequest(sender, target)
	if err := storeTransaction.Perform(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	subscribers, err := s.GetSubscribers(target)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(subscribers) != 1 || subscribers[0] != sender {
		t.Fatalf("subscribers should have only sender, found %v", subscribers)
	}
	subscribers, err = s.GetSubscribers(sender)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(subscribers) != 0 {
		t.Fatalf("subscribers should be empty")
	}
}

func TestGetSubscribersEmpty(t *testing.T) {
	s := getRepository(t)
	uid := randomUid()
	subscribers, err := s.GetSubscribers(uid)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(subscribers) != 0 {
		t.Fatalf("subscribers should be empty")
	}
}

func TestGetSubsriptions(t *testing.T) {
	s := getRepository(t)
	sender := randomUid()
	target := randomUid()

	storeTransaction := s.StoreFriendRequest(sender, target)
	if err := storeTransaction.Perform(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	subsriptions, err := s.GetSubscriptions(sender)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(subsriptions) != 1 || subsriptions[0] != target {
		t.Fatalf("subsriptions should have only target, found %v", subsriptions)
	}
	subsriptions, err = s.GetSubscriptions(target)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(subsriptions) != 0 {
		t.Fatalf("subsriptions should be empty, found %v", subsriptions)
	}
}

func TestGetSubscriptionsEmpty(t *testing.T) {
	s := getRepository(t)
	uid := randomUid()
	subscriptions, err := s.GetSubscriptions(uid)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(subscriptions) != 0 {
		t.Fatalf("incoming should be empty")
	}
}

func TestGetFriends(t *testing.T) {
	s := getRepository(t)
	sender := randomUid()
	target := randomUid()
	storeTransaction := s.StoreFriendRequest(sender, target)
	if err := storeTransaction.Perform(); err != nil {
		t.Fatalf("cannot store request from sender to target err: %v", err)
	}
	friends, err := s.GetFriends(sender)
	if err != nil {
		t.Fatalf("cannot get senders friends err: %v", err)
	}
	if len(friends) != 0 {
		t.Fatalf("senders friends should be empty, found %v", friends)
	}
	friends, err = s.GetFriends(target)
	if err != nil {
		t.Fatalf("cannot get targets friends err: %v", err)
	}
	if len(friends) != 0 {
		t.Fatalf("targets friends should be empty, found %v", friends)
	}
	storeTransaction = s.StoreFriendRequest(target, sender)
	if err := storeTransaction.Perform(); err != nil {
		t.Fatalf("cannot store request from sender to target err: %v", err)
	}
	friends, err = s.GetFriends(sender)
	if err != nil {
		t.Fatalf("cannot get senders friends err: %v", err)
	}
	if len(friends) != 1 || friends[0] != target {
		t.Fatalf("senders friends should contain only target, found %v", friends)
	}
	friends, err = s.GetFriends(target)
	if err != nil {
		t.Fatalf("cannot get targets friends err: %v", err)
	}
	if len(friends) != 1 || friends[0] != sender {
		t.Fatalf("targets friends should contain only sender, found %v", friends)
	}
	removeTransaction := s.RemoveFriendRequest(sender, target)
	if err := removeTransaction.Perform(); err != nil {
		t.Fatalf("cannot remove senders request err: %v", err)
	}
	friends, err = s.GetFriends(sender)
	if err != nil {
		t.Fatalf("cannot get senders friends err: %v", err)
	}
	if len(friends) != 0 {
		t.Fatalf("senders friends should be empty, found %v", friends)
	}
	friends, err = s.GetFriends(target)
	if err != nil {
		t.Fatalf("cannot get targets friends err: %v", err)
	}
	if len(friends) != 0 {
		t.Fatalf("targets friends should be empty, found %v", friends)
	}
}

func TestGetFriendsEmpty(t *testing.T) {
	s := getRepository(t)
	uid := randomUid()
	friends, err := s.GetFriends(uid)
	if err != nil {
		t.Fatalf("cannot get uids friends err: %v", err)
	}
	if len(friends) != 0 {
		t.Fatalf("uids friends should be empty, found %v", friends)
	}
}
