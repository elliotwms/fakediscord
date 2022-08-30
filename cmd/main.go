package main

import (
	"github.com/elliotwms/fake-discord/internal/fakediscord"
	"github.com/elliotwms/fake-discord/pkg/config"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

func main() {
	if err := fakediscord.Run(readConfig()); err != nil {
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
