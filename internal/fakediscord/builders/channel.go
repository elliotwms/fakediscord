package builders

import (
	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/elliotwms/fakediscord/pkg/config"
)

type Channel struct {
	c *discordgo.Channel
}

func NewChannel(name string) *Channel {
	return &Channel{
		c: &discordgo.Channel{
			ID:   snowflake.Generate().String(),
			Name: name,
		},
	}
}

func NewChannelFromConfig(config config.Channel) *Channel {
	channel := NewChannel(config.Name)

	if config.ID != nil {
		channel.WithID(config.ID.String())
	}

	return channel
}

func (c *Channel) Build() *discordgo.Channel {
	return c.c
}

func (c *Channel) WithID(id string) *Channel {
	c.c.ID = id

	return c
}
