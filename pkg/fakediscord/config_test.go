package fakediscord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfigure(t *testing.T) {
	baseURL := "http://localhost:8080/"
	Configure(baseURL)

	require.Equal(t, discordgo.EndpointDiscord, baseURL)
}
