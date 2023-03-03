package tests

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

type MessageStage struct {
	require     *require.Assertions
	session     *discordgo.Session
	guild       *discordgo.Guild
	channel     *discordgo.Channel
	messageSend *discordgo.MessageSend
	messageID   string
}

func NewMessageStage(t *testing.T) (given, then, when *MessageStage) {
	r := require.New(t)

	s := &MessageStage{
		require: r,
		session: newSession(r),
	}

	s.setup()

	return s, s, s
}

func (s *MessageStage) setup() {
	s.require.NoError(s.session.Open())

	var err error
	s.guild, err = s.session.GuildCreate("message_test")
	s.require.NoError(err)

	s.channel, err = s.session.GuildChannelCreate(s.guild.ID, "test", discordgo.ChannelTypeGuildText)
	s.require.NoError(err)
}

func (s *MessageStage) and() *MessageStage {
	return s
}

func (s *MessageStage) a_message() *MessageStage {
	s.messageSend = &discordgo.MessageSend{Content: "Hello, World!"}
	return s
}

func (s *MessageStage) the_message_is_sent() *MessageStage {
	m, err := s.session.ChannelMessageSendComplex(s.channel.ID, s.messageSend)
	s.require.NoError(err)

	s.messageID = m.ID

	return s
}

func (s *MessageStage) the_message_is_received() *MessageStage {
	s.require.Eventually(
		func() bool {
			message, err := s.session.State.Message(s.channel.ID, s.messageID)
			return err == nil && message != nil
		},
		1*time.Second,
		10*time.Millisecond,
		"message should have been received and stored in state",
	)

	return s
}

func (s *MessageStage) the_message_can_be_fetched() {
	m, err := s.session.ChannelMessage(s.channel.ID, s.messageID)
	s.require.NoError(err)
	s.require.NotNil(m)
}
