package ws

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/sequence"
	"github.com/gorilla/websocket"
)

func ready(ws *websocket.Conn, u *discordgo.User) error {
	log.Print("sending READY")

	return ws.WriteJSON(Event{
		Type:     "READY",
		Sequence: sequence.Next(),
		Data:     buildReady(u),
	})
}

func buildReady(u *discordgo.User) discordgo.Ready {
	r := discordgo.Ready{
		User: u,
	}

	storage.State.RLock()
	defer storage.State.RUnlock()

	for _, guild := range storage.State.Guilds {
		r.Guilds = append(r.Guilds, &discordgo.Guild{
			// READY returns a stripped down guild containing just the ID and availability
			ID:          guild.ID,
			Unavailable: true,
		})
	}

	return r
}
