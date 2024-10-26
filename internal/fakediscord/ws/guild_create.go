package ws

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/sequence"
	"github.com/gorilla/websocket"
)

func sendSignOnGuildCreateEvents(ws *websocket.Conn) {
	storage.State.RLock()
	defer storage.State.RUnlock()
	for _, guild := range storage.State.Guilds {
		guild := guild
		guildCreate(ws, guild)
	}
}

func guildCreate(ws *websocket.Conn, g *discordgo.Guild) {
	slog.With("guild_id", g.ID).Info("Sending GUILD_CREATE")

	err := ws.WriteJSON(Event{
		Sequence: sequence.Next(),
		Type:     "GUILD_CREATE",
		Data: discordgo.GuildCreate{
			Guild: g,
		},
	})

	if err != nil {
		panic(err)
	}
}
