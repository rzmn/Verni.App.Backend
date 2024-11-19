package defaultController_test

import (
	"errors"
	"testing"

	"github.com/rzmn/governi/internal/controllers/friends"
	defaultController "github.com/rzmn/governi/internal/controllers/friends/default"
	"github.com/rzmn/governi/internal/repositories"
	friendsRepository "github.com/rzmn/governi/internal/repositories/friends"
	friends_mock "github.com/rzmn/governi/internal/repositories/friends/mock"
	standartOutputLoggingService "github.com/rzmn/governi/internal/services/logging/standartOutput"

	"github.com/google/uuid"
)

func TestAcceptRequestFailedToCheckIfRequestExists(t *testing.T) {
	repository := friends_mock.RepositoryMock{
		HasFriendRequestImpl: func(sender friendsRepository.UserId, target friendsRepository.UserId) (bool, error) {
			return false, errors.New("some error")
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.AcceptFriendRequest(friends.UserId(uuid.New().String()), friends.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`AcceptFriendRequest` should fail, found nil err")
	}
	if err.Code != friends.AcceptFriendRequestErrorInternal {
		t.Fatalf("`AcceptFriendRequest` should fail with `internal`, found %v", err)
	}
}

func TestAcceptRequestFailedNoSuchRequest(t *testing.T) {
	repository := friends_mock.RepositoryMock{
		HasFriendRequestImpl: func(sender friendsRepository.UserId, target friendsRepository.UserId) (bool, error) {
			return false, nil
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.AcceptFriendRequest(friends.UserId(uuid.New().String()), friends.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`AcceptFriendRequest` should fail, found nil err")
	}
	if err.Code != friends.AcceptFriendRequestErrorNoSuchRequest {
		t.Fatalf("`AcceptFriendRequest` should fail with `no such request`, found %v", err)
	}
}

func TestAcceptRequestFailedToAccept(t *testing.T) {
	repository := friends_mock.RepositoryMock{
		HasFriendRequestImpl: func(sender friendsRepository.UserId, target friendsRepository.UserId) (bool, error) {
			return true, nil
		},
		StoreFriendRequestImpl: func(sender, target friendsRepository.UserId) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return errors.New("some error")
				},
			}
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.AcceptFriendRequest(friends.UserId(uuid.New().String()), friends.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`AcceptFriendRequest` should fail, found nil err")
	}
	if err.Code != friends.AcceptFriendRequestErrorInternal {
		t.Fatalf("`AcceptFriendRequest` should fail with `internal`, found %v", err)
	}
}

func TestAcceptRequestOk(t *testing.T) {
	storeCalls := 0
	repository := friends_mock.RepositoryMock{
		HasFriendRequestImpl: func(sender friendsRepository.UserId, target friendsRepository.UserId) (bool, error) {
			return true, nil
		},
		StoreFriendRequestImpl: func(sender, target friendsRepository.UserId) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					storeCalls += 1
					return nil
				},
			}
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.AcceptFriendRequest(friends.UserId(uuid.New().String()), friends.UserId(uuid.New().String()))
	if err != nil {
		t.Fatalf("`AcceptFriendRequest` should not fail, found err %v", err)
	}
	if storeCalls != 1 {
		t.Fatalf("should accept request once, found %d", storeCalls)
	}
}

func TestGetFriendsGetOnlyRequestedStatus(t *testing.T) {
	testGetFriendsGetOnlyRequestedStatus(t, []friends.FriendStatus{})
	testGetFriendsGetOnlyRequestedStatus(t, []friends.FriendStatus{friends.FriendStatusSubscriber})
	testGetFriendsGetOnlyRequestedStatus(t, []friends.FriendStatus{friends.FriendStatusSubscription})
	testGetFriendsGetOnlyRequestedStatus(t, []friends.FriendStatus{friends.FriendStatusFriends})
	testGetFriendsGetOnlyRequestedStatus(t, []friends.FriendStatus{friends.FriendStatusSubscriber, friends.FriendStatusSubscription})
	testGetFriendsGetOnlyRequestedStatus(t, []friends.FriendStatus{friends.FriendStatusSubscriber, friends.FriendStatusFriends})
	testGetFriendsGetOnlyRequestedStatus(t, []friends.FriendStatus{friends.FriendStatusSubscription, friends.FriendStatusFriends})
	testGetFriendsGetOnlyRequestedStatus(t, []friends.FriendStatus{friends.FriendStatusSubscriber, friends.FriendStatusSubscription, friends.FriendStatusFriends})
}

func testGetFriendsGetOnlyRequestedStatus(t *testing.T, statuses []friends.FriendStatus) {
	getSubscribersCalls := 0
	getSubscriptionsCalls := 0
	getFriendsCalls := 0
	repository := friends_mock.RepositoryMock{
		GetSubscribersImpl: func(userId friendsRepository.UserId) ([]friendsRepository.UserId, error) {
			if getSubscribersCalls >= 1 {
				t.Fatalf("get subscribers should be called at most once")
			}
			getSubscribersCalls += 1
			for _, status := range statuses {
				if status == friends.FriendStatusSubscriber {
					return []friendsRepository.UserId{
						friendsRepository.UserId(uuid.New().String()),
					}, nil
				}
			}
			return []friendsRepository.UserId{}, errors.New("some error")
		},
		GetSubscriptionsImpl: func(userId friendsRepository.UserId) ([]friendsRepository.UserId, error) {
			if getSubscriptionsCalls >= 1 {
				t.Fatalf("get subscriptions should be called at most once")
			}
			getSubscriptionsCalls += 1
			for _, status := range statuses {
				if status == friends.FriendStatusSubscription {
					return []friendsRepository.UserId{
						friendsRepository.UserId(uuid.New().String()),
					}, nil
				}
			}
			return []friendsRepository.UserId{}, errors.New("some error")
		},
		GetFriendsImpl: func(userId friendsRepository.UserId) ([]friendsRepository.UserId, error) {
			if getFriendsCalls >= 1 {
				t.Fatalf("get friends should be called at most once")
			}
			getFriendsCalls += 1
			for _, status := range statuses {
				if status == friends.FriendStatusFriends {
					return []friendsRepository.UserId{
						friendsRepository.UserId(uuid.New().String()),
					}, nil
				}
			}
			return []friendsRepository.UserId{}, errors.New("some error")
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	friendsMap, err := controller.GetFriends(statuses, friends.UserId(uuid.New().String()))
	if err != nil {
		t.Fatalf("`GetFriends` should not be failed, found err: %v", err)
	}
	callsCount := getSubscribersCalls + getSubscriptionsCalls + getFriendsCalls
	if callsCount != len(statuses) {
		t.Fatalf("calls count should be equal to statuses count (%d), found: %d", len(statuses), callsCount)
	}
	if len(friendsMap) != len(statuses) {
		t.Fatalf("keys count should be equal to statuses count, found: %v", friendsMap)
	}
	for _, status := range statuses {
		if len(friendsMap[status]) != 1 {
			t.Fatalf("each status users list should contain 1 element, found: %v", friendsMap[status])
		}
	}
}

func TestGetFriendsGetFailedEachStatus(t *testing.T) {
	repository := friends_mock.RepositoryMock{
		GetSubscribersImpl: func(userId friendsRepository.UserId) ([]friendsRepository.UserId, error) {
			return []friendsRepository.UserId{}, errors.New("some error")
		},
		GetSubscriptionsImpl: func(userId friendsRepository.UserId) ([]friendsRepository.UserId, error) {
			return []friendsRepository.UserId{}, errors.New("some error")
		},
		GetFriendsImpl: func(userId friendsRepository.UserId) ([]friendsRepository.UserId, error) {
			return []friendsRepository.UserId{}, errors.New("some error")
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	testGetFriendsGetFailed(t, controller, friends.FriendStatusSubscriber)
	testGetFriendsGetFailed(t, controller, friends.FriendStatusSubscription)
	testGetFriendsGetFailed(t, controller, friends.FriendStatusFriends)
}

func testGetFriendsGetFailed(t *testing.T, controller friends.Controller, status friends.FriendStatus) {
	_, err := controller.GetFriends([]friends.FriendStatus{status}, friends.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`GetFriends` should fail, found success")
	}
	if err.Code != friends.GetFriendsErrorInternal {
		t.Fatalf("`GetFriends` should fail with `internal`, found %v", err)
	}
}

func TestRollbackFailedToCheckIfRequestExists(t *testing.T) {
	repository := friends_mock.RepositoryMock{
		HasFriendRequestImpl: func(sender friendsRepository.UserId, target friendsRepository.UserId) (bool, error) {
			return false, errors.New("some error")
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.RollbackFriendRequest(friends.UserId(uuid.New().String()), friends.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`RollbackFriendRequest` should fail, found nil err")
	}
	if err.Code != friends.RollbackFriendRequestErrorInternal {
		t.Fatalf("`RollbackFriendRequest` should fail with `internal`, found %v", err)
	}
}

func TestRollbackFailedNoSuchRequest(t *testing.T) {
	repository := friends_mock.RepositoryMock{
		HasFriendRequestImpl: func(sender friendsRepository.UserId, target friendsRepository.UserId) (bool, error) {
			return false, nil
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.RollbackFriendRequest(friends.UserId(uuid.New().String()), friends.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`RollbackFriendRequest` should fail, found nil err")
	}
	if err.Code != friends.RollbackFriendRequestErrorNoSuchRequest {
		t.Fatalf("`RollbackFriendRequest` should fail with `internal`, found %v", err)
	}
}

func TestRollbackFailedToRemove(t *testing.T) {
	repository := friends_mock.RepositoryMock{
		HasFriendRequestImpl: func(sender friendsRepository.UserId, target friendsRepository.UserId) (bool, error) {
			return true, nil
		},
		RemoveFriendRequestImpl: func(sender, target friendsRepository.UserId) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return errors.New("some error")
				},
			}
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.RollbackFriendRequest(friends.UserId(uuid.New().String()), friends.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`RollbackFriendRequest` should fail, found nil err")
	}
	if err.Code != friends.RollbackFriendRequestErrorInternal {
		t.Fatalf("`RollbackFriendRequest` should fail with `internal`, found %v", err)
	}
}

func TestRollbackOk(t *testing.T) {
	storeCalls := 0
	repository := friends_mock.RepositoryMock{
		HasFriendRequestImpl: func(sender friendsRepository.UserId, target friendsRepository.UserId) (bool, error) {
			return true, nil
		},
		RemoveFriendRequestImpl: func(sender, target friendsRepository.UserId) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					storeCalls += 1
					return nil
				},
			}
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.RollbackFriendRequest(friends.UserId(uuid.New().String()), friends.UserId(uuid.New().String()))
	if err != nil {
		t.Fatalf("`RollbackFriendRequest` should not fail, found err %v", err)
	}
	if storeCalls != 1 {
		t.Fatalf("should accept request once, found %d", storeCalls)
	}
}

func TestSendRequestFailedToCheckStatus(t *testing.T) {
	repository := friends_mock.RepositoryMock{
		GetStatusesImpl: func(sender friendsRepository.UserId, ids []friendsRepository.UserId) (map[friendsRepository.UserId]friendsRepository.FriendStatus, error) {
			return map[friendsRepository.UserId]friendsRepository.FriendStatus{}, errors.New("some error")
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.SendFriendRequest(friends.UserId(uuid.New().String()), friends.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`SendFriendRequest` should fail, found nil err")
	}
	if err.Code != friends.SendFriendRequestErrorInternal {
		t.Fatalf("`SendFriendRequest` should fail with `internal`, found %v", err)
	}
}

func TestSendRequestFailedUnknownStatus(t *testing.T) {
	repository := friends_mock.RepositoryMock{
		GetStatusesImpl: func(sender friendsRepository.UserId, ids []friendsRepository.UserId) (map[friendsRepository.UserId]friendsRepository.FriendStatus, error) {
			return map[friendsRepository.UserId]friendsRepository.FriendStatus{}, nil
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.SendFriendRequest(friends.UserId(uuid.New().String()), friends.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`SendFriendRequest` should fail, found nil err")
	}
	if err.Code != friends.SendFriendRequestErrorIncorrectUserStatus {
		t.Fatalf("`SendFriendRequest` should fail with `incorrect status`, found %v", err)
	}
}

func TestSendRequestFailedTargetIsMe(t *testing.T) {
	target := friends.UserId(uuid.New().String())
	repository := friends_mock.RepositoryMock{
		GetStatusesImpl: func(sender friendsRepository.UserId, ids []friendsRepository.UserId) (map[friendsRepository.UserId]friendsRepository.FriendStatus, error) {
			return map[friendsRepository.UserId]friendsRepository.FriendStatus{
				friendsRepository.UserId(target): friendsRepository.FriendStatusMe,
			}, nil
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.SendFriendRequest(friends.UserId(uuid.New().String()), target)
	if err == nil {
		t.Fatalf("`SendFriendRequest` should fail, found nil err")
	}
	if err.Code != friends.SendFriendRequestErrorIncorrectUserStatus {
		t.Fatalf("`SendFriendRequest` should fail with `incorrect status`, found %v", err)
	}
}

func TestSendRequestAlreadySent(t *testing.T) {
	target := friends.UserId(uuid.New().String())
	repository := friends_mock.RepositoryMock{
		GetStatusesImpl: func(sender friendsRepository.UserId, ids []friendsRepository.UserId) (map[friendsRepository.UserId]friendsRepository.FriendStatus, error) {
			return map[friendsRepository.UserId]friendsRepository.FriendStatus{
				friendsRepository.UserId(target): friendsRepository.FriendStatusSubscription,
			}, nil
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.SendFriendRequest(friends.UserId(uuid.New().String()), target)
	if err == nil {
		t.Fatalf("`SendFriendRequest` should fail, found nil err")
	}
	if err.Code != friends.SendFriendRequestErrorAlreadySent {
		t.Fatalf("`SendFriendRequest` should fail with `already sent`, found %v", err)
	}
}

func TestSendRequestHaveIncoming(t *testing.T) {
	target := friends.UserId(uuid.New().String())
	repository := friends_mock.RepositoryMock{
		GetStatusesImpl: func(sender friendsRepository.UserId, ids []friendsRepository.UserId) (map[friendsRepository.UserId]friendsRepository.FriendStatus, error) {
			return map[friendsRepository.UserId]friendsRepository.FriendStatus{
				friendsRepository.UserId(target): friendsRepository.FriendStatusSubscriber,
			}, nil
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.SendFriendRequest(friends.UserId(uuid.New().String()), target)
	if err == nil {
		t.Fatalf("`SendFriendRequest` should fail, found nil err")
	}
	if err.Code != friends.SendFriendRequestErrorHaveIncomingRequest {
		t.Fatalf("`SendFriendRequest` should fail with `have incoming`, found %v", err)
	}
}

func TestSendRequestFailedToStoreRequest(t *testing.T) {
	target := friends.UserId(uuid.New().String())
	repository := friends_mock.RepositoryMock{
		GetStatusesImpl: func(sender friendsRepository.UserId, ids []friendsRepository.UserId) (map[friendsRepository.UserId]friendsRepository.FriendStatus, error) {
			return map[friendsRepository.UserId]friendsRepository.FriendStatus{
				friendsRepository.UserId(target): friendsRepository.FriendStatusNo,
			}, nil
		},
		StoreFriendRequestImpl: func(sender, target friendsRepository.UserId) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return errors.New("some error")
				},
			}
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.SendFriendRequest(friends.UserId(uuid.New().String()), target)
	if err == nil {
		t.Fatalf("`SendFriendRequest` should fail, found nil err")
	}
	if err.Code != friends.SendFriendRequestErrorInternal {
		t.Fatalf("`SendFriendRequest` should fail with `internal`, found %v", err)
	}
}

func TestSendRequestOk(t *testing.T) {
	storeCalls := 0
	target := friends.UserId(uuid.New().String())
	repository := friends_mock.RepositoryMock{
		GetStatusesImpl: func(sender friendsRepository.UserId, ids []friendsRepository.UserId) (map[friendsRepository.UserId]friendsRepository.FriendStatus, error) {
			return map[friendsRepository.UserId]friendsRepository.FriendStatus{
				friendsRepository.UserId(target): friendsRepository.FriendStatusNo,
			}, nil
		},
		StoreFriendRequestImpl: func(sender, target friendsRepository.UserId) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					storeCalls += 1
					return nil
				},
			}
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.SendFriendRequest(friends.UserId(uuid.New().String()), target)
	if err != nil {
		t.Fatalf("`SendFriendRequest` should not fail, found err %v", err)
	}
	if storeCalls != 1 {
		t.Fatalf("`store should be called once, found %d", storeCalls)
	}
}

func TestUnfriendFailedToCheckStatus(t *testing.T) {
	repository := friends_mock.RepositoryMock{
		GetStatusesImpl: func(sender friendsRepository.UserId, ids []friendsRepository.UserId) (map[friendsRepository.UserId]friendsRepository.FriendStatus, error) {
			return map[friendsRepository.UserId]friendsRepository.FriendStatus{}, errors.New("some error")
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.Unfriend(friends.UserId(uuid.New().String()), friends.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`Unfriend` should fail, found nil err")
	}
	if err.Code != friends.UnfriendErrorInternal {
		t.Fatalf("`Unfriend` should fail with `internal`, found %v", err)
	}
}

func TestUnfriendUnknownStatus(t *testing.T) {
	repository := friends_mock.RepositoryMock{
		GetStatusesImpl: func(sender friendsRepository.UserId, ids []friendsRepository.UserId) (map[friendsRepository.UserId]friendsRepository.FriendStatus, error) {
			return map[friendsRepository.UserId]friendsRepository.FriendStatus{}, nil
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.Unfriend(friends.UserId(uuid.New().String()), friends.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`Unfriend` should fail, found nil err")
	}
	if err.Code != friends.UnfriendErrorNotAFriend {
		t.Fatalf("`Unfriend` should fail with `not a friend`, found %v", err)
	}
}

func TestUnfriendNotAFriend(t *testing.T) {
	testUnfriendNotAFriend(t, friendsRepository.FriendStatusMe)
	testUnfriendNotAFriend(t, friendsRepository.FriendStatusNo)
	testUnfriendNotAFriend(t, friendsRepository.FriendStatusSubscriber)
	testUnfriendNotAFriend(t, friendsRepository.FriendStatusSubscription)
}

func testUnfriendNotAFriend(t *testing.T, status friendsRepository.FriendStatus) {
	target := friends.UserId(uuid.New().String())
	repository := friends_mock.RepositoryMock{
		GetStatusesImpl: func(sender friendsRepository.UserId, ids []friendsRepository.UserId) (map[friendsRepository.UserId]friendsRepository.FriendStatus, error) {
			return map[friendsRepository.UserId]friendsRepository.FriendStatus{
				friendsRepository.UserId(target): status,
			}, nil
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.Unfriend(friends.UserId(uuid.New().String()), target)
	if err == nil {
		t.Fatalf("`Unfriend` should fail, found nil err")
	}
	if err.Code != friends.UnfriendErrorNotAFriend {
		t.Fatalf("`Unfriend` should fail with `not a friend`, found %v", err)
	}
}

func TestUnfriendFailedToRemove(t *testing.T) {
	target := friends.UserId(uuid.New().String())
	repository := friends_mock.RepositoryMock{
		GetStatusesImpl: func(sender friendsRepository.UserId, ids []friendsRepository.UserId) (map[friendsRepository.UserId]friendsRepository.FriendStatus, error) {
			return map[friendsRepository.UserId]friendsRepository.FriendStatus{
				friendsRepository.UserId(target): friendsRepository.FriendStatusFriend,
			}, nil
		},
		RemoveFriendRequestImpl: func(sender, target friendsRepository.UserId) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return errors.New("some error")
				},
			}
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.Unfriend(friends.UserId(uuid.New().String()), target)
	if err == nil {
		t.Fatalf("`Unfriend` should fail, found nil err")
	}
	if err.Code != friends.UnfriendErrorInternal {
		t.Fatalf("`Unfriend` should fail with `internal`, found %v", err)
	}
}

func TestUnfriendOk(t *testing.T) {
	removeCalls := 0
	target := friends.UserId(uuid.New().String())
	repository := friends_mock.RepositoryMock{
		GetStatusesImpl: func(sender friendsRepository.UserId, ids []friendsRepository.UserId) (map[friendsRepository.UserId]friendsRepository.FriendStatus, error) {
			return map[friendsRepository.UserId]friendsRepository.FriendStatus{
				friendsRepository.UserId(target): friendsRepository.FriendStatusFriend,
			}, nil
		},
		RemoveFriendRequestImpl: func(sender, target friendsRepository.UserId) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					removeCalls += 1
					return nil
				},
			}
		},
	}
	controller := defaultController.New(&repository, standartOutputLoggingService.New())
	err := controller.Unfriend(friends.UserId(uuid.New().String()), target)
	if err != nil {
		t.Fatalf("`Unfriend` should not fail, found err %v", err)
	}
	if removeCalls != 1 {
		t.Fatalf("`remove should be called once, found %d", removeCalls)
	}
}
