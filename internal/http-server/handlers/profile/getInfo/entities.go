package getInfo

import (
	"accounty/internal/http-server/responses"
	"accounty/internal/storage"
)

type Error struct {
	responses.Error
}

func Success(user storage.ProfileInfo) responses.Response[storage.ProfileInfo] {
	return responses.Success(user)
}

func Failure(err Error) responses.Response[responses.Error] {
	return responses.Failure(err.Error)
}

func ErrInternal() Error {
	return Error{responses.Error{Code: responses.CodeInternal}}
}
