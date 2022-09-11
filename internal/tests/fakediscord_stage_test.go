package tests

import (
	"sync"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

const defaultWait = 1 * time.Second
const defaultTick = 10 * time.Millisecond

type stage struct {
	session *discordgo.Session

	require *require.Assertions

	ready *discordgo.Ready

	guildCreateMX sync.RWMutex
	guildCreate   []*discordgo.GuildCreate

	messagesMX sync.RWMutex
	messages   []*discordgo.MessageCreate

	testGuild *discordgo.GuildCreate
}

func NewStage(t *testing.T) (*stage, *stage, *stage) {
	s := &stage{
		require: require.New(t),
	}

	return s, s, s
}

func (s *stage) and() *stage {
	return s
}

func (s *stage) a_new_session() *stage {
	var err error
	s.session, err = discordgo.New("Bot token")
	s.session.LogLevel = discordgo.LogDebug

	s.require.NoError(err, "New should be called without error")

	return s
}

func (s *stage) the_session_watches_for_ready_events() *stage {
	s.session.AddHandler(func(_ *discordgo.Session, r *discordgo.Ready) {
		s.ready = r
	})

	return s
}

func (s *stage) the_session_watches_for_guild_create_events() *stage {
	s.session.AddHandler(func(_ *discordgo.Session, c *discordgo.GuildCreate) {
		s.guildCreateMX.Lock()
		defer s.guildCreateMX.Unlock()

		s.guildCreate = append(s.guildCreate, c)

		// keep track of the test guild for IDs later on
		if c.Name == "Test Guild" {
			s.testGuild = c
		}
	})

	return s
}

func (s *stage) the_session_watches_for_message_created_events() *stage {
	s.session.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		s.messagesMX.Lock()
		defer s.messagesMX.Unlock()

		s.messages = append(s.messages, m)
	})

	return s
}

func (s *stage) the_session_is_opened() *stage {
	s.require.NoError(s.session.Open(), "session should open successfully")

	return s
}

func (s *stage) the_session_is_ready() *stage {
	s.require.Eventually(func() bool {
		return s.ready != nil
	}, defaultWait, defaultTick, "Ready event should eventually be fired")

	return s
}

func (s *stage) the_ready_has_n_guilds(n int) *stage {
	s.require.Len(s.ready.Guilds, n)

	return s
}

func (s *stage) the_session_receives_n_guild_create_events(n int) *stage {
	currLen := 0

	s.require.Eventuallyf(
		func() bool {
			s.guildCreateMX.RLock()
			defer s.guildCreateMX.RUnlock()
			currLen = len(s.guildCreate)

			return currLen == n
		},
		defaultWait,
		defaultTick,
		"did not receive expected GUILD_CREATE events. Expected: '%d', actual: '%d'", n, currLen,
	)

	return s
}

func (s *stage) an_established_session() *stage { //nolint:unused
	return s.
		a_new_session().and().
		the_session_watches_for_ready_events().and().
		the_session_watches_for_guild_create_events().and().
		the_session_is_opened().and().
		the_session_is_ready().and().
		the_session_receives_n_guild_create_events(len(s.ready.Guilds))
}

func (s *stage) a_message_is_created() *stage {
	_, err := s.session.ChannelMessageSend(s.testGuild.Channels[0].ID, "Hello, world!")
	s.require.NoError(err)

	return s
}

func (s *stage) n_message_created_events_are_received(n int) {
	currLen := 0

	s.require.Eventuallyf(
		func() bool {
			s.messagesMX.RLock()
			defer s.messagesMX.RUnlock()

			currLen = len(s.messages)
			return currLen == n
		},
		defaultWait,
		defaultTick,
		"Message Created event should eventually be fired Expected: '%d', actual: '%d'", n, currLen,
	)
}
