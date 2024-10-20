package storage

import (
	"context"
	"fmt"
	"log"
	"verni/internal/common"

	"github.com/ydb-platform/ydb-go-sdk/v3"
	yc "github.com/ydb-platform/ydb-go-yc"
)

type FriendStatus int

const (
	FriendStatusNo = iota
	FriendStatusIncomingRequest
	FriendStatusOutgoingRequest
	FriendStatusFriends
	FriendStatusMe
)

type UserId string
type DealId string
type AvatarId string

type AvatarData struct {
	Id         AvatarId `json:"id"`
	Base64Data *string  `json:"base64Data"`
}

type Avatar struct {
	Id *AvatarId `json:"id"`
}

type ProfileInfo struct {
	User          User   `json:"user"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
}

type User struct {
	Id           UserId       `json:"id"`
	DisplayName  string       `json:"displayName"`
	Avatar       Avatar       `json:"avatar"`
	FriendStatus FriendStatus `json:"friendStatus"`
}

type UserCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthenticatedSession struct {
	Id           UserId `json:"id"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type Deal struct {
	Timestamp int64      `json:"timestamp"`
	Details   string     `json:"details"`
	Cost      int64      `json:"cost"`
	Currency  string     `json:"currency"`
	Spendings []Spending `json:"spendings"`
}

type IdentifiableDeal struct {
	Deal
	Id DealId `json:"id"`
}

type Spending struct {
	UserId UserId `json:"userId"`
	Cost   int64  `json:"cost"`
}

type SpendingsPreview struct {
	Counterparty string           `json:"counterparty"`
	Balance      map[string]int64 `json:"balance"`
}

type Storage interface {
	GetAccountInfo(uid UserId) (*ProfileInfo, error)

	GetUserId(email string) (*UserId, error)
	StoreEmailValidationToken(email string, token string) error
	ExtractEmailValidationToken(email string) (*string, error)
	ValidateEmail(email string) error
	UpdateEmail(uid UserId, email string) error
	IsEmailExists(email string) (bool, error)

	IsUserExists(uid UserId) (bool, error)
	CheckCredentials(credentials UserCredentials) (bool, error)
	CheckPasswordForId(uid UserId, password string) (bool, error)
	UpdatePasswordForId(uid UserId, password string) error
	StoreCredentials(uid UserId, credentials UserCredentials) error

	StorePushToken(uid UserId, token string) error
	GetPushToken(uid UserId) (*string, error)

	StoreDisplayName(uid UserId, displayName string) error
	StoreAvatarBase64(uid UserId, avatarBase64 string) (AvatarId, error)
	GetAvatarsBase64(aids []AvatarId) (map[AvatarId]AvatarData, error)

	StoreRefreshToken(token string, uid UserId) error
	GetRefreshToken(uid UserId) (*string, error)
	RemoveRefreshToken(uid UserId) error

	GetFriends(uid UserId) ([]UserId, error)

	GetUsers(sender UserId, ids []UserId) ([]User, error)
	SearchUsers(sender UserId, query string) ([]User, error)

	Close()
}

type YDBConfig struct {
	Endpoint        string `json:"endpoint"`
	CredentialsPath string `json:"credentialsPath"`
}

func YDB(config YDBConfig) (Storage, error) {
	const op = "storage.ydb.New"
	ctx := context.Background()
	db, err := ydb.Open(ctx, config.Endpoint, yc.WithServiceAccountKeyFileCredentials(common.AbsolutePath(config.CredentialsPath)))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	log.Printf("%s: connection created", op)
	return &ydbStorage{
		Db:  db,
		Ctx: ctx,
	}, nil
}
