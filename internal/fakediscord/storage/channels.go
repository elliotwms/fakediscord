package storage

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/elliotwms/fakediscord/pkg/config"
)

var Channels sync.Map

func BuildTestChannel(c config.Channel) *discordgo.Channel {
	channel := discordgo.Channel{
		ID:   snowflake.Generate().String(),
		Name: c.Name,
	}

	if c.ID != nil {
		channel.ID = c.ID.String()
	}

	Channels.Store(channel.ID, channel)

	return &channel
}
