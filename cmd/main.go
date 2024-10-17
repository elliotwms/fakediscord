package main

import (
	"context"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/elliotwms/fakediscord/internal/fakediscord"
	"github.com/elliotwms/fakediscord/pkg/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	if err := fakediscord.Run(ctx, readConfig()); err != nil {
		panic(err)
	}
}

func readConfig() config.Config {
	bs, err := os.ReadFile("config.yml")
	if err != nil {
		log.Printf("could not read config.yml: %v", err)
		return config.Config{}
	}

	var c config.Config

	if err := yaml.Unmarshal(bs, &c); err != nil {
		log.Printf("could not read config.yml: %v", err)
		return config.Config{}
	}

	return c
}
