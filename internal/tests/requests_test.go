package tests

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

func TestGuild_CreateChannel_MissingGuild(t *testing.T) {
	s := newSession(botToken)

	_, err := s.GuildChannelCreate("1", "test", discordgo.ChannelTypeGuildText)

	var restErr *discordgo.RESTError
	require.ErrorAs(t, err, &restErr)
	require.Equal(t, http.StatusNotFound, restErr.Response.StatusCode)
}

func TestRequest_MalformedJSON(t *testing.T) {
	for name, path := range map[string]string{
		"create guild":   "guilds",
		"create command": "applications/" + appID + "/commands",
	} {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/v9/"+path, strings.NewReader("{"))
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bot "+botToken)
			req.Header.Set("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			require.Equal(t, http.StatusBadRequest, res.StatusCode)
		})
	}
}

func TestRequest_MissingAuthorization(t *testing.T) {
	res, err := http.Get("http://localhost:8080/api/v9/guilds/1")
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestApplicationCommands_EmptyListIsNotNull(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/api/v9/applications/1/commands", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bot "+botToken)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	bs, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.JSONEq(t, "[]", string(bs))
}
