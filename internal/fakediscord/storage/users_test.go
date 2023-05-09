package storage

import (
	"github.com/elliotwms/fakediscord/internal/fakediscord/builders"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/stretchr/testify/require"
	"testing"
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

	Users.Store(u.ID, *u)

	require.NotNil(t, Authenticate("foo_token"))
}

func TestUsers_Authenticate_TokenNotFound(t *testing.T) {
	require.Nil(t, Authenticate("notfound"))
}
