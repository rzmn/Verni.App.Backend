package confirmEmail

import (
	"accounty/internal/http-server/responses"
)

type Request struct {
	Code string `json:"code"`
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

func ErrInternal() Error {
	return Error{responses.Error{Code: responses.CodeInternal}}
}

func ErrBadRequest(description string) Error {
	return Error{responses.Error{Code: responses.CodeBadRequest, Description: &description}}
}

func ErrIncorrect() Error {
	return Error{responses.Error{Code: responses.CodeIncorrectCredentials}}
}
