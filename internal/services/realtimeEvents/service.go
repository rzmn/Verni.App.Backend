package realtimeEvents

type UserId string

type Service interface {
	CounterpartiesUpdated(uid UserId)
	ExpensesUpdated(uid UserId, counterparty UserId)
	FriendsUpdated(uid UserId)
}
