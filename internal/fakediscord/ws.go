package fakediscord

import (
	"log"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Event struct {
	Operation int         `json:"op"`
	Sequence  int64       `json:"s"`
	Type      string      `json:"t"`
	Data      interface{} `json:"d"`
}

type helloOp struct {
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
}

func handleWS(c *gin.Context) {
	log.Print("handling websocket request")
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	defer func() {
		if err := ws.Close(); err != nil {
			log.Printf("failed to close websocket: %s", err)
			return
		}
	}()

	if err = establishConnection(ws); err != nil {
		log.Printf("error establishing connection: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	for {
		if err := handleMessage(ws); err != nil {
			log.Printf("error handling message: %s", err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}
}

func establishConnection(ws *websocket.Conn) error {
	log.Print("establishing connection")
	err := ws.WriteJSON(Event{
		Operation: 10,
		Data:      helloOp{HeartbeatInterval: 10 * time.Second},
	})
	if err != nil {
		return err
	}

	log.Print("waiting for identify")

	i := Event{Data: discordgo.Identify{}}
	err = ws.ReadJSON(&i)
	if err != nil {
		return err
	}

	log.Print("sending READY")
	err = ws.WriteJSON(Event{
		Type:     "READY",
		Sequence: 1,
		Data:     discordgo.Ready{}},
	)
	if err != nil {
		return err
	}

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
