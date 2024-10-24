package profile

import (
	"net/http"
	profileController "verni/internal/controllers/profile"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/middleware"
	"verni/internal/http-server/responses"

	"github.com/gin-gonic/gin"
)

type ProfileController profileController.Controller

func RegisterRoutes(
	router *gin.Engine,
	tokenChecker middleware.AccessTokenChecker,
	profile ProfileController,
) {
	methodGroup := router.Group("/profile", tokenChecker.Handler)
	methodGroup.GET("/getInfo", func(c *gin.Context) {
		info, err := profile.GetProfileInfo(profileController.UserId(tokenChecker.AccessToken(c)))
		if err != nil {
			switch err.Code {
			case profileController.GetInfoErrorNotFound:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeNoSuchRequest)
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
		aid, err := profile.UpdateAvatar(request.DataBase64, profileController.UserId(tokenChecker.AccessToken(c)))
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
		if err := profile.UpdateDisplayName(request.DisplayName, profileController.UserId(tokenChecker.AccessToken(c))); err != nil {
			switch err.Code {
			case profileController.UpdateDisplayNameErrorNotFound:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeNoSuchUser)
			case profileController.UpdateDisplayNameErrorWrongFormat:
				httpserver.Answer(c, err, http.StatusUnprocessableEntity, responses.CodeWrongFormat)
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.OK())
	})
}
