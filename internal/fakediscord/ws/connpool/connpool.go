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

// Conn wraps a websocket connection so that writes from multiple goroutines are serialised, as gorilla/websocket
// supports at most one concurrent writer per connection
type Conn struct {
	mx sync.Mutex
	ws *websocket.Conn
}

func NewConn(ws *websocket.Conn) *Conn {
	return &Conn{ws: ws}
}

func (c *Conn) WriteJSON(v any) error {
	c.mx.Lock()
	defer c.mx.Unlock()

	return c.ws.WriteJSON(v)
}

// ReadJSON reads the next message from the connection. Only one goroutine may read at a time
func (c *Conn) ReadJSON(v any) error {
	return c.ws.ReadJSON(v)
}

func (c *Conn) Close() error {
	return c.ws.Close()
}

type ConnPool struct {
	conns sync.Map
	log   *slog.Logger

	// dispatch is held while an event is sequenced and written, so that every connection receives events in
	// sequence order
	dispatch sync.Mutex
}

func New(logger *slog.Logger) *ConnPool {
	return &ConnPool{
		log: logger,
	}
}

// Add adds an established connection to the pool to receive events/broadcasts
func (p *ConnPool) Add(u string, c *Conn) (k Key) {
	k = Key{
		ID:     snowflake.Generate().String(),
		UserID: u,
	}

	p.conns.Store(k, c)

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

	p.dispatch.Lock()
	defer p.dispatch.Unlock()

	e := discordgo.Event{
		Sequence: sequence.Next(),
		Type:     t,
		RawData:  bs,
	}

	var errs []error
	p.conns.Range(func(_, value any) bool {
		if writeErr := value.(*Conn).WriteJSON(e); writeErr != nil {
			errs = append(errs, writeErr)
		} else {
			n++
		}
		return true // continue to all connections
	})

	return n, errors.Join(errs...)
}

// Send sends an event of type t and payload body to the first registered connection for a given user ID
// returns ok if a match was found
func (p *ConnPool) Send(userID, t string, body interface{}) (ok bool, err error) {
	bs, err := json.Marshal(body)
	if err != nil {
		return false, err
	}

	p.dispatch.Lock()
	defer p.dispatch.Unlock()

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
			err := value.(*Conn).WriteJSON(event)
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
