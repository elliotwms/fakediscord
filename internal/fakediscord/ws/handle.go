package ws

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/auth"
	"github.com/gorilla/websocket"
)

func Handle(ws *websocket.Conn) error {
	if err := establishConnection(ws); err != nil {
		return err
	}

	id := register(ws)
	defer deregister(id)

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

func establishConnection(c *websocket.Conn) error {
	log.Print("establishing connection")
	err := c.WriteJSON(Event{
		Operation: 10,
		Data:      helloOp{HeartbeatInterval: 10 * time.Second},
	})
	if err != nil {
		return err
	}

	log.Print("waiting for identify")

	i := &discordgo.Identify{}

	err = c.ReadJSON(&Event{Data: i})
	if err != nil && errors.Is(err, &json.UnmarshalTypeError{}) {
		// todo fix json.UnmarshalTypeError
		return err
	}

	u, err := authUser(i.Token)
	if err != nil {
		log.Printf("error authing user: %s\n", err)
		return c.Close()
	}

	if err = ready(c, u); err != nil {
		return err
	}

	go sendSignOnGuildCreateEvents(c)

	return nil
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
