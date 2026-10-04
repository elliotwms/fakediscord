package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGateway_URLUsesRequestHost(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/api/v9/gateway", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bot "+botToken)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()

	require.Equal(t, http.StatusOK, res.StatusCode)

	var body struct {
		URL string `json:"url"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
	require.Equal(t, "ws://127.0.0.1:8080/ws", body.URL)
}

func TestGateway_Bot(t *testing.T) {
	s := newSession(botToken)

	g, err := s.GatewayBot()
	require.NoError(t, err)

	// discordgo appends a trailing slash to the gateway URL
	require.Equal(t, "ws://localhost:8080/ws/", g.URL)
	require.Equal(t, 1, g.Shards)
	require.Equal(t, 1, g.SessionStartLimit.MaxConcurrency)
	require.NotZero(t, g.SessionStartLimit.Remaining)
}
