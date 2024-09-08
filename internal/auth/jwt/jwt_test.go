package jwt

import (
	"testing"
	"time"
)

func TestOkAccessToken(t *testing.T) {
	uid := "uid"
	token, err := generateAccessToken(uid, time.Now())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	err = ValidateAccessToken(*token)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestBadAccessToken(t *testing.T) {
	err := ValidateAccessToken("blabla")
	if err != ErrBadToken {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestOutdatedAccessToken(t *testing.T) {
	uid := "uid"
	token, err := generateAccessToken(uid, time.Now().Add(-(accessTokenLifetime + time.Hour)))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	err = ValidateAccessToken(*token)
	if err != ErrTokenExpired {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestPassRefreshTokenInsteadOfAccess(t *testing.T) {
	uid := "uid"
	token, err := generateRefreshToken(uid, time.Now())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	err = ValidateAccessToken(*token)
	if err != ErrBadToken {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestIssueTokens(t *testing.T) {
	uid := "uid"
	tokens, err := IssueTokens(uid)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := ValidateAccessToken(tokens.Access); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := ValidateAccessToken(tokens.Refresh); err != ErrBadToken {
		t.Fatalf("unexpected err: %v", err)
	}
	if uid != tokens.Subject {
		t.Fatalf("wrong subject in issued token: %v", err)
	}
}

func TestRefreshTokenWithOkRefreshToken(t *testing.T) {
	uid := "uid"
	token, err := generateRefreshToken(uid, time.Now())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	tokens, err := IssueTokensBasedOnRefreshToken(*token)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := ValidateAccessToken(tokens.Access); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := ValidateAccessToken(tokens.Refresh); err != ErrBadToken {
		t.Fatalf("unexpected err: %v", err)
	}
	if uid != tokens.Subject {
		t.Fatalf("wrong subject in issued token: %v", err)
	}
}

func TestRefreshTokenWithOutdatedAccessToken(t *testing.T) {
	uid := "uid"
	token, err := generateRefreshToken(uid, time.Now().Add(-(refreshTokenLifetime + time.Hour)))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	_, err = IssueTokensBasedOnRefreshToken(*token)
	if err != ErrTokenExpired {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestRefreshTokenWithBadAccessToken(t *testing.T) {
	_, err := IssueTokensBasedOnRefreshToken("bla bla")
	if err != ErrBadToken {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestRefreshTokenWithAccessTokenInsteadOfAccess(t *testing.T) {
	uid := "uid"
	token, err := generateAccessToken(uid, time.Now())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	_, err = IssueTokensBasedOnRefreshToken(*token)
	if err != ErrBadToken {
		t.Fatalf("unexpected err: %v", err)
	}
}
