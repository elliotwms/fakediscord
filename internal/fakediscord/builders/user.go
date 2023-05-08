package builders

import (
	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/elliotwms/fakediscord/pkg/config"
)

type User struct {
	u *discordgo.User
}

func NewUser(username, discriminator string) *User {
	return &User{u: &discordgo.User{
		ID:            snowflake.Generate().String(),
		Username:      username,
		Discriminator: discriminator,
		Token:         snowflake.Generate().String(),
		Verified:      true,
	}}
}

func NewUserFromConfig(config config.User) *User {
	user := NewUser(config.Username, config.Discriminator)

	if config.Token != "" {
		user.WithToken(config.Token)
	}

	return user
}

func (u *User) Build() *discordgo.User {
	return u.u
}

func (u *User) WithToken(token string) *User {
	u.u.Token = token

	return u
}
