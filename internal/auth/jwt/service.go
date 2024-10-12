package jwt

import (
	"fmt"
	"time"
)

type ErrorCode int

const (
	_ ErrorCode = iota
	CodeTokenInvalid
	CodeTokenExpired
	CodeInternal
)

type Error struct {
	Code        ErrorCode
	Description *string
}

func (e *Error) Error() string {
	base := fmt.Sprintf("%d [%s]", e.Code, e.Code.Message())
	if e.Description != nil {
		return fmt.Sprintf("%s - %s", base, *e.Description)
	} else {
		return base
	}
}

func (c ErrorCode) Message() string {
	switch c {
	case CodeTokenExpired:
		return "token expired"
	case CodeTokenInvalid:
		return "token has invalid format"
	case CodeInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}

type Subject string
type AccessToken string
type RefreshToken string

type Service interface {
	IssueRefreshToken(subject Subject) (RefreshToken, *Error)
	IssueAccessToken(subject Subject) (AccessToken, *Error)

	ValidateRefreshToken(token RefreshToken) *Error
	ValidateAccessToken(token AccessToken) *Error

	GetRefreshTokenSubject(token RefreshToken) (Subject, *Error)
	GetAccessTokenSubject(token AccessToken) (Subject, *Error)
}

func DefaultService(
	refreshTokenLifetime time.Duration,
	accessTokenLifetime time.Duration,
	currentTime func() time.Time,
) Service {
	return &defaultService{
		refreshTokenLifetime: refreshTokenLifetime,
		accessTokenLifetime:  accessTokenLifetime,
		currentTime:          currentTime,
	}
}
