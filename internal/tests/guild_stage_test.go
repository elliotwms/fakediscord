package tests

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

type GuildStage struct {
	t       *testing.T
	require *require.Assertions
	session *discordgo.Session

	guildName string

	err         error
	guild       *discordgo.Guild
	guildCreate *discordgo.GuildCreate
	guildDelete *discordgo.GuildDelete
}

func NewGuildStage(t *testing.T) (*GuildStage, *GuildStage, *GuildStage) {
	session, closer := newOpenSession(t, botToken)
	t.Cleanup(closer)

	s := &GuildStage{
		t:       t,
		require: require.New(t),
		session: session,
	}

	return s, s, s
}

func (s *GuildStage) and() *GuildStage {
	return s
}

func (s *GuildStage) a_guild_named(name string) *GuildStage {
	s.guildName = name

	return s
}

func (s *GuildStage) the_session_expects_a_guild_create_event_for_the_guild() {
	s.session.AddHandler(func(_ *discordgo.Session, e *discordgo.GuildCreate) {
		s.t.Logf("Received %s event for guild '%s'", "GUILD_CREATE", e.Guild.Name)
		if e.Guild.Name == s.guildName {
			s.guildCreate = e
		}
	})
}

func (s *GuildStage) the_guild_is_created() *GuildStage {
	s.guild, s.err = s.session.GuildCreate(s.guildName)

	return s
}

func (s *GuildStage) no_error_should_be_returned() *GuildStage {
	s.require.NoError(s.err)

	return s
}

func (s *GuildStage) the_session_should_have_received_the_guild_create_event() {
	s.require.Eventually(func() bool {
		return s.guildCreate != nil
	}, time.Second, time.Millisecond*10)
}

func (s *GuildStage) the_session_expects_a_guild_delete_event_for_the_guild() *GuildStage {
	s.session.AddHandler(func(_ *discordgo.Session, e *discordgo.GuildDelete) {
		s.t.Logf("Received %s event for guild '%s'", "GUILD_DELETE", e.Guild.ID)
		if e.Guild.ID == s.guild.ID {
			s.guildDelete = e
		}
	})

	return s
}

func (s *GuildStage) the_guild_is_deleted() {
	s.err = s.session.GuildDelete(s.guild.ID)
}

func (s *GuildStage) the_session_should_have_received_the_guild_deleted_event() {
	s.require.Eventually(func() bool {
		return s.guildDelete != nil
	}, time.Second, time.Millisecond*10)
}

func (s *GuildStage) the_guild_is_fetched() {
	s.guild, s.err = s.session.Guild(s.guild.ID)
}

func (s *GuildStage) an_error_should_be_returned() {
	s.require.Error(s.err)
}
