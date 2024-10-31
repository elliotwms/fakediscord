package connpool

import (
	"encoding/json"
	"errors"
	"log/slog"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/sequence"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/gorilla/websocket"
)

type Key struct{ ID, UserID string }

type ConnPool struct {
	conns sync.Map
	log   *slog.Logger
}

func New(logger *slog.Logger) *ConnPool {
	return &ConnPool{
		log: logger,
	}
}

// Add adds an established connection to the pool to receive events/broadcasts
func (p *ConnPool) Add(u string, ws *websocket.Conn) (k Key) {
	k = Key{
		ID:     snowflake.Generate().String(),
		UserID: u,
	}

	p.conns.Store(k, ws)

	return
}

func (p *ConnPool) Remove(k Key) (ok bool) {
	_, ok = p.conns.LoadAndDelete(k)

	return
}

// Broadcast sends event with type t and payload body to all registered connections
func (p *ConnPool) Broadcast(t string, body interface{}) (n int, err error) {
	p.log.Info("Broadcasting event", "type", t)

	bs, err := json.Marshal(body)
	if err != nil {
		return n, err
	}

	e := discordgo.Event{
		Sequence: sequence.Next(),
		Type:     t,
		RawData:  bs,
	}

	p.conns.Range(func(_, value any) bool {
		err = value.(*websocket.Conn).WriteJSON(e)
		return err != nil
	})

	return 0, err
}

// Send sends an event of type t and payload body to the first registered connection for a given user ID
// returns ok if a match was found
func (p *ConnPool) Send(userID, t string, body interface{}) (ok bool, err error) {
	bs, err := json.Marshal(body)
	if err != nil {
		return false, err
	}

	return p.send(userID, discordgo.Event{
		Sequence: sequence.Next(),
		Type:     t,
		RawData:  bs,
	})
}

func (p *ConnPool) send(userID string, event discordgo.Event) (ok bool, err error) {
	var errs []error

	p.conns.Range(func(k, value any) bool {
		if k.(Key).UserID == userID {
			err := value.(*websocket.Conn).WriteJSON(event)
			if err != nil {
				errs = append(errs, err)
			} else {
				ok = true
			}

			return err != nil
		}

		return true
	})

	return ok, errors.Join(errs...)
}
