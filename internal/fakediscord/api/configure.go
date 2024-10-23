package api

import "github.com/gin-gonic/gin"

var controllers = map[string]func(r *gin.RouterGroup){
	"applications": applicationsController,
	"channels":     channelController,
	"gateway":      gatewayController,
	"guilds":       guildsController,
	"users":        usersController,
	"interactions": interactionsController,
	"webhooks":     webhooksController,
}

func Configure(api *gin.RouterGroup) {
	for path, group := range controllers {
		group(api.Group(path))
	}
}
