package ws

import (
	"encoding/json"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/sequence"
	"github.com/gorilla/websocket"
)

var conns []*websocket.Conn
var connsMX sync.RWMutex

func register(ws *websocket.Conn) {
	connsMX.Lock()
	defer connsMX.Unlock()

	conns = append(conns, ws)
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
	connsMX.RLock()
	defer connsMX.RUnlock()

	for i := range conns {
		err := conns[i].WriteJSON(e)
		if err != nil {
			return err
		}
	}

	return nil
}
