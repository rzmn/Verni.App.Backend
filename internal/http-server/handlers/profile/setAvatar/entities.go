package setAvatar

import (
	"accounty/internal/http-server/responses"
	"accounty/internal/storage"
)

type Request struct {
	DataBase64 string `json:"dataBase64"`
}

type Error struct {
	responses.Error
}

func Success(aid storage.AvatarId) responses.Response[storage.AvatarId] {
	return responses.Success(aid)
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
