package refresh

import (
	"accounty/internal/http-server/responses"
	"accounty/internal/storage"
)

type Request struct {
	RefreshToken string `json:"refreshToken"`
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

func ErrTokenExpired() Error {
	return Error{responses.Error{Code: responses.CodeTokenExpired}}
}

func ErrWrongAccessToken() Error {
	return Error{responses.Error{Code: responses.CodeWrongAccessToken}}
}

func ErrBadRequest(description string) Error {
	return Error{responses.Error{Code: responses.CodeBadRequest, Description: &description}}
}

func ErrInternal() Error {
	return Error{responses.Error{Code: responses.CodeInternal}}
}
