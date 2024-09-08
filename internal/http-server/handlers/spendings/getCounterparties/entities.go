package getCounterparties

import (
	"accounty/internal/http-server/responses"
	"accounty/internal/storage"
)

type Request struct{}

type Error struct {
	responses.Error
}

func Success(previews []storage.SpendingsPreview) responses.Response[[]storage.SpendingsPreview] {
	return responses.Success(previews)
}

func Failure(err Error) responses.Response[responses.Error] {
	return responses.Failure(err.Error)
}

func ErrInternal() Error {
	return Error{responses.Error{Code: responses.CodeInternal}}
}
