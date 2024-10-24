package tests

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

type GuildCommandStage struct {
	t          *testing.T
	require    *require.Assertions
	session    *discordgo.Session
	guild      *discordgo.Guild
	channel    *discordgo.Channel
	command    *discordgo.ApplicationCommand
	err        error
	originalID string
}

func NewGuildCommandStage(t *testing.T) (*GuildCommandStage, *GuildCommandStage, *GuildCommandStage) {
	s := &GuildCommandStage{
		t:       t,
		require: require.New(t),
		session: newSession(botToken),
	}

	s.require.NoError(s.session.Open())
	t.Cleanup(func() {
		s.require.NoError(s.session.Close())
	})

	var err error
	s.guild, s.channel, err = setupGuild(s.session, "message")
	s.require.NoError(err)

	return s, s, s
}

func (s *GuildCommandStage) and() *GuildCommandStage {
	return s
}

func (s *GuildCommandStage) a_guild_command() *GuildCommandStage {
	s.command = &discordgo.ApplicationCommand{
		Type: 1,
		Name: "command",
	}

	return s
}

func (s *GuildCommandStage) the_command_is_created() {
	s.command, s.err = s.session.ApplicationCommandCreate(appID, s.guild.ID, s.command)
}

func (s *GuildCommandStage) no_error_should_be_returned() *GuildCommandStage {
	s.require.NoError(s.err)

	return s
}

func (s *GuildCommandStage) the_command_is_fetched() *GuildCommandStage {
	s.command, s.err = s.session.ApplicationCommand(appID, s.guild.ID, s.command.ID)

	return s
}

func (s *GuildCommandStage) the_command_is_deleted() {
	s.err = s.session.ApplicationCommandDelete(appID, s.guild.ID, s.command.ID)
}

func (s *GuildCommandStage) an_error_should_be_returned() *GuildCommandStage {
	s.require.Error(s.err)

	return s
}

func (s *GuildCommandStage) the_command_is_recreated() {
	s.originalID = s.command.ID
	s.command, s.err = s.session.ApplicationCommandCreate(appID, s.guild.ID, &discordgo.ApplicationCommand{
		Type: s.command.Type,
		Name: s.command.Name,
	})
}

func (s *GuildCommandStage) the_id_should_not_have_changed() {
	s.require.Equal(s.originalID, s.command.ID)
}
