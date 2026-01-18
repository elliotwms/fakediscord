package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReactionStore_LoadMessageReaction_MissingMessage(t *testing.T) {
	users, ok := Reactions.LoadMessageReaction("missing", "💩")

	assert.Empty(t, users)
	assert.False(t, ok)
}

func TestReactionStore_LoadMessageReaction_MissingReaction(t *testing.T) {
	Reactions.Store("found", "1️⃣", "@me")

	users, ok := Reactions.LoadMessageReaction("found", "2️⃣")

	assert.Empty(t, users)
	assert.False(t, ok)
}

func TestReactionStore_StoreAndLoad(t *testing.T) {
	Reactions.Store("foo", "🧀", "bar")

	users, ok := Reactions.LoadMessageReaction("foo", "🧀")

	require.Len(t, users, 1)
	assert.Contains(t, users, "bar")
	assert.True(t, ok)
}

func TestReactionStore_DeleteMessageReaction(t *testing.T) {
	Reactions.Store("deleteme", "🗑", "foo")
	Reactions.Store("deleteme", "🗑", "bar")

	Reactions.DeleteMessageReaction("deleteme", "🗑", "foo")
	users, ok := Reactions.LoadMessageReaction("deleteme", "🗑")

	require.NotEmpty(t, users)
	assert.NotContains(t, users, "foo")
	assert.True(t, ok)
}

func TestReactionStore_DeleteMessageReactions(t *testing.T) {
	Reactions.Store("deleteme", "🗑", "@me")

	Reactions.DeleteMessageReactions("deleteme")
	users, ok := Reactions.LoadMessageReaction("deleteme", "🗑")

	require.Empty(t, users)
	require.False(t, ok)
}
