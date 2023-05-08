package config

import "github.com/bwmarrin/snowflake"

type Config struct {
	Users  []User  `yaml:"users"`
	Guilds []Guild `yaml:"guilds"`
}

type User struct {
	ID            *snowflake.ID `yaml:"id,omitempty"`
	Token         string        `yaml:"token,omitempty"`
	Username      string        `yaml:"username"`
	Discriminator string        `yaml:"discriminator"`
	Bot           bool          `yaml:"bot,omitempty"`
}

type Guild struct {
	ID       *snowflake.ID `yaml:"id,omitempty"`
	Name     string        `yaml:"name"`
	Channels []Channel     `yaml:"channels"`
}

type Channel struct {
	ID   *snowflake.ID `yaml:"id,omitempty"`
	Name string        `yaml:"name"`
}
