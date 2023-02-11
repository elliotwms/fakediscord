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
	require.Equal(t, discordgo.EndpointGateway, baseURL+"gateway")

	url := baseURL + "api/v9/applications/1/guilds/2/commands"
	require.Equal(t, url, discordgo.EndpointApplicationGuildCommands("1", "2"))
}
