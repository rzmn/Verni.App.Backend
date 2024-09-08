package jwt

import (
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrBadToken     = errors.New("bad token format")
	ErrTokenExpired = errors.New("token expired")
	ErrInternal     = errors.New("internal error")
)

var (
	tokenTypeRefresh = "refresh"
	tokenTypeAccess  = "access"
)

type jwtClaims struct {
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}

type Tokens struct {
	Refresh string
	Access  string
	Subject string
}

func IssueTokens(subject string) (*Tokens, error) {
	const op = "auth.jwt.IssueTokens"

	access, err := generateAccessToken(subject, time.Now())
	if err != nil {
		log.Printf("%s: cannot generate token (access) %v", op, err)
		return nil, ErrInternal
	}

	refresh, err := generateRefreshToken(subject, time.Now())
	if err != nil {
		log.Printf("%s: cannot generate token (refresh) %v", op, err)
		return nil, ErrInternal
	}

	return &Tokens{
		Access:  *access,
		Refresh: *refresh,
		Subject: subject,
	}, nil
}

func IssueTokensBasedOnRefreshToken(signedRefreshToken string) (*Tokens, error) {
	const op = "auth.jwt.IssueTokensBasedOnRefreshToken"

	token, err := parseToken(signedRefreshToken, refreshTokenSecret)
	expired := errors.Is(err, jwt.ErrTokenExpired)
	if token == nil || (err != nil && !expired) {
		log.Printf("%s: bad jwt token %v", op, err)
		return nil, ErrBadToken
	}

	if expired {
		log.Printf("%s: token expired", op)
		return nil, ErrTokenExpired
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		log.Printf("%s: bad jwt token claims", op)
		return nil, ErrBadToken
	}

	if claims.TokenType != tokenTypeRefresh || claims.ExpiresAt == nil {
		log.Printf("%s: bad token claims %s", op, claims)
		return nil, ErrBadToken
	}

	return IssueTokens(claims.Subject)
}

func ValidateAccessToken(signedToken string) error {
	const op = "auth.jwt.ValidateAccessToken"

	token, err := parseToken(signedToken, accessTokenSecret)
	expired := errors.Is(err, jwt.ErrTokenExpired)
	if token == nil || (err != nil && !expired) {
		log.Printf("%s: bad jwt token %v", op, err)
		return ErrBadToken
	}
	if expired {
		return ErrTokenExpired
	}
	return nil
}

func GetAccessTokenSubject(signedToken string) (*string, error) {
	const op = "auth.jwt.ValidateAccessToken"

	token, err := parseToken(signedToken, accessTokenSecret)
	if err != nil {
		log.Printf("%s: bad jwt token %v", op, err)
		return nil, ErrBadToken
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		log.Printf("%s: bad jwt token claims", op)
		return nil, ErrBadToken
	}
	return &claims.Subject, nil
}

func parseToken(signedToken string, secret []byte) (*jwt.Token, error) {
	return jwt.ParseWithClaims(signedToken, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
}

var (
	refreshTokenLifetime = time.Hour * 24 * 30
	accessTokenLifetime  = time.Hour
)

var accessTokenSecret = []byte("accessTokenSecret")

func generateAccessToken(id string, currentTime time.Time) (*string, error) {
	return generateToken(jwtClaims{
		TokenType: tokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   id,
			ExpiresAt: jwt.NewNumericDate(currentTime.Add(accessTokenLifetime)),
			IssuedAt:  jwt.NewNumericDate(currentTime),
		},
	}, accessTokenSecret)
}

var refreshTokenSecret = []byte("refreshTokenSecret")

func generateRefreshToken(id string, currentTime time.Time) (*string, error) {
	return generateToken(jwtClaims{
		TokenType: tokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   id,
			ExpiresAt: jwt.NewNumericDate(currentTime.Add(refreshTokenLifetime)),
			IssuedAt:  jwt.NewNumericDate(currentTime),
		},
	}, refreshTokenSecret)
}

func generateToken(claims jwtClaims, secret []byte) (*string, error) {
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		return nil, err
	}
	return &tokenString, nil
}
