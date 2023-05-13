package tests

import (
	"testing"
)

func TestSession_Connects(t *testing.T) {
	given, when, then, cleanup := NewSessionStage(t)
	defer cleanup()

	given.
		a_new_session().and().
		the_session_watches_for_ready_events().and().
		the_session_watches_for_guild_create_events()

	when.
		the_session_is_opened()

	then.
		the_session_is_ready().and().
		the_session_receives_guild_create_events()
}

func TestSession_ConnectsWithGeneratedUser(t *testing.T) {
	given, when, then, cleanup := NewSessionStage(t)
	defer cleanup()

	given.
		a_new_session_with_token("username").and().
		the_session_watches_for_ready_events().and().
		the_session_watches_for_guild_create_events()

	when.
		the_session_is_opened()

	then.
		the_session_is_ready().and().
		the_session_has_username("username").and().
		the_session_receives_guild_create_events()
}
