package tests

import (
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
	"os"
)

func newSession(require *require.Assertions) *discordgo.Session {
	session, err := discordgo.New("Bot token")
	require.NoError(err)

	if os.Getenv("DEBUG") != "" {
		session.LogLevel = discordgo.LogDebug
		session.Debug = true
	}

	session.State.MaxMessageCount = 100

	return session
}
