package tests

import (
	"testing"
)

func TestSession_Connects(t *testing.T) {
	given, when, then, cleanup := NewStage(t)
	defer cleanup()

	given.
		a_new_session().and().
		the_session_watches_for_ready_events().and().
		the_session_watches_for_guild_create_events()

	when.
		the_session_is_opened()

	then.
		the_session_is_ready().and().
		the_ready_has_n_guilds(2).and().
		the_session_receives_n_guild_create_events(2)
}

func TestSession_CreateMessage(t *testing.T) {
	given, when, then, cleanup := NewStage(t)
	defer cleanup()

	given.
		an_established_session().and().
		the_session_watches_for_message_created_events()

	when.
		a_message_is_created()

	then.
		n_message_created_events_are_received(1)
}
