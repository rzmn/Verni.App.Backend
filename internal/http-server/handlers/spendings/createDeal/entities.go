package createDeal

import (
	"accounty/internal/http-server/responses"
	"accounty/internal/storage"
)

type Request struct {
	Deal storage.Deal `json:"deal"`
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

func ErrNoSuchUser() Error {
	return Error{responses.Error{Code: responses.CodeNoSuchUser}}
}

func ErrNotAFriend() Error {
	return Error{responses.Error{Code: responses.CodeNotAFriend}}
}

func ErrBadRequest(description string) Error {
	return Error{responses.Error{Code: responses.CodeBadRequest, Description: &description}}
}

func ErrInternal() Error {
	return Error{responses.Error{Code: responses.CodeInternal}}
}
