package main

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/yaml.v2"

	"github.com/elliotwms/fakediscord/internal/fakediscord"
	"github.com/elliotwms/fakediscord/pkg/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	c, err := readConfig(getEnv("CONFIG_PATH", "config.yml"))
	if err != nil {
		slog.Error("Could not read config", "err", err)
		os.Exit(1)
	}

	if err := fakediscord.Run(ctx, ":"+getEnv("PORT", "8080"), c); err != nil {
		slog.Error("fakediscord stopped", "err", err)
		os.Exit(1)
	}
}

// readConfig reads the optional config file at path. A missing file results in an empty config
func readConfig(path string) (config.Config, error) {
	var c config.Config

	bs, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		slog.Info("No config file found, starting without config", "path", path)
		return c, nil
	}
	if err != nil {
		return c, err
	}

	if err := yaml.Unmarshal(bs, &c); err != nil {
		return c, err
	}

	return c, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
