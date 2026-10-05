package fakediscord

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// discordHost is the host discordgo sends REST requests to. The gateway URL is fetched via REST, so redirecting this
// host is enough to point a session's websocket at fakediscord too
const discordHost = "discord.com"

// Transport returns an http.RoundTripper which sends requests addressed to Discord to the fakediscord instance at
// baseURL instead, and passes all other requests through to next (http.DefaultTransport if nil).
//
// It can be used with any http.Client, for bots which call the Discord API without discordgo.
func Transport(baseURL *url.URL, next http.RoundTripper) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}

	return &transport{base: baseURL, next: next}
}

type transport struct {
	base *url.URL
	next http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host != discordHost {
		return t.next.RoundTrip(req)
	}

	// a RoundTripper must not modify the request it is given
	req = req.Clone(req.Context())
	req.URL.Scheme = t.base.Scheme
	req.URL.Host = t.base.Host
	req.Host = ""

	// support fakediscord being served under a path prefix, e.g. http://proxy/fakediscord/
	if prefix := strings.TrimSuffix(t.base.Path, "/"); prefix != "" {
		req.URL.Path = prefix + req.URL.Path
		if req.URL.RawPath != "" {
			req.URL.RawPath = prefix + req.URL.RawPath
		}
	}

	return t.next.RoundTrip(req)
}

// ConfigureSession points a discordgo session at the fakediscord instance at baseURL (e.g. "http://localhost:8080/"),
// by wrapping the session's HTTP client with Transport. Unlike Configure, it does not modify discordgo's package-level
// endpoints, so it covers every endpoint and sessions can be pointed at different instances.
//
// Call it before opening the session.
func ConfigureSession(s *discordgo.Session, baseURL string) error {
	u, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	c := &http.Client{}
	if s.Client != nil {
		*c = *s.Client
	}
	c.Transport = Transport(u, c.Transport)

	s.Client = c

	return nil
}
