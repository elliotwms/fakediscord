package builders

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/snowflake"
)

type Message struct {
	m *discordgo.Message
}

func NewMessage(author *discordgo.User, channelD, guildID string) *Message {
	return &Message{
		m: &discordgo.Message{
			ID:        snowflake.Generate().String(),
			ChannelID: channelD,
			GuildID:   guildID,
			Author:    author,
			Timestamp: time.Now(),
		},
	}
}

func (b *Message) Build() *discordgo.Message {
	return b.m
}

func (b *Message) WithID(id string) *Message {
	b.m.ID = id

	return b
}

func (b *Message) WithContent(s string) *Message {
	b.m.Content = s

	return b
}

func (b *Message) WithEmbeds(embeds []*discordgo.MessageEmbed) *Message {
	b.m.Embeds = embeds

	return b
}

func (b *Message) WithAttachments(attachments []*discordgo.MessageAttachment) *Message {
	b.m.Attachments = attachments

	return b
}

func (b *Message) WithComponents(components []discordgo.MessageComponent) *Message {
	b.m.Components = components

	return b
}

func (b *Message) WithType(t discordgo.MessageType) *Message {
	b.m.Type = t

	return b
}

func (b *Message) WithFlags(flags discordgo.MessageFlags) *Message {
	b.m.Flags = flags

	return b
}
