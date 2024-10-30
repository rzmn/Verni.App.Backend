package jwt_mock

import (
	"verni/internal/services/jwt"
)

type ServiceMock struct {
	IssueRefreshTokenImpl      func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error)
	IssueAccessTokenImpl       func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error)
	ValidateRefreshTokenImpl   func(token jwt.RefreshToken) *jwt.Error
	ValidateAccessTokenImpl    func(token jwt.AccessToken) *jwt.Error
	GetRefreshTokenSubjectImpl func(token jwt.RefreshToken) (jwt.Subject, *jwt.Error)
	GetAccessTokenSubjectImpl  func(token jwt.AccessToken) (jwt.Subject, *jwt.Error)
}

func (s *ServiceMock) IssueRefreshToken(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
	return s.IssueRefreshTokenImpl(subject)
}

func (s *ServiceMock) IssueAccessToken(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
	return s.IssueAccessTokenImpl(subject)
}

func (s *ServiceMock) ValidateRefreshToken(token jwt.RefreshToken) *jwt.Error {
	return s.ValidateRefreshTokenImpl(token)
}

func (s *ServiceMock) ValidateAccessToken(token jwt.AccessToken) *jwt.Error {
	return s.ValidateAccessTokenImpl(token)
}

func (s *ServiceMock) GetRefreshTokenSubject(token jwt.RefreshToken) (jwt.Subject, *jwt.Error) {
	return s.GetRefreshTokenSubjectImpl(token)
}

func (s *ServiceMock) GetAccessTokenSubject(token jwt.AccessToken) (jwt.Subject, *jwt.Error) {
	return s.GetAccessTokenSubjectImpl(token)
}
