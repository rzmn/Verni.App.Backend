package updateEmail

import (
	"accounty/internal/http-server/responses"
	"accounty/internal/storage"
)

type Request struct {
	Email string `json:"email"`
}

type Error struct {
	responses.Error
}

func Success(token storage.AuthenticatedSession) responses.Response[storage.AuthenticatedSession] {
	return responses.Success(token)
}

func Failure(err Error) responses.Response[responses.Error] {
	return responses.Failure(err.Error)
}

func ErrInternal() Error {
	return Error{responses.Error{Code: responses.CodeInternal}}
}

func ErrBadRequest(description string) Error {
	return Error{responses.Error{Code: responses.CodeBadRequest, Description: &description}}
}

func ErrWrongFormat() Error {
	return Error{responses.Error{Code: responses.CodeWrongFormat}}
}

func ErrCodeAlreadyTaken() Error {
	return Error{responses.Error{Code: responses.CodeAlreadyTaken}}
}
