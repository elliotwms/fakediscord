package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/auth"
	"github.com/elliotwms/fakediscord/internal/fakediscord/ws/connpool"
	"github.com/gorilla/websocket"
)

var Connections = connpool.New(slog.Default())

// https://discord.com/developers/docs/topics/opcodes-and-status-codes#gateway-gateway-opcodes
const (
	opHeartbeat    = 1
	opIdentify     = 2
	opHello        = 10
	opHeartbeatACK = 11
)

const heartbeatInterval = 10 * time.Second

func Handle(ws *websocket.Conn) error {
	conn := connpool.NewConn(ws)

	u, err := establishConnection(conn)
	if err != nil {
		return err
	}

	// once a connection is established it can be added to the pool
	// todo consider race condition between connection being established and events being broadcast intended for it

	id := Connections.Add(u.ID, conn)
	defer Connections.Remove(id)

	// todo consider refactoring
	// this is a bit of a leaky abstraction as connections are used after being added to the pool
	for {
		if err := handleMessage(conn); err != nil {
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
	// HeartbeatInterval is in milliseconds
	HeartbeatInterval int64 `json:"heartbeat_interval"`
}

func establishConnection(c *connpool.Conn) (*discordgo.User, error) {
	err := c.WriteJSON(Event{
		Operation: opHello,
		Data:      helloOp{HeartbeatInterval: heartbeatInterval.Milliseconds()},
	})
	if err != nil {
		return nil, err
	}

	slog.Debug("Waiting for identify")

	token, err := readIdentify(c)
	if err != nil {
		return nil, err
	}

	u, err := authUser(token)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	if err = ready(c, u); err != nil {
		return nil, err
	}

	if err = sendSignOnGuildCreateEvents(c); err != nil {
		return nil, err
	}

	return u, nil
}

// readIdentify waits for the identify payload and returns its token. Heartbeats sent before identifying are
// acknowledged, and any other payload is rejected
func readIdentify(c *connpool.Conn) (string, error) {
	for {
		var e struct {
			Operation int             `json:"op"`
			Data      json.RawMessage `json:"d"`
		}

		if err := c.ReadJSON(&e); err != nil {
			return "", fmt.Errorf("read identify: %w", err)
		}

		switch e.Operation {
		case opHeartbeat:
			if err := c.WriteJSON(Event{Operation: opHeartbeatACK}); err != nil {
				return "", err
			}
		case opIdentify:
			// only the token is needed from the identify payload. discordgo.Identify is not used as it does not
			// round-trip (e.g. presence.game.created_at is sent as a string but unmarshalled as an int64)
			var i struct {
				Token string `json:"token"`
			}
			if err := json.Unmarshal(e.Data, &i); err != nil {
				return "", fmt.Errorf("read identify: %w", err)
			}

			return i.Token, nil
		default:
			// resuming (op 6) is not supported, so clients must identify
			return "", fmt.Errorf("expected identify (op %d), received op %d", opIdentify, e.Operation)
		}
	}
}

func authUser(token string) (u *discordgo.User, err error) {
	s := strings.SplitN(token, " ", 2)
	if len(s) != 2 {
		return nil, errors.New("malformed token")
	}

	return auth.Authenticate(s[1]), nil
}

func handleMessage(c *connpool.Conn) error {
	var e Event

	err := c.ReadJSON(&e)
	if err != nil {
		return err
	}

	slog.Debug("Read gateway message", "op", e.Operation)

	if e.Operation == opHeartbeat {
		return c.WriteJSON(Event{Operation: opHeartbeatACK})
	}

	return nil
}
