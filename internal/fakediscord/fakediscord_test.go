package fakediscord

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fake-discord/pkg/fakediscord"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	fakediscord.Configure("http://localhost:8080/")

	go func() {
		if err := Run(); err != nil {
			panic(err)
		}
	}()

	m.Run()
}

func TestSessionConnects(t *testing.T) {
	s := buildSession(t)

	ready := false
	s.AddHandler(func(s *discordgo.Session, _ *discordgo.Ready) {
		ready = true
	})

	require.NoError(t, s.Open())
	require.Eventually(t, func() bool {
		return ready
	}, 1*time.Second, 10*time.Millisecond, "Ready event should eventually be fired")
}

func buildSession(t *testing.T) *discordgo.Session {
	s, err := discordgo.New("Bot token")
	s.LogLevel = discordgo.LogDebug
	require.NoError(t, err)
	require.NotNil(t, s)

	return s
}
