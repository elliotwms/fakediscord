package ws

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/sequence"
	"github.com/gorilla/websocket"
)

func ready(ws *websocket.Conn) error {
	log.Print("sending READY")

	return ws.WriteJSON(Event{
		Type:     "READY",
		Sequence: sequence.Next(),
		Data:     buildReady(),
	})
}

func buildReady() discordgo.Ready {
	r := discordgo.Ready{
		// todo determine caller
		User: &discordgo.User{ID: "@me"},
	}

	storage.Guilds.Range(func(key, value any) bool {
		r.Guilds = append(r.Guilds, &discordgo.Guild{
			// READY returns a stripped down guild containing just the ID and availability
			ID:          key.(string),
			Unavailable: true,
		})

		return true
	})

	return r
}
