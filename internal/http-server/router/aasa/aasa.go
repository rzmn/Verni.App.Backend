package aasa

import (
	"net/http"
	"verni/internal/storage"

	"github.com/gin-gonic/gin"
)

type Config struct {
	WebCredentials WebCredentials `json:"webcredentials"`
	AppClips       AppClips       `json:"appclips"`
	AppLinks       AppLinks       `json:"applinks"`
}

type WebCredentials struct {
	Apps []string `json:"apps"`
}

type AppClips struct{}

type AppLinks struct {
	Details []AppLinksDetails `json:"details"`
}

type AppLinksDetails struct {
	AppId string   `json:"appID"`
	Paths []string `json:"paths"`
}

const AppIdentifier = "NPZKGHFT2A.com.rzmn.accountydev.app"

func RegisterRoutes(e *gin.Engine, storage storage.Storage) {
	var config = Config{
		WebCredentials: WebCredentials{
			Apps: []string{AppIdentifier},
		},
		AppClips: AppClips{},
		AppLinks: AppLinks{
			Details: []AppLinksDetails{
				{
					AppId: AppIdentifier,
					Paths: []string{"*"},
				},
			},
		},
	}
	e.GET("/.well-known/apple-app-site-association", func(c *gin.Context) {
		c.JSON(http.StatusOK, config)
	})
	e.GET("/apple-app-site-association", func(c *gin.Context) {
		c.JSON(http.StatusOK, config)
	})
}
