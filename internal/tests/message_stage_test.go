package tests

import (
	"os"
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
	attachments []*discordgo.MessageAttachment
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

func (s *MessageStage) the_message_should_be_received() *MessageStage {
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

func (s *MessageStage) the_message_is_pinned() {
	s.require.NoError(s.session.ChannelMessagePin(s.channel.ID, s.messageID))
}

func (s *MessageStage) the_message_has_been_pinned() {
	pinned, err := s.session.ChannelMessagesPinned(s.channel.ID)
	s.require.NoError(err)

	found := false
	for _, message := range pinned {
		if message.ID == s.messageID {
			found = true
		}
	}

	s.require.True(found)
}

func (s *MessageStage) the_message_is_reacted_to_with(emoji string) *MessageStage {
	err := s.session.MessageReactionAdd(s.channel.ID, s.messageID, emoji)
	s.require.NoError(err)

	return s
}

func (s *MessageStage) the_message_should_have_n_reactions_to_emoji(n int, emoji string) {
	reactions, err := s.session.MessageReactions(s.channel.ID, s.messageID, emoji, 0, "", "")
	s.require.NoError(err)
	s.require.Len(reactions, n)
}

func (s *MessageStage) the_message_reactions_are_removed() {
	err := s.session.MessageReactionsRemoveAll(s.channel.ID, s.messageID)
	s.require.NoError(err)
}

func (s *MessageStage) an_attachment(filename, contentType string) {
	f, err := os.Open("files/" + filename)
	s.require.NoError(err)

	s.messageSend.Files = append(s.messageSend.Files, &discordgo.File{
		Name:        filename,
		ContentType: contentType,
		Reader:      f,
	})
}

func (s *MessageStage) the_message_should_have_n_attachments(n int) *MessageStage {
	s.require.Eventually(func() bool {
		m, err := s.session.ChannelMessage(s.channel.ID, s.messageID)

		s.attachments = m.Attachments

		return err == nil && len(m.Attachments) == n
	}, defaultWait, defaultTick)

	return s
}

func (s *MessageStage) the_message_should_have_an_attachment() *MessageStage {
	return s.the_message_should_have_n_attachments(1)
}

func (s *MessageStage) the_first_attachment_should_have_a_resolution_set() {
	s.require.NotEmpty(s.attachments)

	attachment := s.attachments[0]
	s.require.NotEmpty(attachment.Width)
	s.require.NotEmpty(attachment.Height)
}
