package tests

import (
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

func newSession(require *require.Assertions, token string) *discordgo.Session {
	session, err := discordgo.New("Bot " + token)
	require.NoError(err)

	if os.Getenv("DEBUG") != "" {
		session.LogLevel = discordgo.LogDebug
		session.Debug = true
	}

	session.State.MaxMessageCount = 100

	return session
}
