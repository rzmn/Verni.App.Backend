package search

import (
	"accounty/internal/http-server/responses"
	"accounty/internal/storage"
)

type Request struct {
	Query string `json:"query"`
}

type Error struct {
	responses.Error
}

func Success(users []storage.User) responses.Response[[]storage.User] {
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
