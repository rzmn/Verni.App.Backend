package deleteDeal

import (
	"accounty/internal/http-server/responses"
	"accounty/internal/storage"
)

type Request struct {
	DealId storage.DealId `json:"dealId"`
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

func ErrDealNotFound() Error {
	return Error{responses.Error{Code: responses.CodeDealNotFound}}
}

func ErrNotAFriend() Error {
	return Error{responses.Error{Code: responses.CodeNotAFriend}}
}

func ErrIsNotYourDeal() Error {
	return Error{responses.Error{Code: responses.CodeIsNotYourDeal}}
}

func ErrBadRequest(description string) Error {
	return Error{responses.Error{Code: responses.CodeBadRequest, Description: &description}}
}

func ErrInternal() Error {
	return Error{responses.Error{Code: responses.CodeInternal}}
}
