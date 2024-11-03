package avatars

import (
	"net/http"
	"verni/internal/common"
	avatarsController "verni/internal/controllers/avatars"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/responses"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, controller avatarsController.Controller) {
	router.GET("/avatars/get", func(c *gin.Context) {
		type GetAvatarsRequest struct {
			Ids []httpserver.ImageId `json:"ids"`
		}
		var request GetAvatarsRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		info, err := controller.GetAvatars(common.Map(request.Ids, func(id httpserver.ImageId) avatarsController.AvatarId {
			return avatarsController.AvatarId(id)
		}))
		if err != nil {
			switch err.Code {
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		response := map[httpserver.ImageId]httpserver.Image{}
		for _, avatar := range info {
			response[httpserver.ImageId(avatar.Id)] = httpserver.Image{
				Id:         httpserver.ImageId(avatar.Id),
				Base64Data: avatar.Base64,
			}
		}
		c.JSON(http.StatusOK, responses.Success(response))
	})
}
