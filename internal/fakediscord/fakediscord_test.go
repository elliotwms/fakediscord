package fakediscord

import (
	"testing"
)

func TestSessionConnects(t *testing.T) {
	given, when, then := NewStage(t)

	given.
		a_new_session().and().
		the_session_watches_for_ready_events()

	when.
		the_session_is_opened()

	then.
		the_session_is_ready()
}
