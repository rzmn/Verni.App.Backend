package schema

import (
	"fmt"
	"verni/internal/common"
)

type Code int
type Error struct {
	Code        Code    `json:"code"`
	Description *string `json:"description,omitempty"`
}

func (e *Error) Error() string {
	base := fmt.Sprintf("%d [%s]", e.Code, e.Code.Message())
	if e.Description != nil {
		return fmt.Sprintf("%s - %s", base, *e.Description)
	} else {
		return base
	}
}

type VoidResponse struct {
	Status string `json:"status"`
}

type Response[T any] struct {
	Response T      `json:"response"`
	Status   string `json:"status"`
}

func OK() VoidResponse {
	return VoidResponse{
		Status: "ok",
	}
}

func Success[T any](result T) Response[T] {
	return Response[T]{
		Status:   "ok",
		Response: result,
	}
}

func Failure(error common.CodeBasedError[Code]) Response[Error] {
	description := error.Error()
	return Response[Error]{
		Status: "failed",
		Response: Error{
			Code:        error.Code,
			Description: &description,
		},
	}
}

const (
	_ Code = iota
	CodeIncorrectCredentials
	CodeWrongFormat
	CodeAlreadyTaken
	CodeTokenExpired
	CodeWrongAccessToken
	CodeInternal
	CodeNoSuchUser
	CodeNoSuchRequest
	CodeAlreadySend
	CodeHaveIncomingRequest
	CodeAlreadyFriends
	CodeNotAFriend
	CodeBadRequest
	CodeExpenseNotFound
	CodeIsNotYourExpense
	CodeNotDelivered
	CodeAlreadyConfirmed
	CodeLongpollNoEvents
)

func (c Code) Message() string {
	switch c {
	case CodeIncorrectCredentials:
		return "login or password is incorrect"
	case CodeWrongFormat:
		return "wrong format"
	case CodeAlreadyTaken:
		return "login already taken"
	case CodeTokenExpired:
		return "token expired"
	case CodeWrongAccessToken:
		return "token is wrong"
	case CodeInternal:
		return "internal error"
	case CodeNoSuchUser:
		return "no such user"
	case CodeNoSuchRequest:
		return "no such request"
	case CodeAlreadySend:
		return "alerady sent friend request"
	case CodeHaveIncomingRequest:
		return "have incoming request"
	case CodeAlreadyFriends:
		return "alerady friends"
	case CodeBadRequest:
		return "bad request"
	case CodeExpenseNotFound:
		return "expense not found"
	case CodeNotAFriend:
		return "should be friends"
	case CodeIsNotYourExpense:
		return "not your expense"
	default:
		return "unknown error"
	}
}
