package api

import "github.com/gin-gonic/gin"

var controllers = map[string]func(r *gin.RouterGroup){
	"gateway":  gatewayController,
	"channels": channelController,
}

func Configure(api *gin.RouterGroup) {
	for path, group := range controllers {
		group(api.Group(path))
	}
}