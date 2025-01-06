package tests

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/elliotwms/fakediscord/pkg/fakediscord"
	"github.com/stretchr/testify/require"
)

type InteractionsStage struct {
	t                 *testing.T
	require           *require.Assertions
	session           *discordgo.Session
	guild             *discordgo.Guild
	channel           *discordgo.Channel
	interactionCreate *discordgo.InteractionCreate
	interaction       *discordgo.InteractionCreate
	handlerCalled     int
	err               error
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
	s.guild, s.channel, err = setupGuild(t, s.session, "message")
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

func (s *InteractionsStage) the_interaction_is_triggered() *InteractionsStage {
	if s.interactionCreate == nil {
		s.a_valid_interaction()
	}

	s.interaction, s.err = fakediscord.NewClient(botToken).Interaction(s.interactionCreate)

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

func (s *InteractionsStage) an_interaction() *InteractionsStage {
	s.interactionCreate = &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{},
	}

	return s
}

func (s *InteractionsStage) a_valid_interaction() *InteractionsStage {
	s.interactionCreate = &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			AppID: s.session.State.User.ID,
			Type:  discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				ID:          snowflake.Generate().String(),
				Name:        "interaction",
				CommandType: discordgo.ChatApplicationCommand,
			},
			GuildID:   s.guild.ID,
			ChannelID: s.channel.ID,
		},
	}

	return s
}

func (s *InteractionsStage) no_error_should_be_returned() *InteractionsStage {
	s.require.NoError(s.err)

	return s
}

func (s *InteractionsStage) an_error_should_be_returned() *InteractionsStage {
	s.require.Error(s.err)

	return s
}

func (s *InteractionsStage) the_error_should_contain(contains string) *InteractionsStage {
	s.require.ErrorContains(s.err, contains)

	return s
}

func (s *InteractionsStage) the_interaction_has(modifier func(i *discordgo.InteractionCreate)) {
	modifier(s.interactionCreate)
}
