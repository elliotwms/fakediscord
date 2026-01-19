package tests

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

type MessageStage struct {
	t           *testing.T
	require     *require.Assertions
	session     *discordgo.Session
	guild       *discordgo.Guild
	channel     *discordgo.Channel
	messageSend *discordgo.MessageSend
	messageID   string
	attachments []*discordgo.MessageAttachment
	embeds      []*discordgo.MessageEmbed

	mu      sync.Mutex
	adds    []*discordgo.MessageReactionAdd
	removes []*discordgo.MessageReactionRemove
}

func NewMessageStage(t *testing.T) (given, then, when *MessageStage) {
	r := require.New(t)

	s := &MessageStage{
		t:       t,
		require: r,
		session: newSession(botToken),
	}

	s.setup()

	return s, s, s
}

func (s *MessageStage) setup() {
	s.require.NoError(s.session.Open())

	var err error
	s.guild, s.channel, err = setupGuild(s.t, s.session, "message")
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

func (s *MessageStage) the_message_can_be_fetched() *MessageStage {
	m, err := s.session.ChannelMessage(s.channel.ID, s.messageID)
	s.require.NoError(err)
	s.require.NotNil(m)

	return s
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

func (s *MessageStage) the_message_should_have_n_reactions_to_emoji(n int, emoji string) *MessageStage {
	reactions, err := s.session.MessageReactions(s.channel.ID, s.messageID, emoji, 0, "", "")
	s.require.NoError(err)
	s.require.Len(reactions, n)

	return s
}

func (s *MessageStage) the_message_reactions_are_removed() {
	err := s.session.MessageReactionsRemoveAll(s.channel.ID, s.messageID)
	s.require.NoError(err)
}

func (s *MessageStage) the_message_reaction_is_removed(name string) *MessageStage {
	err := s.session.MessageReactionRemove(s.channel.ID, s.messageID, name, "@me")
	s.require.NoError(err)

	return s
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

func (s *MessageStage) the_message_has_the_author_as_the_session_user() {
	message, err := s.session.State.Message(s.channel.ID, s.messageID)
	s.require.NoError(err)
	s.require.Equal(s.session.State.User.ID, message.Author.ID)
}

func (s *MessageStage) the_message_has_a_link() {
	s.messageSend.Content += " https://github.com/elliotwms/fakediscord"
}

func (s *MessageStage) the_message_should_have_an_embed() *MessageStage {
	return s.the_message_should_have_n_embeds(1)
}

func (s *MessageStage) the_message_should_have_n_embeds(n int) *MessageStage {
	s.require.Eventually(func() bool {
		m, err := s.session.ChannelMessage(s.channel.ID, s.messageID)

		s.embeds = m.Embeds

		return err == nil && len(m.Embeds) == n
	}, defaultWait, defaultTick)

	return s
}

func (s *MessageStage) we_listen_for_message_reaction_events() *MessageStage {
	s.adds = []*discordgo.MessageReactionAdd{}
	s.session.AddHandler(func(_ *discordgo.Session, e *discordgo.MessageReactionAdd) {
		s.mu.Lock()
		s.adds = append(s.adds, e)
		s.mu.Unlock()
	})

	s.removes = []*discordgo.MessageReactionRemove{}
	s.session.AddHandler(func(_ *discordgo.Session, e *discordgo.MessageReactionRemove) {
		s.mu.Lock()
		s.removes = append(s.removes, e)
		s.mu.Unlock()
	})

	return s
}

func (s *MessageStage) a_message_reaction_add_event_should_have_been_received_with_id_and_name(id, name string) {
	// todo fix test
	s.t.Skip("concurrency issue -- events are not received consistently")
	s.require.Eventually(func() bool {
		s.mu.Lock()
		adds := s.adds
		s.mu.Unlock()
		for _, e := range adds {
			if e.Emoji.ID == id && e.Emoji.Name == name {
				return true
			}
		}

		return false
	}, defaultWait, defaultTick)
}

func (s *MessageStage) a_message_reaction_remove_event_should_have_been_received_with_id_and_name(id, name string) {
	// todo fix test
	s.t.Skip("concurrency issue -- events are not received consistently")
	s.require.Eventually(func() bool {
		s.mu.Lock()
		removes := s.removes
		s.mu.Unlock()
		for _, e := range removes {
			if e.Emoji.ID == id && e.Emoji.Name == name {
				return true
			}
		}

		return false
	}, defaultWait, defaultTick)
}
