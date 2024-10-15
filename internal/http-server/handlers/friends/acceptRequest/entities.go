package acceptRequest

import (
	"verni/internal/http-server/responses"
	"verni/internal/storage"
)

type Request struct {
	Sender storage.UserId `json:"sender"`
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

func ErrNoSuchRequest() Error {
	return Error{responses.Error{Code: responses.CodeNoSuchRequest}}
}

func ErrBadRequest(description string) Error {
	return Error{responses.Error{Code: responses.CodeBadRequest, Description: &description}}
}

func ErrInternal() Error {
	return Error{responses.Error{Code: responses.CodeInternal}}
}
