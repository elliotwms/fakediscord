package tests

import (
	"embed"
	"github.com/elliotwms/fake-discord/pkg/config"
	"gopkg.in/yaml.v2"
	"testing"

	"github.com/elliotwms/fake-discord/internal/fakediscord"
	pkgfakediscord "github.com/elliotwms/fake-discord/pkg/fakediscord"
)

//go:embed files/config.yml
var configDir embed.FS

func TestMain(m *testing.M) {
	setup()

	m.Run()
}

func setup() {
	pkgfakediscord.Configure("http://localhost:8080/")

	c := readConfig()

	go func() {
		if err := fakediscord.Run(c); err != nil {
			panic(err)
		}
	}()
}

func readConfig() config.Config {
	bs, err := configDir.ReadFile("files/config.yml")
	if err != nil {
		panic(err)
	}

	var c config.Config
	if err := yaml.Unmarshal(bs, &c); err != nil {
		panic(err)
	}

	return c
}
