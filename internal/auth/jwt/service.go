package jwt

import (
	"time"
)

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
