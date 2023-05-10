package ws

import (
	"encoding/json"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/sequence"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/gorilla/websocket"
)

var conns sync.Map
var connMutex sync.Mutex

func register(ws *websocket.Conn) string {
	id := snowflake.Generate().String()

	conns.Store(id, ws)

	return id
}

func deregister(id string) bool {
	_, ok := conns.LoadAndDelete(id)

	return ok
}

func DispatchEvent(t string, body interface{}) error {
	bs, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return Dispatch(discordgo.Event{
		Sequence: sequence.Next(),
		Type:     t,
		RawData:  bs,
	})
}

func Dispatch(e discordgo.Event) error {
	var err error

	connMutex.Lock()
	defer connMutex.Unlock()

	conns.Range(func(_, value any) bool {
		err = value.(*websocket.Conn).WriteJSON(e)
		return err != nil
	})

	return err
}
