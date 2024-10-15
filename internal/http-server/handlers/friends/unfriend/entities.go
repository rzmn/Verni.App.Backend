package unfriend

import (
	"verni/internal/http-server/responses"
	"verni/internal/storage"
)

type Request struct {
	Target storage.UserId `json:"target"`
}

type Error struct {
	responses.Error
}

func Success() responses.VoidResponse {
	return responses.OK()
}

func Failure(err Error) responses.Response[responses.Error] {
	return responses.Failure(err.Error)
}

func ErrNotAFriend() Error {
	return Error{responses.Error{Code: responses.CodeNotAFriend}}
}

func ErrNoSuchUser() Error {
	return Error{responses.Error{Code: responses.CodeNoSuchUser}}
}

func ErrBadRequest(description string) Error {
	return Error{responses.Error{Code: responses.CodeBadRequest, Description: &description}}
}

func ErrInternal() Error {
	return Error{responses.Error{Code: responses.CodeInternal}}
}
