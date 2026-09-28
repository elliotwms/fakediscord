package ws

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/fakediscord/ws/connpool"
	"github.com/elliotwms/fakediscord/internal/sequence"
)

func sendSignOnGuildCreateEvents(ws *connpool.Conn) error {
	storage.State.RLock()
	defer storage.State.RUnlock()
	for _, guild := range storage.State.Guilds {
		if err := guildCreate(ws, guild); err != nil {
			return err
		}
	}

	return nil
}

func guildCreate(ws *connpool.Conn, g *discordgo.Guild) error {
	slog.With("guild_id", g.ID).Info("Sending GUILD_CREATE")

	return ws.WriteJSON(Event{
		Sequence: sequence.Next(),
		Type:     "GUILD_CREATE",
		Data: discordgo.GuildCreate{
			Guild: g,
		},
	})
}
