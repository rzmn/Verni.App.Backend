package longpoll

import (
	"verni/internal/auth/jwt"
	"verni/internal/storage"

	"github.com/gin-gonic/gin"
)

type UserId storage.UserId

type Service interface {
	CounterpartiesUpdated(uid UserId)
	SpendingsUpdated(uid UserId)
	FriendsUpdated(uid UserId)
	RegisterRoutes()
}

func DefaultService(e *gin.Engine, db storage.Storage, jwtService jwt.Service) Service {
	return &defaultService{
		engine:     e,
		db:         db,
		jwtService: jwtService,
	}
}
