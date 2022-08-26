package fakediscord

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

type stage struct {
	session *discordgo.Session
	ready   bool

	require *require.Assertions
}

func NewStage(t *testing.T) (*stage, *stage, *stage) {
	s := &stage{
		session: nil,
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
	s.session.AddHandler(func(_ *discordgo.Session, _ *discordgo.Ready) {
		s.ready = true
	})

	return s
}

func (s *stage) the_session_is_opened() *stage {
	s.require.NoError(s.session.Open(), "session should open successfully")

	return s
}

func (s *stage) the_session_is_ready() {
	s.require.Eventually(func() bool {
		return s.ready
	}, 1*time.Second, 10*time.Millisecond, "Ready event should eventually be fired")
}
