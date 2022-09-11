package tests

import (
	"embed"
	"gopkg.in/yaml.v2"
	"testing"

	"github.com/elliotwms/fakediscord/internal/fakediscord"
	"github.com/elliotwms/fakediscord/pkg/config"
	pkgfakediscord "github.com/elliotwms/fakediscord/pkg/fakediscord"
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
