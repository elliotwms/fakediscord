package storage

import (
	"github.com/bwmarrin/discordgo"
	"sync"
)

var Users sync.Map

func Authenticate(token string) (u *discordgo.User) {
	Users.Range(func(key, value any) bool {
		v := value.(discordgo.User)

		if v.Token == token {
			u = &v
			return false
		}

		return true
	})

	return
}
