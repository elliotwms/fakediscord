package builders

import (
	"testing"

	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	_ = snowflake.Configure(0)

	m.Run()
}

func TestPublic(t *testing.T) {
	u := NewUser("foo", "0001").WithToken("secret").Build()

	public := Public(u)

	require.Empty(t, public.Token)
	require.Equal(t, u.ID, public.ID)
	require.Equal(t, "secret", u.Token, "original user should not be modified")
}

func TestNewMessage_AuthorHasNoToken(t *testing.T) {
	u := NewUser("foo", "0001").WithToken("secret").Build()

	m := NewMessage(u, "channel", "guild").Build()

	require.Equal(t, u.ID, m.Author.ID)
	require.Empty(t, m.Author.Token)
}

func TestNewMember_UserHasNoToken(t *testing.T) {
	u := NewUser("foo", "0001").WithToken("secret").Build()

	m := NewMember("guild", u).Build()

	require.Equal(t, u.ID, m.User.ID)
	require.Empty(t, m.User.Token)
}
