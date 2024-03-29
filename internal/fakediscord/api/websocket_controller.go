package api

import (
	"log"
	"net/http"

	internalws "github.com/elliotwms/fakediscord/internal/fakediscord/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func WebsocketController(r *gin.RouterGroup) {
	r.GET("/", handleWS)
}

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

	if err = internalws.Handle(ws); websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
		log.Printf("websocket error: %s", err)
	}
}
