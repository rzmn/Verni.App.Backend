package profile

import (
	"net/http"
	"verni/internal/auth/jwt"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/middleware"
	"verni/internal/http-server/responses"
	profileController "verni/internal/http-server/router/profile"
	"verni/internal/storage"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, db storage.Storage, jwtService jwt.Service) {
	ensureLoggedIn := middleware.EnsureLoggedIn(db, jwtService)
	hostFromToken := func(c *gin.Context) profileController.UserId {
		return profileController.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	}
	controller := profileController.DefaultController(db)
	methodGroup := router.Group("/profile", ensureLoggedIn)
	methodGroup.GET("/getInfo", func(c *gin.Context) {
		info, err := controller.GetProfileInfo(hostFromToken(c))
		if err != nil {
			switch err.Code {
			case profileController.GetInfoErrorNotFound:
				c.JSON(
					http.StatusConflict,
					responses.Failure(
						responses.Error{
							Code: responses.CodeNoSuchUser,
						},
					),
				)
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.Success(info))
	})
	methodGroup.PUT("/setAvatar", func(c *gin.Context) {
		type UpdateAvatarRequest struct {
			DataBase64 string `json:"dataBase64"`
		}
		var request UpdateAvatarRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		aid, err := controller.UpdateAvatar(request.DataBase64, hostFromToken(c))
		if err != nil {
			switch err.Code {
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.Success(aid))
	})
	methodGroup.PUT("/setDisplayName", func(c *gin.Context) {
		type UpdateDisplayNameRequest struct {
			DisplayName string `json:"displayName"`
		}
		var request UpdateDisplayNameRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		if err := controller.UpdateDisplayName(request.DisplayName, hostFromToken(c)); err != nil {
			switch err.Code {
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.OK())
	})
}
