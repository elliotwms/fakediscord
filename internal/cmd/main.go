package main

import "github.com/elliotwms/fake-discord/internal/fakediscord"

func main() {
	if err := fakediscord.Run(); err != nil {
		panic(err)
	}
}
