package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPins_Store_Idempotent(t *testing.T) {
	Pins.Store("pins-idempotent", "foo")
	Pins.Store("pins-idempotent", "foo")

	require.Equal(t, []string{"foo"}, Pins.Load("pins-idempotent"))
}

func TestPins_Delete(t *testing.T) {
	Pins.Store("pins-delete", "foo")
	Pins.Store("pins-delete", "bar")

	loaded := Pins.Load("pins-delete")

	Pins.Delete("pins-delete", "foo")

	require.Equal(t, []string{"bar"}, Pins.Load("pins-delete"))
	require.Equal(t, []string{"foo", "bar"}, loaded, "previously loaded pins should not be modified")
}

func TestPins_Load_Missing(t *testing.T) {
	require.Empty(t, Pins.Load("pins-missing"))
}
