package jwt

import (
	"errors"
	"time"
	"verni/internal/services/logging"

	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenTypeRefresh = "refresh"
	tokenTypeAccess  = "access"
)

type defaultService struct {
	refreshTokenLifetime time.Duration
	accessTokenLifetime  time.Duration
	refreshTokenSecret   string
	accessTokenSecret    string
	currentTime          func() time.Time
	logger               logging.Service
}

func (c *defaultService) IssueRefreshToken(subject Subject) (RefreshToken, *Error) {
	const op = "jwt.defaultService.IssueRefreshToken"
	currentTime := c.currentTime()
	rawToken, err := generateToken(jwtClaims{
		TokenType: tokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   string(subject),
			ExpiresAt: jwt.NewNumericDate(currentTime.Add(c.refreshTokenLifetime)),
			IssuedAt:  jwt.NewNumericDate(currentTime),
		},
	}, []byte(c.refreshTokenSecret))
	if err != nil {
		c.logger.Log("%s: cannot generate token %v", op, err)
		return RefreshToken(rawToken), &Error{
			Code: CodeInternal,
		}
	}
	return RefreshToken(rawToken), nil
}

func (c *defaultService) IssueAccessToken(subject Subject) (AccessToken, *Error) {
	const op = "jwt.defaultService.IssueAccessToken"
	currentTime := c.currentTime()
	rawToken, err := generateToken(jwtClaims{
		TokenType: tokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   string(subject),
			ExpiresAt: jwt.NewNumericDate(currentTime.Add(c.accessTokenLifetime)),
			IssuedAt:  jwt.NewNumericDate(currentTime),
		},
	}, []byte(c.accessTokenSecret))
	if err != nil {
		c.logger.Log("%s: cannot generate token %v", op, err)
		return AccessToken(rawToken), &Error{
			Code: CodeInternal,
		}
	}
	return AccessToken(rawToken), nil
}

func (c *defaultService) ValidateRefreshToken(token RefreshToken) *Error {
	const op = "jwt.defaultService.ValidateRefreshToken"
	rawToken, err := parseToken(string(token), []byte(c.refreshTokenSecret))
	expired := errors.Is(err, jwt.ErrTokenExpired)
	if rawToken == nil || (err != nil && !expired) {
		c.logger.Log("%s: bad jwt token %v", op, err)
		return &Error{
			Code: CodeTokenInvalid,
		}
	}
	claims, ok := rawToken.Claims.(*jwtClaims)
	if !ok {
		c.logger.Log("%s: bad jwt token claims", op)
		return &Error{
			Code: CodeTokenInvalid,
		}
	}
	if claims.TokenType != tokenTypeRefresh || claims.ExpiresAt == nil {
		c.logger.Log("%s: bad token claims %s", op, claims)
		return &Error{
			Code: CodeTokenInvalid,
		}
	}
	if expired {
		return &Error{
			Code: CodeTokenExpired,
		}
	}
	return nil
}

func (c *defaultService) ValidateAccessToken(token AccessToken) *Error {
	const op = "jwt.defaultService.ValidateAccessToken"
	rawToken, err := parseToken(string(token), []byte(c.accessTokenSecret))
	expired := errors.Is(err, jwt.ErrTokenExpired)
	if rawToken == nil || (err != nil && !expired) {
		c.logger.Log("%s: bad jwt token %v", op, err)
		return &Error{
			Code: CodeTokenInvalid,
		}
	}
	claims, ok := rawToken.Claims.(*jwtClaims)
	if !ok {
		c.logger.Log("%s: bad jwt token claims", op)
		return &Error{
			Code: CodeTokenInvalid,
		}
	}
	if claims.TokenType != tokenTypeAccess || claims.ExpiresAt == nil {
		c.logger.Log("%s: bad token claims %s", op, claims)
		return &Error{
			Code: CodeTokenInvalid,
		}
	}
	if expired {
		return &Error{
			Code: CodeTokenExpired,
		}
	}
	return nil
}

func (c *defaultService) GetRefreshTokenSubject(token RefreshToken) (Subject, *Error) {
	const op = "jwt.defaultService.GetRefreshTokenSubject"
	rawToken, err := parseToken(string(token), []byte(c.refreshTokenSecret))
	if rawToken == nil || err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			c.logger.Log("%s: jwt token expired %v", op, err)
			return "", &Error{
				Code: CodeTokenExpired,
			}
		} else {
			c.logger.Log("%s: bad jwt token %v", op, err)
			return "", &Error{
				Code: CodeTokenInvalid,
			}
		}
	}
	claims, ok := rawToken.Claims.(*jwtClaims)
	if !ok || claims.TokenType != tokenTypeRefresh {
		c.logger.Log("%s: bad jwt token claims", op)
		return "", &Error{
			Code: CodeTokenInvalid,
		}
	}
	return Subject(claims.Subject), nil
}

func (c *defaultService) GetAccessTokenSubject(token AccessToken) (Subject, *Error) {
	const op = "jwt.defaultService.GetAccessTokenSubject"
	rawToken, err := parseToken(string(token), []byte(c.accessTokenSecret))
	if rawToken == nil || err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			c.logger.Log("%s: jwt token expired %v", op, err)
			return "", &Error{
				Code: CodeTokenExpired,
			}
		} else {
			c.logger.Log("%s: bad jwt token %v", op, err)
			return "", &Error{
				Code: CodeTokenInvalid,
			}
		}
	}
	claims, ok := rawToken.Claims.(*jwtClaims)
	if !ok || claims.TokenType != tokenTypeAccess {
		c.logger.Log("%s: bad jwt token claims", op)
		return "", &Error{
			Code: CodeTokenInvalid,
		}
	}
	return Subject(claims.Subject), nil
}

type jwtClaims struct {
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}

func parseToken(signedToken string, secret []byte) (*jwt.Token, error) {
	return jwt.ParseWithClaims(signedToken, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
}

func generateToken(claims jwtClaims, secret []byte) (string, error) {
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
