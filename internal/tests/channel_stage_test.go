package tests

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

type ChannelStage struct {
	t       *testing.T
	require *require.Assertions
	session *discordgo.Session
	guild   *discordgo.Guild
	channel *discordgo.Channel
}

func NewChannelStage(t *testing.T) (*ChannelStage, *ChannelStage, *ChannelStage) {
	r := require.New(t)

	session := newSession(botToken)
	s := &ChannelStage{
		t:       t,
		require: r,
		session: session,
	}

	s.require.NoError(session.Open())
	t.Cleanup(func() { s.require.NoError(session.Close()) })

	var err error
	s.guild, _, err = setupGuild(session, "channel")
	s.require.NoError(err)

	return s, s, s
}

func (s *ChannelStage) and() *ChannelStage {
	return s
}

func (s *ChannelStage) a_channel_is_created_named(name string) *ChannelStage {
	var err error
	s.channel, err = s.session.GuildChannelCreate(s.guild.ID, name, discordgo.ChannelTypeGuildText)
	s.require.NoError(err)

	s.require.NotEmpty(s.channel.ID)
	s.require.Equal(s.channel.GuildID, s.guild.ID)
	s.require.Equal(name, s.channel.Name)
	s.require.Equal(s.channel.Type, discordgo.ChannelTypeGuildText)

	return s
}

func (s *ChannelStage) a_channel_does_not_exist_named(name string) *ChannelStage {
	res, err := s.session.GuildChannels(s.guild.ID)
	s.require.NoError(err)

	var found bool
	for _, channel := range res {
		if channel.Name == name {
			found = true
		}
	}

	s.require.False(found)

	return s
}

func (s *ChannelStage) state_contains_the_channel() *ChannelStage {
	s.require.Eventually(func() bool {
		channel, err := s.session.State.Channel(s.channel.ID)
		return !(err != nil || channel == nil)
	}, time.Second, time.Millisecond*100)

	return s
}

func (s *ChannelStage) state_does_not_contain_the_channel() *ChannelStage {
	s.require.Eventually(func() bool {
		channel, err := s.session.State.Channel(s.channel.ID)
		return err != nil || channel == nil
	}, time.Second, time.Millisecond*100)

	return s
}

func (s *ChannelStage) the_channel_is_deleted() {
	_, err := s.session.ChannelDelete(s.channel.ID)
	s.require.NoError(err)
}

func (s *ChannelStage) guild_has_channel() {
	s.require.Eventually(func() bool {
		channels, err := s.session.GuildChannels(s.guild.ID)
		if err != nil {
			return false
		}

		for _, c := range channels {
			if c.ID == s.channel.ID {
				return true
			}
		}

		return false
	}, time.Second, time.Millisecond*100)
}
