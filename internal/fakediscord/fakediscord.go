package fakediscord

import (
	"github.com/elliotwms/fakediscord/internal/fakediscord/api"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/elliotwms/fakediscord/pkg/config"
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

	// register a shim to override the websocket
	api.WebsocketController(router.Group("ws"))

	// mock the HTTP api
	v9 := router.Group("api/v9")
	api.Configure(v9)

	return router.Run(":8080")
}
