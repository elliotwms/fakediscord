package ws

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"
)

func Handle(ws *websocket.Conn) error {
	if err := establishConnection(ws); err != nil {
		return err
	}

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

	i := Event{Data: discordgo.Identify{}}
	err = c.ReadJSON(&i)
	if err != nil {
		return err
	}

	if err = ready(c); err != nil {
		return err
	}

	go sendSignOnGuildCreateEvents(c)

	return nil
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
