package tests

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

const defaultWait = 1 * time.Second
const defaultTick = 10 * time.Millisecond

type SessionStage struct {
	session *discordgo.Session

	require *require.Assertions

	ready *discordgo.Ready

	guildCreateMX sync.RWMutex
	guildCreate   []*discordgo.GuildCreate

	messagesMX sync.RWMutex
	messages   []*discordgo.MessageCreate

	testGuild *discordgo.GuildCreate
}

func NewSessionStage(t *testing.T) (given, when, then *SessionStage, cleanup func()) {
	s := &SessionStage{
		require: require.New(t),
	}

	return s, s, s, s.the_session_is_closed
}

func (s *SessionStage) and() *SessionStage {
	return s
}

func (s *SessionStage) a_new_session() *SessionStage {
	s.session = newSession(s.require)

	return s
}

func (s *SessionStage) the_session_watches_for_ready_events() *SessionStage {
	s.session.AddHandler(func(_ *discordgo.Session, r *discordgo.Ready) {
		s.ready = r
	})

	return s
}

func (s *SessionStage) the_session_watches_for_guild_create_events() *SessionStage {
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

func (s *SessionStage) the_session_watches_for_message_created_events() *SessionStage {
	s.session.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		s.messagesMX.Lock()
		defer s.messagesMX.Unlock()

		log.Print("MESSAGE RECEIVED")

		s.messages = append(s.messages, m)
	})

	return s
}

func (s *SessionStage) the_session_is_opened() *SessionStage {
	s.require.NoError(s.session.Open(), "session should open successfully")

	return s
}

func (s *SessionStage) the_session_is_closed() {
	s.require.NoError(s.session.Close(), "session should close successfully")
}

func (s *SessionStage) the_session_is_ready() *SessionStage {
	s.require.Eventually(func() bool {
		return s.ready != nil
	}, defaultWait, defaultTick, "Ready event should eventually be fired")

	return s
}

func (s *SessionStage) the_session_receives_guild_create_events() *SessionStage {
	n := len(s.session.State.Ready.Guilds)

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