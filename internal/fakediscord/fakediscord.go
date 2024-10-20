package fakediscord

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/elliotwms/fakediscord/internal/fakediscord/api"
	"github.com/elliotwms/fakediscord/internal/fakediscord/builders"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/elliotwms/fakediscord/pkg/config"
	"github.com/gin-gonic/gin"
)

// Version describes the build version
// it should be set via ldflags when building (see Dockerfile)
var Version = "v0.0.0+unknown"

func Run(ctx context.Context, c config.Config) error {
	// initiate the single-node snowflake ID generator
	if err := snowflake.Configure(0); err != nil {
		return err
	}

	slog.Info("Starting fakediscord", "version", Version)

	generate(c)

	return serve(ctx)
}

// generate generates resources based on the config provided, such as setting up users and guilds from a provided
// YAML file
func generate(c config.Config) {
	for _, user := range c.Users {
		u := builders.NewUserFromConfig(user).Build()
		slog.Info("Creating test user", "username", u.Username, "id", u.ID, "bot", u.Bot)

		storage.Users.Store(u.ID, *u)
	}

	for _, guild := range c.Guilds {
		g := builders.NewGuildFromConfig(guild).Build()

		slog.Info("Creating test guild", "name", g.Name, "id", g.ID)

		storage.Guilds.Store(g.ID, *g)
		for _, channel := range g.Channels {
			slog.Info("Creating test channel", "name", channel.Name, "id", channel.ID)
			storage.Channels.Store(channel.ID, *channel)
		}
	}
}

func serve(ctx context.Context) error {
	router := gin.Default()

	// register a shim to override the websocket
	api.WebsocketController(router.Group("ws"))

	// mock the HTTP api
	api.Configure(router.Group("api/:version"))

	s := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listen", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down server...")
	return s.Shutdown(ctx)
}
