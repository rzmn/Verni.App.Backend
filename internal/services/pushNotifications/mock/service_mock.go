package pushNotifications_mock

import "verni/internal/services/pushNotifications"

type ServiceMock struct {
	FriendRequestHasBeenAcceptedImpl func(receiver pushNotifications.UserId, acceptedBy pushNotifications.UserId)
	FriendRequestHasBeenReceivedImpl func(receiver pushNotifications.UserId, sentBy pushNotifications.UserId)
	NewExpenseReceivedImpl           func(receiver pushNotifications.UserId, expense pushNotifications.Expense, author pushNotifications.UserId)
}

func (s *ServiceMock) FriendRequestHasBeenAccepted(receiver pushNotifications.UserId, acceptedBy pushNotifications.UserId) {
	s.FriendRequestHasBeenAcceptedImpl(receiver, acceptedBy)
}

func (s *ServiceMock) FriendRequestHasBeenReceived(receiver pushNotifications.UserId, sentBy pushNotifications.UserId) {
	s.FriendRequestHasBeenReceivedImpl(receiver, sentBy)
}

func (s *ServiceMock) NewExpenseReceived(receiver pushNotifications.UserId, expense pushNotifications.Expense, author pushNotifications.UserId) {
	s.NewExpenseReceivedImpl(receiver, expense, author)
}
