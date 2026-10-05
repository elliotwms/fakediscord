package fakediscord

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

// recorder starts a server which responds with the request's host and escaped path
func recorder(t *testing.T) *url.URL {
	t.Helper()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, r.Host+r.URL.EscapedPath())
	}))
	t.Cleanup(s.Close)

	u, err := url.Parse(s.URL)
	require.NoError(t, err)

	return u
}

func get(t *testing.T, c *http.Client, rawURL string) string {
	t.Helper()

	res, err := c.Get(rawURL)
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()

	bs, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return string(bs)
}

func TestTransport_RedirectsDiscord(t *testing.T) {
	fake := recorder(t)
	c := &http.Client{Transport: Transport(fake, nil)}

	require.Equal(t, fake.Host+"/api/v9/channels/1", get(t, c, "https://discord.com/api/v9/channels/1"))
}

func TestTransport_PassesThroughOtherHosts(t *testing.T) {
	fake := recorder(t)
	other := recorder(t)
	c := &http.Client{Transport: Transport(fake, nil)}

	require.Equal(t, other.Host+"/foo", get(t, c, other.String()+"/foo"))
}

func TestTransport_PathPrefix(t *testing.T) {
	fake := recorder(t)
	prefixed := *fake
	prefixed.Path = "/fakediscord/"
	c := &http.Client{Transport: Transport(&prefixed, nil)}

	require.Equal(t, fake.Host+"/fakediscord/api/v9/gateway", get(t, c, "https://discord.com/api/v9/gateway"))

	// escaped paths, such as reactions with a # emoji, are kept escaped
	require.Equal(t,
		fake.Host+"/fakediscord/api/v9/channels/1/messages/2/reactions/%23%E2%83%A3/@me",
		get(t, c, "https://discord.com/api/v9/channels/1/messages/2/reactions/%23%E2%83%A3/@me"),
	)
}

func TestTransport_DoesNotModifyRequest(t *testing.T) {
	fake := recorder(t)
	c := &http.Client{Transport: Transport(fake, nil)}

	req, err := http.NewRequest(http.MethodGet, "https://discord.com/api/v9/gateway", nil)
	require.NoError(t, err)

	res, err := c.Do(req)
	require.NoError(t, err)
	_ = res.Body.Close()

	require.Equal(t, "https://discord.com/api/v9/gateway", req.URL.String())
}

func TestConfigureSession(t *testing.T) {
	fake := recorder(t)

	s, err := discordgo.New("Bot token")
	require.NoError(t, err)
	s.Client.Timeout = 5 * time.Second

	require.NoError(t, ConfigureSession(s, fake.String()+"/"))

	require.Equal(t, 5*time.Second, s.Client.Timeout, "existing client settings should be kept")
	require.Equal(t, fake.Host+"/api/v9/foo", get(t, s.Client, "https://discord.com/api/v9/foo"))
}

func TestConfigureSession_InvalidURL(t *testing.T) {
	s, err := discordgo.New("Bot token")
	require.NoError(t, err)

	require.Error(t, ConfigureSession(s, "://"))
}
