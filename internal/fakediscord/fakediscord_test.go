package fakediscord

import (
	"testing"

	"github.com/elliotwms/fake-discord/pkg/fakediscord"
)

func TestMain(m *testing.M) {
	fakediscord.Configure("http://localhost:8080/")

	go func() {
		if err := Run(); err != nil {
			panic(err)
		}
	}()

	m.Run()
}

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
