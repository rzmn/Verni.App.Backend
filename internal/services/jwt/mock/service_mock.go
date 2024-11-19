package jwt_mock

import (
	"github.com/rzmn/Verni.App.Backend/internal/services/jwt"
)

type ServiceMock struct {
	IssueRefreshTokenImpl      func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error)
	IssueAccessTokenImpl       func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error)
	ValidateRefreshTokenImpl   func(token jwt.RefreshToken) *jwt.Error
	ValidateAccessTokenImpl    func(token jwt.AccessToken) *jwt.Error
	GetRefreshTokenSubjectImpl func(token jwt.RefreshToken) (jwt.Subject, *jwt.Error)
	GetAccessTokenSubjectImpl  func(token jwt.AccessToken) (jwt.Subject, *jwt.Error)
}

func (c *ServiceMock) IssueRefreshToken(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
	return c.IssueRefreshTokenImpl(subject)
}

func (c *ServiceMock) IssueAccessToken(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
	return c.IssueAccessTokenImpl(subject)
}

func (c *ServiceMock) ValidateRefreshToken(token jwt.RefreshToken) *jwt.Error {
	return c.ValidateRefreshTokenImpl(token)
}

func (c *ServiceMock) ValidateAccessToken(token jwt.AccessToken) *jwt.Error {
	return c.ValidateAccessTokenImpl(token)
}

func (c *ServiceMock) GetRefreshTokenSubject(token jwt.RefreshToken) (jwt.Subject, *jwt.Error) {
	return c.GetRefreshTokenSubjectImpl(token)
}

func (c *ServiceMock) GetAccessTokenSubject(token jwt.AccessToken) (jwt.Subject, *jwt.Error) {
	return c.GetAccessTokenSubjectImpl(token)
}
