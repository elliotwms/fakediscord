package tests

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

type ApplicationCommandStage struct {
	t          *testing.T
	require    *require.Assertions
	session    *discordgo.Session
	guild      *discordgo.Guild
	channel    *discordgo.Channel
	command    *discordgo.ApplicationCommand
	err        error
	originalID string
}

func NewApplicationCommandStage(t *testing.T) (*ApplicationCommandStage, *ApplicationCommandStage, *ApplicationCommandStage) {
	s := &ApplicationCommandStage{
		t:       t,
		require: require.New(t),
		session: newSession(botToken),
	}

	s.require.NoError(s.session.Open())
	t.Cleanup(func() {
		s.require.NoError(s.session.Close())
	})

	return s, s, s
}

func (s *ApplicationCommandStage) and() *ApplicationCommandStage {
	return s
}

func (s *ApplicationCommandStage) a_guild() *ApplicationCommandStage {
	var err error
	s.guild, s.channel, err = setupGuild(s.t, s.session, "message")
	s.require.NoError(err)

	return s
}

func (s *ApplicationCommandStage) an_application_command() *ApplicationCommandStage {
	s.command = &discordgo.ApplicationCommand{
		Type: 1,
		Name: "command",
	}

	return s
}

func (s *ApplicationCommandStage) the_command_is_created_globally() {
	s.command, s.err = s.session.ApplicationCommandCreate(appID, "", s.command)
}

func (s *ApplicationCommandStage) the_command_is_created_in_guild() {
	s.command, s.err = s.session.ApplicationCommandCreate(appID, s.guild.ID, s.command)
}

func (s *ApplicationCommandStage) no_error_should_be_returned() *ApplicationCommandStage {
	s.require.NoError(s.err)

	return s
}

func (s *ApplicationCommandStage) the_command_is_fetched() *ApplicationCommandStage {
	s.command, s.err = s.session.ApplicationCommand(appID, s.guildID(), s.command.ID)

	return s
}

func (s *ApplicationCommandStage) the_command_is_deleted() {
	s.err = s.session.ApplicationCommandDelete(appID, s.guildID(), s.command.ID)
}

func (s *ApplicationCommandStage) an_error_should_be_returned() *ApplicationCommandStage {
	s.require.Error(s.err)

	return s
}

func (s *ApplicationCommandStage) the_command_is_recreated() {
	s.originalID = s.command.ID

	s.command, s.err = s.session.ApplicationCommandCreate(appID, s.guildID(), &discordgo.ApplicationCommand{
		Type: s.command.Type,
		Name: s.command.Name,
	})
}

func (s *ApplicationCommandStage) guildID() string {
	guildID := ""
	if s.guild != nil {
		guildID = s.guild.ID
	}
	return guildID
}

func (s *ApplicationCommandStage) the_id_should_not_have_changed() {
	s.require.Equal(s.originalID, s.command.ID)
}
