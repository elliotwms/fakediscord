package api

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/fakediscord/ws"
)

// sendMessage stores the message in fakediscord's internal state and dispatches a create/update event (depending on
// if the message is new).
func sendMessage(m *discordgo.Message) (*discordgo.MessageCreate, error) {
	t := "MESSAGE_CREATE"

	_, loaded := storage.Messages.Swap(m.ID, *m)
	if loaded {
		// message already exists, update it instead
		t = "MESSAGE_UPDATE"
	}

	messageCreate := &discordgo.MessageCreate{Message: m}
	if err := ws.DispatchEvent(t, messageCreate); err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}
	return messageCreate, nil
}
