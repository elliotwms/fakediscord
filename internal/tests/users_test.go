package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUser_Guilds(t *testing.T) {
	s := newSession(botToken)

	guilds, err := s.UserGuilds(0, "", "", false)
	require.NoError(t, err)
	require.NotEmpty(t, guilds)

	for _, g := range guilds {
		require.NotNil(t, g)
		require.NotEmpty(t, g.ID)
	}
}
