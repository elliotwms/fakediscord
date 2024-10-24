package tests

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/pkg/fakediscord"
	"github.com/stretchr/testify/require"
)

type InteractionsStage struct {
	t             *testing.T
	require       *require.Assertions
	session       *discordgo.Session
	guild         *discordgo.Guild
	channel       *discordgo.Channel
	interaction   *discordgo.InteractionCreate
	handlerCalled int
}

func NewInteractionStage(t *testing.T) (given, when, then *InteractionsStage) {
	r := require.New(t)

	s := &InteractionsStage{
		t:       t,
		require: r,
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

func (s *InteractionsStage) and() *InteractionsStage {
	return s
}

func (s *InteractionsStage) a_registered_message_command_handler() *InteractionsStage {
	s.session.AddHandler(func(_ *discordgo.Session, e *discordgo.InteractionCreate) {
		s.handlerCalled++
	})

	return s
}

func (s *InteractionsStage) an_interaction_is_triggered() *InteractionsStage {
	var err error
	s.interaction, err = fakediscord.NewClient(botToken).Interaction(&discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{},
		},
	})
	s.require.NoError(err)

	return s
}

func (s *InteractionsStage) the_command_handler_should_have_been_triggered() *InteractionsStage {
	s.require.Eventually(func() bool {
		return s.handlerCalled > 0
	}, time.Second, 50*time.Millisecond)

	return s
}

func (s *InteractionsStage) the_interaction_should_be_valid() *InteractionsStage {
	s.require.NotEmpty(s.interaction.ID)
	s.require.NotEmpty(s.interaction.Token)
	s.require.NotEmpty(s.interaction.AppID)

	return s
}

func (s *InteractionsStage) the_interaction_callback_is_triggered() *InteractionsStage {
	err := s.session.InteractionRespond(s.interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	s.require.NoError(err)

	return s
}

func (s *InteractionsStage) the_interaction_callback_is_triggered_with_a_message() *InteractionsStage {
	err := s.session.InteractionRespond(s.interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Responding to interaction",
		},
	})
	s.require.NoError(err)

	return s
}

func (s *InteractionsStage) a_message_should_have_been_posted_in_the_channel() {
	res, err := s.session.InteractionResponse(s.interaction.Interaction)
	s.require.NoError(err)

	s.require.Equal(res.Content, "Responding to interaction")
}

func (s *InteractionsStage) the_interaction_message_is_updated() {
	content := "Responding to interaction"
	_, err := s.session.InteractionResponseEdit(s.interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})

	s.require.NoError(err)
}
