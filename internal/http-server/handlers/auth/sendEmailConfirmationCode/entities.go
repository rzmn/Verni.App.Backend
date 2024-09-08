package sendEmailConfirmationCode

import (
	"accounty/internal/http-server/responses"
)

type Request struct{}

type Error struct {
	responses.Error
}

func Success() responses.VoidResponse {
	return responses.OK()
}

func Failure(err Error) responses.Response[responses.Error] {
	return responses.Failure(err.Error)
}

func ErrAlreadyConfirmed() Error {
	return Error{responses.Error{Code: responses.CodeAlreadyConfirmed}}
}

func ErrNotDelivered() Error {
	return Error{responses.Error{Code: responses.CodeNotDelivered}}
}

func ErrBadRequest(description string) Error {
	return Error{responses.Error{Code: responses.CodeBadRequest, Description: &description}}
}

func ErrInternal() Error {
	return Error{responses.Error{Code: responses.CodeInternal}}
}
