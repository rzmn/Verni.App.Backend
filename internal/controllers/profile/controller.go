package profile

import (
	"verni/internal/common"
)

type UserId string
type AvatarId string

type ProfileInfo struct {
	Id            UserId
	DisplayName   string
	AvatarId      *AvatarId
	Email         string
	EmailVerified bool
}

type Controller interface {
	GetProfileInfo(id UserId) (ProfileInfo, *common.CodeBasedError[GetInfoErrorCode])
	UpdateDisplayName(name string, id UserId) *common.CodeBasedError[UpdateDisplayNameErrorCode]
	UpdateAvatar(base64 string, id UserId) (AvatarId, *common.CodeBasedError[UpdateAvatarErrorCode])
}
