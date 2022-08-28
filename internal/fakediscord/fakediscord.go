package fakediscord

import (
	"net/http"

	"github.com/bwmarrin/snowflake"
	"github.com/elliotwms/fake-discord/pkg/config"
	"github.com/gin-gonic/gin"
)

var node *snowflake.Node

func Run(c config.Config) error {
	if err := setupNode(0); err != nil { // todo set node ID
		return err
	}

	if err := importConfig(c); err != nil {
		return err
	}

	return setupRouter()
}

func setupNode(i int64) (err error) {
	node, err = snowflake.NewNode(i)

	return
}

func importConfig(c config.Config) error {
	return buildTestGuilds(c.Guilds)
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
