package api

import (
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
)

func gatewayController(r *gin.RouterGroup) {
	r.Use(auth)
	r.GET("", getGateway)
	r.GET("/", getGateway)
	r.GET("/bot", getGatewayBot)
}

// https://discord.com/developers/docs/topics/gateway#get-gateway
// overrides to provide a shim to the local websocket handler, handleWS
func getGateway(c *gin.Context) {
	c.JSON(http.StatusOK, struct {
		URL string `json:"url"`
	}{
		gatewayURL(c),
	})
}

// https://discord.com/developers/docs/topics/gateway#get-gateway-bot
func getGatewayBot(c *gin.Context) {
	c.JSON(http.StatusOK, discordgo.GatewayBotResponse{
		URL:    gatewayURL(c),
		Shards: 1,
		SessionStartLimit: discordgo.SessionInformation{
			Total:          1000,
			Remaining:      1000,
			ResetAfter:     24 * 60 * 60 * 1000,
			MaxConcurrency: 1,
		},
	})
}

// gatewayURL builds the websocket URL from the host the client used to reach fakediscord, so that it works behind
// docker-compose service names and remapped ports rather than assuming localhost:8080
func gatewayURL(c *gin.Context) string {
	scheme := "ws"
	if c.Request.TLS != nil {
		scheme = "wss"
	}

	return scheme + "://" + c.Request.Host + "/ws"
}
