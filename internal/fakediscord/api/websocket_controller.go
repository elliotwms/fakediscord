package api

import (
	"errors"
	"log/slog"
	"net"
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
	slog.Debug("Handling websocket request")
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	defer func() {
		// the connection may already have been closed after a failed write
		if err := ws.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			slog.Error("Failed to close websocket", "err", err)
		}
	}()

	if err = internalws.Handle(ws); err != nil && !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		slog.Warn("Websocket error", "err", err)
	}
}
