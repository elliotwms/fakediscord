package fakediscord

import (
	"log"
	"net/http"

	internalws "github.com/elliotwms/fake-discord/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
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

	if err = internalws.Handle(ws); err != nil {
		log.Printf("error handling message: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

}
