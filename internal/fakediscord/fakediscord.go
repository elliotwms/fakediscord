package fakediscord

import (
	"github.com/elliotwms/fake-discord/internal/fakediscord/api"
	"github.com/elliotwms/fake-discord/internal/fakediscord/storage"
	"github.com/elliotwms/fake-discord/internal/snowflake"
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

	api.GatewayController(router.Group("api/v9/gateway"))
	api.WebsocketController(router.Group("ws"))

	return router.Run("localhost:8080")
}
