package tests

import (
	"context"
	"embed"
	"fmt"
	"os"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord"
	"github.com/elliotwms/fakediscord/pkg/config"
	pkgfakediscord "github.com/elliotwms/fakediscord/pkg/fakediscord"
	"gopkg.in/yaml.v2"
)

const botToken = "token"
const appID = "1290742494824366183"

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
		if err := fakediscord.Run(context.Background(), c); err != nil {
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

func newSession(token string) *discordgo.Session {
	session, _ := discordgo.New("Bot " + token)

	if os.Getenv("DEBUG") != "" {
		session.LogLevel = discordgo.LogDebug
		session.Debug = true
	}

	session.State.MaxMessageCount = 100

	return session
}

func newOpenSession(t *testing.T, token string) (session *discordgo.Session, closer func()) {
	session = newSession(token)

	err := session.Open()
	if err != nil {
		t.Fatal(err)
	}

	return session, func() {
		err = session.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func setupGuild(s *discordgo.Session, name string) (*discordgo.Guild, *discordgo.Channel, error) {
	guild, err := s.GuildCreate(fmt.Sprintf("%s_test", name))
	if err != nil {
		return nil, nil, err
	}

	channel, err := s.GuildChannelCreate(guild.ID, "test", discordgo.ChannelTypeGuildText)

	return guild, channel, err
}
