package fakediscord

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/snowflake"
	"github.com/elliotwms/fake-discord/pkg/config"
)

var guilds sync.Map

func buildTestGuilds(cgs []config.Guild) error {
	for _, c := range cgs {
		g := buildTestGuild(c)
		id, err := snowflake.ParseString(g.ID)
		if err != nil {
			return err
		}
		guilds.Store(id, g)
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
		g.ID = node.Generate().String()
	}

	for _, cc := range c.Channels {
		channel := &discordgo.Channel{
			ID:   node.Generate().String(),
			Name: cc,
		}

		g.Channels = append(g.Channels, channel)
	}

	return g
}
