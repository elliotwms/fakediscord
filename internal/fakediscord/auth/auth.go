package auth

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/builders"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
)

func Authenticate(token string) *discordgo.User {
	u := checkState(token)

	if u != nil {
		return u
	}

	u = builders.
		NewUser(token, fmt.Sprintf("%04d", rand.Intn(10000))).
		WithToken(token).
		Build()

	log.Printf("Created user %s with username %s for token %s\n", u.ID, u.Username, u.Token)

	storage.Users.Store(u.ID, *u)

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
