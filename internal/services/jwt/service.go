package jwt

import (
	"time"
	"verni/internal/services/logging"
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

type DefaultConfig struct {
	AccessTokenLifetimeHours  int    `json:"accessTokenLifetimeHours"`
	RefreshTokenLifetimeHours int    `json:"refreshTokenLifetimeHours"`
	RefreshTokenSecret        string `json:"refreshTokenSecret"`
	AccessTokenSecret         string `json:"accessTokenSecret"`
}

func DefaultService(
	config DefaultConfig,
	logger logging.Service,
	currentTime func() time.Time,
) Service {
	return &defaultService{
		refreshTokenLifetime: time.Hour * time.Duration(config.RefreshTokenLifetimeHours),
		accessTokenLifetime:  time.Hour * time.Duration(config.AccessTokenLifetimeHours),
		refreshTokenSecret:   config.RefreshTokenSecret,
		accessTokenSecret:    config.AccessTokenSecret,
		currentTime:          currentTime,
		logger:               logger,
	}
}
