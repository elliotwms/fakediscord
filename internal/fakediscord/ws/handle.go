package ws

import (
	"encoding/json"
	"errors"
	"github.com/elliotwms/fakediscord/internal/fakediscord/ws/connpool"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/auth"
	"github.com/gorilla/websocket"
)

var Connections = connpool.New(slog.Default())

func Handle(ws *websocket.Conn) error {
	u, err := establishConnection(ws)
	if err != nil {
		return err
	}

	// once a connection is established it can be added to the pool
	// todo consider race condition between connection being established and events being broadcast intended for it

	id := Connections.Add(u.ID, ws)
	defer Connections.Remove(id)

	// todo consider refactoring
	// this is a bit of a leaky abstraction as connections are used after being added to the pool
	for {
		if err := handleMessage(ws); err != nil {
			return err
		}
	}
}

type Event struct {
	Operation int    `json:"op"`
	Sequence  int64  `json:"s"`
	Type      string `json:"t"`
	Data      any    `json:"d"`
}

type helloOp struct {
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
}

func establishConnection(c *websocket.Conn) (*discordgo.User, error) {
	err := c.WriteJSON(Event{
		Operation: 10,
		Data:      helloOp{HeartbeatInterval: 10 * time.Second},
	})
	if err != nil {
		return nil, err
	}

	log.Print("waiting for identify")

	i := &discordgo.Identify{}

	err = c.ReadJSON(&Event{Data: i})
	if err != nil && errors.Is(err, &json.UnmarshalTypeError{}) {
		// todo fix json.UnmarshalTypeError
		return nil, err
	}

	u, err := authUser(i.Token)
	if err != nil {
		log.Printf("error authing user: %s\n", err)
		return nil, c.Close()
	}

	if err = ready(c, u); err != nil {
		return nil, err
	}

	sendSignOnGuildCreateEvents(c)

	return u, nil
}

func authUser(token string) (u *discordgo.User, err error) {
	s := strings.SplitN(token, " ", 2)
	if len(s) != 2 {
		return nil, errors.New("malformed token")
	}

	return auth.Authenticate(s[1]), nil
}

func handleMessage(ws *websocket.Conn) error {
	var e Event

	err := ws.ReadJSON(&e)
	if err != nil {
		return err
	}

	log.Printf("read message %d, %v", e.Operation, e.Data)
	return nil
}
