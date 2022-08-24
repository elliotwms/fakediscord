package fake_discord

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

func TestSessionConnects(t *testing.T) {
	OverrideEndpoints()
	go func() {
		require.NoError(t, ServeAPI())
	}()

	s, err := discordgo.New("Bot token")
	s.LogLevel = discordgo.LogDebug
	require.NoError(t, err)
	require.NotNil(t, s)

	ready := false
	s.AddHandler(func(s *discordgo.Session, _ *discordgo.Ready) {
		ready = true
	})

	err = s.Open()
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return ready
	}, 1*time.Second, 10*time.Millisecond, "Ready event should eventually be fired")
}
