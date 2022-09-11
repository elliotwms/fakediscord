package config

import "github.com/bwmarrin/snowflake"

type Config struct {
	Guilds []Guild `yaml:"guilds"`
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
