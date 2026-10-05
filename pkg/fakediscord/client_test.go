package fakediscord

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_WithBaseURL_AddsTrailingSlash(t *testing.T) {
	c := NewClient("token").WithBaseURL("http://localhost:8080")

	require.Equal(t, "http://localhost:8080/", c.baseURL)
}

func TestClient_BaseURLNotSet(t *testing.T) {
	c := NewClient("token")
	c.baseURL = "https://discord.com/"

	_, err := c.Interaction(nil)
	require.ErrorContains(t, err, "base URL not set")
}
