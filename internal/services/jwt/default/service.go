package defaultJwtService

import (
	"errors"
	"time"

	jwtService "github.com/rzmn/Verni.App.Backend/internal/services/jwt"
	"github.com/rzmn/Verni.App.Backend/internal/services/logging"

	"github.com/golang-jwt/jwt/v5"
)

type DefaultConfig struct {
	AccessTokenLifetimeHours  int    `json:"accessTokenLifetimeHours"`
	RefreshTokenLifetimeHours int    `json:"refreshTokenLifetimeHours"`
	RefreshTokenSecret        string `json:"refreshTokenSecret"`
	AccessTokenSecret         string `json:"accessTokenSecret"`
}

func New(
	config DefaultConfig,
	logger logging.Service,
	currentTime func() time.Time,
) jwtService.Service {
	return &defaultService{
		refreshTokenLifetime: time.Hour * time.Duration(config.RefreshTokenLifetimeHours),
		accessTokenLifetime:  time.Hour * time.Duration(config.AccessTokenLifetimeHours),
		refreshTokenSecret:   config.RefreshTokenSecret,
		accessTokenSecret:    config.AccessTokenSecret,
		currentTime:          currentTime,
		logger:               logger,
	}
}

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

func (c *defaultService) IssueRefreshToken(subject jwtService.Subject) (jwtService.RefreshToken, *jwtService.Error) {
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
		c.logger.LogInfo("%s: cannot generate token %v", op, err)
		return jwtService.RefreshToken(rawToken), &jwtService.Error{
			Code: jwtService.CodeInternal,
		}
	}
	return jwtService.RefreshToken(rawToken), nil
}

func (c *defaultService) IssueAccessToken(subject jwtService.Subject) (jwtService.AccessToken, *jwtService.Error) {
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
		c.logger.LogInfo("%s: cannot generate token %v", op, err)
		return jwtService.AccessToken(rawToken), &jwtService.Error{
			Code: jwtService.CodeInternal,
		}
	}
	return jwtService.AccessToken(rawToken), nil
}

func (c *defaultService) ValidateRefreshToken(token jwtService.RefreshToken) *jwtService.Error {
	const op = "jwt.defaultService.ValidateRefreshToken"
	rawToken, err := parseToken(string(token), []byte(c.refreshTokenSecret))
	expired := errors.Is(err, jwt.ErrTokenExpired)
	if rawToken == nil || (err != nil && !expired) {
		c.logger.LogInfo("%s: bad jwt token %v", op, err)
		return &jwtService.Error{
			Code: jwtService.CodeTokenInvalid,
		}
	}
	claims, ok := rawToken.Claims.(*jwtClaims)
	if !ok {
		c.logger.LogInfo("%s: bad jwt token claims", op)
		return &jwtService.Error{
			Code: jwtService.CodeTokenInvalid,
		}
	}
	if claims.TokenType != tokenTypeRefresh || claims.ExpiresAt == nil {
		c.logger.LogInfo("%s: bad token claims %s", op, claims)
		return &jwtService.Error{
			Code: jwtService.CodeTokenInvalid,
		}
	}
	if expired {
		return &jwtService.Error{
			Code: jwtService.CodeTokenExpired,
		}
	}
	return nil
}

func (c *defaultService) ValidateAccessToken(token jwtService.AccessToken) *jwtService.Error {
	const op = "jwt.defaultService.ValidateAccessToken"
	rawToken, err := parseToken(string(token), []byte(c.accessTokenSecret))
	expired := errors.Is(err, jwt.ErrTokenExpired)
	if rawToken == nil || (err != nil && !expired) {
		c.logger.LogInfo("%s: bad jwt token %v", op, err)
		return &jwtService.Error{
			Code: jwtService.CodeTokenInvalid,
		}
	}
	claims, ok := rawToken.Claims.(*jwtClaims)
	if !ok {
		c.logger.LogInfo("%s: bad jwt token claims", op)
		return &jwtService.Error{
			Code: jwtService.CodeTokenInvalid,
		}
	}
	if claims.TokenType != tokenTypeAccess || claims.ExpiresAt == nil {
		c.logger.LogInfo("%s: bad token claims %s", op, claims)
		return &jwtService.Error{
			Code: jwtService.CodeTokenInvalid,
		}
	}
	if expired {
		return &jwtService.Error{
			Code: jwtService.CodeTokenExpired,
		}
	}
	return nil
}

func (c *defaultService) GetRefreshTokenSubject(token jwtService.RefreshToken) (jwtService.Subject, *jwtService.Error) {
	const op = "jwt.defaultService.GetRefreshTokenSubject"
	rawToken, err := parseToken(string(token), []byte(c.refreshTokenSecret))
	if rawToken == nil || err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			c.logger.LogInfo("%s: jwt token expired %v", op, err)
			return "", &jwtService.Error{
				Code: jwtService.CodeTokenExpired,
			}
		} else {
			c.logger.LogInfo("%s: bad jwt token %v", op, err)
			return "", &jwtService.Error{
				Code: jwtService.CodeTokenInvalid,
			}
		}
	}
	claims, ok := rawToken.Claims.(*jwtClaims)
	if !ok || claims.TokenType != tokenTypeRefresh {
		c.logger.LogInfo("%s: bad jwt token claims", op)
		return "", &jwtService.Error{
			Code: jwtService.CodeTokenInvalid,
		}
	}
	return jwtService.Subject(claims.Subject), nil
}

func (c *defaultService) GetAccessTokenSubject(token jwtService.AccessToken) (jwtService.Subject, *jwtService.Error) {
	const op = "jwt.defaultService.GetAccessTokenSubject"
	rawToken, err := parseToken(string(token), []byte(c.accessTokenSecret))
	if rawToken == nil || err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			c.logger.LogInfo("%s: jwt token expired %v", op, err)
			return "", &jwtService.Error{
				Code: jwtService.CodeTokenExpired,
			}
		} else {
			c.logger.LogInfo("%s: bad jwt token %v", op, err)
			return "", &jwtService.Error{
				Code: jwtService.CodeTokenInvalid,
			}
		}
	}
	claims, ok := rawToken.Claims.(*jwtClaims)
	if !ok || claims.TokenType != tokenTypeAccess {
		c.logger.LogInfo("%s: bad jwt token claims", op)
		return "", &jwtService.Error{
			Code: jwtService.CodeTokenInvalid,
		}
	}
	return jwtService.Subject(claims.Subject), nil
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
