package storage

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/elliotwms/fakediscord/pkg/config"
)

var Guilds sync.Map

func BuildTestGuilds(cgs []config.Guild) error {
	for _, c := range cgs {
		BuildTestGuild(c)
	}

	return nil
}

func BuildTestGuild(g config.Guild) discordgo.Guild {
	guild := discordgo.Guild{
		ID:   snowflake.Generate().String(),
		Name: g.Name,
	}

	if g.ID != nil {
		guild.ID = g.ID.String()
	}

	for _, cc := range g.Channels {
		guild.Channels = append(guild.Channels, BuildTestChannel(cc))
	}

	Guilds.Store(guild.ID, guild)

	return guild
}
