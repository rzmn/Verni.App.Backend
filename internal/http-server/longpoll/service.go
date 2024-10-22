package longpoll

import (
	"verni/internal/auth/jwt"
	authRepository "verni/internal/repositories/auth"

	"github.com/gin-gonic/gin"
)

type UserId string
type AuthRepository authRepository.Repository

type Service interface {
	CounterpartiesUpdated(uid UserId)
	SpendingsUpdated(uid UserId)
	FriendsUpdated(uid UserId)
	RegisterRoutes()
}

func DefaultService(e *gin.Engine, authRepository AuthRepository, jwtService jwt.Service) Service {
	return &defaultService{
		engine:         e,
		authRepository: authRepository,
		jwtService:     jwtService,
	}
}
