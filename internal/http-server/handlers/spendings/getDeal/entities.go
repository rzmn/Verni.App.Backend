package getDeal

import (
	"accounty/internal/http-server/responses"
	"accounty/internal/storage"
)

type Request struct {
	Id storage.DealId `json:"dealId"`
}

type Error struct {
	responses.Error
}

func Success(users storage.Deal) responses.Response[storage.Deal] {
	return responses.Success(users)
}

func Failure(err Error) responses.Response[responses.Error] {
	return responses.Failure(err.Error)
}

func ErrBadRequest(description string) Error {
	return Error{responses.Error{Code: responses.CodeBadRequest, Description: &description}}
}

func ErrDealNotFound() Error {
	return Error{responses.Error{Code: responses.CodeDealNotFound}}
}

func ErrInternal() Error {
	return Error{responses.Error{Code: responses.CodeInternal}}
}
