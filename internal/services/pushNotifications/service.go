package pushNotifications

import (
	"github.com/rzmn/governi/internal/schema"
)

type UserId schema.UserId
type Expense schema.IdentifiableExpense
type ExpenseId schema.ExpenseId
type Cost schema.Cost

type Service interface {
	FriendRequestHasBeenAccepted(receiver UserId, acceptedBy UserId)
	FriendRequestHasBeenReceived(receiver UserId, sentBy UserId)
	NewExpenseReceived(receiver UserId, expense Expense, author UserId)
}
