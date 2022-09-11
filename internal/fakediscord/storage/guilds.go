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
		g := buildTestGuild(c)
		Guilds.Store(g.ID, g)
	}

	return nil
}

func buildTestGuild(g config.Guild) discordgo.Guild {
	guild := discordgo.Guild{
		ID:   snowflake.Generate().String(),
		Name: g.Name,
	}

	if g.ID != nil {
		guild.ID = g.ID.String()
	}

	for _, cc := range g.Channels {
		guild.Channels = append(guild.Channels, buildTestChannel(cc))
	}

	return guild
}

func buildTestChannel(c config.Channel) *discordgo.Channel {
	channel := &discordgo.Channel{
		ID:   snowflake.Generate().String(),
		Name: c.Name,
	}

	if c.ID != nil {
		channel.ID = c.ID.String()
	}

	return channel
}
