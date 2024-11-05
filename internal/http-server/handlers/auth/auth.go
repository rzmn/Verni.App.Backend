package auth

import (
	"net/http"
	authController "verni/internal/controllers/auth"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/middleware"
	"verni/internal/http-server/responses"
	"verni/internal/services/logging"

	"github.com/gin-gonic/gin"
)

type AuthController authController.Controller

func RegisterRoutes(
	router *gin.Engine,
	logger logging.Service,
	tokenChecker middleware.AccessTokenChecker,
	auth AuthController,
) {
	router.PUT("/auth/signup", func(c *gin.Context) {
		type SignupRequest struct {
			Credentials httpserver.Credentials `json:"credentials"`
		}
		var request SignupRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		session, err := auth.Signup(request.Credentials.Email, request.Credentials.Password)
		if err != nil {
			switch err.Code {
			case authController.SignupErrorAlreadyTaken:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeAlreadyTaken)
			case authController.SignupErrorWrongFormat:
				httpserver.Answer(c, err, http.StatusUnprocessableEntity, responses.CodeWrongFormat)
			default:
				logger.LogError("signup request %v failed with unknown err: %v", request, err)
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.Success(session))
	})
	router.PUT("/auth/login", func(c *gin.Context) {
		type LoginRequest struct {
			Credentials httpserver.Credentials `json:"credentials"`
		}
		var request LoginRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		session, err := auth.Login(request.Credentials.Email, request.Credentials.Password)
		if err != nil {
			switch err.Code {
			case authController.LoginErrorWrongCredentials:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeIncorrectCredentials)
			default:
				logger.LogError("login request %v failed with unknown err: %v", request, err)
				httpserver.AnswerWithUnknownError(c, err)
			}
		}
		c.JSON(http.StatusOK, responses.Success(session))
	})
	router.PUT("/auth/refresh", func(c *gin.Context) {
		type RefreshRequest struct {
			RefreshToken string `json:"refreshToken"`
		}
		var request RefreshRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		session, err := auth.Refresh(request.RefreshToken)
		if err != nil {
			switch err.Code {
			case authController.RefreshErrorTokenExpired:
				httpserver.Answer(c, err, http.StatusUnauthorized, responses.CodeTokenExpired)
			case authController.RefreshErrorTokenIsWrong:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeIncorrectCredentials)
			default:
				logger.LogError("refresh request %v failed with unknown err: %v", request, err)
				httpserver.AnswerWithUnknownError(c, err)
			}
		}
		c.JSON(http.StatusOK, responses.Success(session))
	})
	router.PUT("/auth/updateEmail", tokenChecker.Handler, func(c *gin.Context) {
		type UpdateEmailRequest struct {
			Email string `json:"email"`
		}
		var request UpdateEmailRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		session, err := auth.UpdateEmail(request.Email, authController.UserId(tokenChecker.AccessToken(c)))
		if err != nil {
			switch err.Code {
			case authController.UpdateEmailErrorAlreadyTaken:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeAlreadyTaken)
			case authController.UpdateEmailErrorWrongFormat:
				httpserver.Answer(c, err, http.StatusUnprocessableEntity, responses.CodeWrongFormat)
			default:
				logger.LogError("updateEmail request %v failed with unknown err: %v", request, err)
				httpserver.AnswerWithUnknownError(c, err)
			}
		}
		c.JSON(http.StatusOK, responses.Success(session))
	})
	router.PUT("/auth/updatePassword", tokenChecker.Handler, func(c *gin.Context) {
		type UpdatePasswordRequest struct {
			OldPassword string `json:"old"`
			NewPassword string `json:"new"`
		}
		var request UpdatePasswordRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		session, err := auth.UpdatePassword(request.OldPassword, request.NewPassword, authController.UserId(tokenChecker.AccessToken(c)))
		if err != nil {
			switch err.Code {
			case authController.UpdatePasswordErrorOldPasswordIsWrong:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeIncorrectCredentials)
			default:
				logger.LogError("updatePassword request %v failed with unknown err: %v", request, err)
				httpserver.AnswerWithUnknownError(c, err)
			}
		}
		c.JSON(http.StatusOK, responses.Success(session))
	})
	router.DELETE("/auth/logout", tokenChecker.Handler, func(c *gin.Context) {
		if err := auth.Logout(authController.UserId(tokenChecker.AccessToken(c))); err != nil {
			switch err.Code {
			default:
				logger.LogError("logout request failed with unknown err: %v", err)
				httpserver.AnswerWithUnknownError(c, err)
			}
		}
		c.JSON(http.StatusOK, responses.OK())
	})
	router.PUT("/auth/registerForPushNotifications", tokenChecker.Handler, func(c *gin.Context) {
		type RegisterForPushNotificationsRequest struct {
			Token string `json:"token"`
		}
		var request RegisterForPushNotificationsRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		if err := auth.RegisterForPushNotifications(request.Token, authController.UserId(tokenChecker.AccessToken(c))); err != nil {
			switch err.Code {
			default:
				logger.LogError("registerForPushNotifications request %v failed with unknown err: %v", request, err)
				httpserver.AnswerWithUnknownError(c, err)
			}
		}
		c.JSON(http.StatusOK, responses.OK())
	})
}
