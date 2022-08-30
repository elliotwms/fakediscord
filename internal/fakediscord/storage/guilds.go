package storage

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fake-discord/internal/snowflake"
	"github.com/elliotwms/fake-discord/pkg/config"
)

var Guilds sync.Map

func BuildTestGuilds(cgs []config.Guild) error {
	for _, c := range cgs {
		g := buildTestGuild(c)
		Guilds.Store(g.ID, g)
	}

	return nil
}

func buildTestGuild(c config.Guild) discordgo.Guild {
	g := discordgo.Guild{
		Name: c.Name,
	}

	if c.ID != nil {
		g.ID = c.ID.String()
	} else {
		g.ID = snowflake.Generate().String()
	}

	for _, cc := range c.Channels {
		channel := &discordgo.Channel{
			ID:   snowflake.Generate().String(),
			Name: cc,
		}

		g.Channels = append(g.Channels, channel)
	}

	return g
}
