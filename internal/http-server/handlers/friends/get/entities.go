package get

import (
	"accounty/internal/http-server/responses"
	"accounty/internal/storage"
)

type Error struct {
	responses.Error
}

type Status int

const (
	_ Status = iota
	FriendStatusFriends
	FriendStatusSubscription
	FriendStatusSubscriber
)

type Request struct {
	Statuses []Status `json:"statuses"`
}

func Success(users map[Status][]storage.UserId) responses.Response[map[Status][]storage.UserId] {
	return responses.Success(users)
}

func Failure(err Error) responses.Response[responses.Error] {
	return responses.Failure(err.Error)
}

func ErrBadRequest(description string) Error {
	return Error{responses.Error{Code: responses.CodeBadRequest, Description: &description}}
}

func ErrInternal() Error {
	return Error{responses.Error{Code: responses.CodeInternal}}
}
