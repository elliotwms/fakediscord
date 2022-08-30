package fakediscord

import (
	"net/http"

	"github.com/elliotwms/fake-discord/internal/snowflake"
	"github.com/elliotwms/fake-discord/internal/storage"
	"github.com/elliotwms/fake-discord/pkg/config"
	"github.com/gin-gonic/gin"
)

func Run(c config.Config) error {
	if err := snowflake.Configure(0); err != nil { // todo set node ID
		return err
	}

	if err := importConfig(c); err != nil {
		return err
	}

	return setupRouter()
}

func importConfig(c config.Config) error {
	return storage.BuildTestGuilds(c.Guilds)
}

func setupRouter() error {
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
