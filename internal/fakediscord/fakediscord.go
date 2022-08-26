package fakediscord

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Run() error {
	router := gin.Default()

	router.GET("/api/v9/gateway", getGateway)
	router.GET("/ws/", handleWS)

	return router.Run("localhost:8080")
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
