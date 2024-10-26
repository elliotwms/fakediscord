package builders

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/elliotwms/fakediscord/pkg/config"
)

type Guild struct {
	g *discordgo.Guild
}

func NewGuild(name string) *Guild {
	return &Guild{g: &discordgo.Guild{
		ID:   snowflake.Generate().String(),
		Name: name,
	}}
}

func NewGuildFromConfig(config config.Guild) *Guild {
	guild := NewGuild(config.Name)

	if config.ID != nil {
		guild.WithID(config.ID.String())
	}

	for _, channel := range config.Channels {
		guild.WithChannel(
			NewChannelFromConfig(channel).Build(),
		)
	}

	return guild
}

func (g *Guild) Build() *discordgo.Guild {
	return g.g
}

func (g *Guild) WithID(id string) *Guild {
	g.g.ID = id

	return g
}

func (g *Guild) WithChannel(channel *discordgo.Channel) *Guild {
	g.g.Channels = append(g.g.Channels, channel)

	return g
}

func (g *Guild) WithUsers(users []*discordgo.User) *Guild {
	for _, user := range users {
		g.g.Members = append(g.g.Members, NewMember(g.g.ID, user).Build())
	}
	return g
}

type Member struct {
	m *discordgo.Member
}

func NewMember(guildID string, u *discordgo.User) *Member {
	return &Member{
		m: &discordgo.Member{
			GuildID:  guildID,
			JoinedAt: time.Now(),
			User:     u,
		},
	}
}

func (m *Member) Build() *discordgo.Member {
	return m.m
}
