package pushNotifications_mock

import "github.com/rzmn/Verni.App.Backend/internal/services/pushNotifications"

type ServiceMock struct {
	FriendRequestHasBeenAcceptedImpl func(receiver pushNotifications.UserId, acceptedBy pushNotifications.UserId)
	FriendRequestHasBeenReceivedImpl func(receiver pushNotifications.UserId, sentBy pushNotifications.UserId)
	NewExpenseReceivedImpl           func(receiver pushNotifications.UserId, expense pushNotifications.Expense, author pushNotifications.UserId)
}

func (c *ServiceMock) FriendRequestHasBeenAccepted(receiver pushNotifications.UserId, acceptedBy pushNotifications.UserId) {
	c.FriendRequestHasBeenAcceptedImpl(receiver, acceptedBy)
}

func (c *ServiceMock) FriendRequestHasBeenReceived(receiver pushNotifications.UserId, sentBy pushNotifications.UserId) {
	c.FriendRequestHasBeenReceivedImpl(receiver, sentBy)
}

func (c *ServiceMock) NewExpenseReceived(receiver pushNotifications.UserId, expense pushNotifications.Expense, author pushNotifications.UserId) {
	c.NewExpenseReceivedImpl(receiver, expense, author)
}
