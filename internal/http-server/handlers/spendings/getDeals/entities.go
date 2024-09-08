package getDeals

import (
	"accounty/internal/http-server/responses"
	"accounty/internal/storage"
)

type Request struct {
	Counterparty storage.UserId `json:"counterparty"`
}

type Error struct {
	responses.Error
}

func Success(deals []storage.IdentifiableDeal) responses.Response[[]storage.IdentifiableDeal] {
	return responses.Success(deals)
}

func Failure(err Error) responses.Response[responses.Error] {
	return responses.Failure(err.Error)
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
