package fakediscord

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fake-discord/internal/sequence"
	"github.com/gorilla/websocket"
)

func sendSignOnGuildCreateEvents(ws *websocket.Conn) {
	guilds.Range(func(key, value any) bool {
		guildCreate(ws, value.(discordgo.Guild))

		return true
	})
}

func guildCreate(ws *websocket.Conn, g discordgo.Guild) {
	log.Print("SENDING GUILD_CREATE")

	err := ws.WriteJSON(Event{
		Sequence: sequence.Next(),
		Type:     "GUILD_CREATE",
		Data: discordgo.GuildCreate{
			Guild: &g,
		},
	})

	if err != nil {
		panic(err)
	}
}
