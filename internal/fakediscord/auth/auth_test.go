package auth

import (
	"testing"

	"github.com/elliotwms/fakediscord/internal/fakediscord/builders"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	_ = snowflake.Configure(0)

	m.Run()
}

func TestUsers_Authenticate(t *testing.T) {
	u := builders.
		NewUser("foo", "bar").
		WithToken("foo_token").
		Build()

	storage.Users.Store(u.ID, *u)

	authedUser := Authenticate("foo_token")
	require.NotNil(t, authedUser)

	require.Equal(t, u, authedUser)
}

func TestUsers_Authenticate_TokenNotFound(t *testing.T) {
	token := "notfound"
	u := Authenticate(token)

	require.NotNil(t, u)

	require.NotEmpty(t, u.ID)
	require.NotEmpty(t, u.Discriminator)
	require.Len(t, u.Discriminator, 4)
	require.Equal(t, token, u.Username)
}
