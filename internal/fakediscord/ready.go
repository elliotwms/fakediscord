package fakediscord

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/snowflake"
	"github.com/elliotwms/fake-discord/internal/sequence"
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
	r := discordgo.Ready{}

	guilds.Range(func(key, value any) bool {
		r.Guilds = append(r.Guilds, &discordgo.Guild{
			// READY returns a stripped down guild containing just the ID and availability
			ID:          key.(snowflake.ID).String(),
			Unavailable: true,
		})

		return true
	})

	return r
}
