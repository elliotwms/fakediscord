package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func GatewayController(r *gin.RouterGroup) {
	r.GET("/", getGateway)
}

// https://discord.com/developers/docs/topics/gateway#get-gateway
// overrides to provide a shim to the local websocket handler, handleWS
func getGateway(c *gin.Context) {
	c.JSON(http.StatusOK, struct {
		URL string `json:"url"`
	}{
		"ws://localhost:8080/ws",
	})
}
