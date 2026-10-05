package ws

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/fakediscord/ws/connpool"
	"github.com/elliotwms/fakediscord/internal/sequence"
)

func ready(ws *connpool.Conn, u *discordgo.User) error {
	slog.Info("Sending READY", "user_id", u.ID)

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
