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
	thread  *discordgo.Channel
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
	s.guild, _, err = setupGuild(t, session, "channel")
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
	return s.stateContainsChannel(s.channel.ID)
}

func (s *ChannelStage) state_does_not_contain_the_channel() *ChannelStage {
	s.require.Eventually(func() bool {
		channel, err := s.session.State.Channel(s.channel.ID)
		return err != nil || channel == nil
	}, time.Second, time.Millisecond*100)

	return s
}

func (s *ChannelStage) stateContainsChannel(id string) *ChannelStage {
	s.require.Eventually(func() bool {
		channel, err := s.session.State.Channel(id)
		return err == nil && channel != nil
	}, time.Second, time.Millisecond*100)

	return s
}

func (s *ChannelStage) the_channel_is_deleted() {
	_, err := s.session.ChannelDelete(s.channel.ID)
	s.require.NoError(err)
}

func (s *ChannelStage) get_guild_channels_contains_channel() *ChannelStage {
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

	return s
}

func (s *ChannelStage) a_thread_is_created_named(name string) *ChannelStage {
	var err error
	s.thread, err = s.session.ThreadStart(s.channel.ID, name, discordgo.ChannelTypeGuildPublicThread, 0)
	s.require.NoError(err)

	return s
}

func (s *ChannelStage) state_contains_the_thread() *ChannelStage {
	s.stateContainsChannel(s.thread.ID)

	return s
}

func (s *ChannelStage) get_guild_channels_does_not_contain_thread() *ChannelStage {
	channels, err := s.session.GuildChannels(s.guild.ID)
	s.require.NoError(err)

	for _, channel := range channels {
		if channel.ID == s.thread.ID {
			s.t.Fatal("guild channels returned thread")
		}
	}

	return s
}

func (s *ChannelStage) the_thread_should_have_default_auto_archive_time() {
	s.require.Equal(3*24*60, s.thread.ThreadMetadata.AutoArchiveDuration)
}
