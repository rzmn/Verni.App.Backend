package get

import (
	"accounty/internal/http-server/responses"
	"accounty/internal/storage"
)

type Request struct {
	Ids []storage.AvatarId `json:"ids"`
}

type Error struct {
	responses.Error
}

func Success(avatars map[storage.AvatarId]storage.AvatarData) responses.Response[map[storage.AvatarId]storage.AvatarData] {
	return responses.Success(avatars)
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
