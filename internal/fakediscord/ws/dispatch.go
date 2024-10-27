package ws

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/sequence"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/gorilla/websocket"
)

var conns sync.Map

func register(ws *websocket.Conn) string {
	id := snowflake.Generate().String()

	conns.Store(id, ws)

	return id
}

func deregister(id string) bool {
	_, ok := conns.LoadAndDelete(id)

	return ok
}

// Broadcast sends event with type t and payload body to all registered connections
func Broadcast(t string, body interface{}) error {
	bs, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return broadcast(discordgo.Event{
		Sequence: sequence.Next(),
		Type:     t,
		RawData:  bs,
	})
}

func broadcast(e discordgo.Event) error {
	slog.Info("Dispatching event", slog.String("type", e.Type))

	var err error

	conns.Range(func(_, value any) bool {
		err = value.(*websocket.Conn).WriteJSON(e)
		return err != nil
	})

	return err
}
