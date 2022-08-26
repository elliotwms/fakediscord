package fakediscord

import (
	"testing"

	"github.com/elliotwms/fake-discord/pkg/fakediscord"
)

func TestMain(m *testing.M) {
	setup()

	m.Run()
}

func setup() {
	fakediscord.Configure("http://localhost:8080/")

	go func() {
		if err := Run(); err != nil {
			panic(err)
		}
	}()
}
