package fake_discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"net/http"
)

func OverrideEndpoints() {
	discordgo.EndpointDiscord = "http://localhost:8080/"
	discordgo.EndpointAPI = discordgo.EndpointDiscord + "api/v" + discordgo.APIVersion + "/"
	discordgo.EndpointGateway = discordgo.EndpointAPI + "gateway"
}

func ServeAPI() error {
	router := gin.Default()

	router.GET("/api/v9/gateway", getGateway)
	router.GET("/ws/", handleWS)

	return router.Run("localhost:8080")
}

func getGateway(c *gin.Context) {
	c.JSON(http.StatusOK, struct {
		URL string `json:"url"`
	}{
		"ws://localhost:8080/ws",
	})
}
