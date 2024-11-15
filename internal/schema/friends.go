package schema

type AcceptFriendRequest struct {
	Sender UserId `json:"sender"`
}

type GetFriendsRequest struct {
	Statuses []FriendStatus `json:"statuses"`
}

type RejectFriendRequest struct {
	Sender UserId `json:"sender"`
}

type RollbackFriendRequest struct {
	Target UserId `json:"target"`
}

type SendFriendRequest struct {
	Target UserId `json:"target"`
}

type UnfriendRequest struct {
	Target UserId `json:"target"`
}
