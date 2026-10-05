package fakediscord

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
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

// Run starts fakediscord listening on addr (e.g. ":8080"), bootstrapped with the resources in c. It blocks until ctx is
// cancelled or the server fails
func Run(ctx context.Context, addr string, c config.Config) error {
	// initiate the single-node snowflake ID generator
	if err := snowflake.Configure(0); err != nil {
		return err
	}

	slog.Info("Starting fakediscord", "version", Version)

	if err := generate(c); err != nil {
		return err
	}

	return serve(ctx, addr)
}

// generate generates resources based on the config provided, such as setting up users and guilds from a provided
// YAML file
func generate(c config.Config) error {
	users := []*discordgo.User{}
	for _, user := range c.Users {
		u := builders.NewUserFromConfig(user).Build()
		slog.Info("Creating test user", "username", u.Username, "id", u.ID, "bot", u.Bot)

		storage.Users.Store(u.ID, *u)
		users = append(users, u)
	}

	for _, guild := range c.Guilds {
		g := builders.
			NewGuildFromConfig(guild).
			WithUsers(users).
			Build()

		slog.Info("Creating test guild", "name", g.Name, "id", g.ID)

		if err := storage.State.GuildAdd(g); err != nil {
			return fmt.Errorf("add guild %q: %w", g.Name, err)
		}
	}

	return nil
}

func serve(ctx context.Context, addr string) error {
	router := gin.Default()

	// register a shim to override the websocket
	api.WebsocketController(router.Group("ws"))

	// mock the HTTP api
	api.Configure(router.Group("api/:version"))

	s := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errs := make(chan error, 1)
	go func() {
		slog.Info("Listening", "addr", addr)
		errs <- s.ListenAndServe()
	}()

	select {
	case err := <-errs:
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
	}

	slog.Info("Shutting down server...")

	// Create a new context for shutdown with a grace period
	// since the original context is already cancelled
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.Shutdown(shutdownCtx)
}
