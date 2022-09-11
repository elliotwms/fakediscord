package ws

import (
	"github.com/bwmarrin/discordgo"
	"sync"

	"github.com/gorilla/websocket"
)

var conns []*websocket.Conn
var connsMX sync.RWMutex

func register(ws *websocket.Conn) {
	connsMX.Lock()
	defer connsMX.Unlock()

	conns = append(conns, ws)
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
