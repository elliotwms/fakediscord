package auth

import (
	"fmt"
	"log/slog"
	"math/rand"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/builders"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
)

// mx serialises Authenticate so that concurrent first requests with the same unknown token (e.g. a bot's gateway
// connection and an HTTP request) resolve to the same generated user
var mx sync.Mutex

func Authenticate(token string) *discordgo.User {
	mx.Lock()
	defer mx.Unlock()

	u := checkState(token)

	if u != nil {
		return u
	}

	u = builders.
		NewUser(token, fmt.Sprintf("%04d", rand.Intn(10000))).
		WithToken(token).
		Build()

	slog.Info("Created user for unknown token", "id", u.ID, "username", u.Username)

	storage.Users.Store(u.ID, *u)

	// todo add to State.Members?

	return u
}

func checkState(token string) (u *discordgo.User) {
	storage.Users.Range(func(key, value any) bool {
		v := value.(discordgo.User)

		if v.Token == token {
			u = &v
			return false
		}

		return true
	})

	return
}
