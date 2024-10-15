package avatars

import (
	"accounty/internal/common"
	httpserver "accounty/internal/http-server"
	"accounty/internal/http-server/responses"
	avatarsController "accounty/internal/http-server/router/avatars"
	"accounty/internal/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, db storage.Storage) {
	controller := avatarsController.DefaultController(db)
	router.GET("/avatars/get", func(c *gin.Context) {
		type GetAvatarsRequest struct {
			Ids []storage.AvatarId `json:"ids"`
		}
		var request GetAvatarsRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		info, err := controller.GetAvatars(common.Map(request.Ids, func(id storage.AvatarId) avatarsController.AvatarId {
			return avatarsController.AvatarId(id)
		}))
		if err != nil {
			switch err.Code {
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.Success(info))
	})
}
