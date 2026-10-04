package connpool

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	_ = snowflake.Configure(0)

	m.Run()
}

// TestConnPool_ConcurrentBroadcast checks that concurrent broadcasts do not write to the same connection
// concurrently (which panics in gorilla/websocket), and that events are received in sequence order
func TestConnPool_ConcurrentBroadcast(t *testing.T) {
	const n = 200

	p := New(slog.Default())
	upgrader := websocket.Upgrader{}
	added := make(chan struct{})

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		k := p.Add("user", NewConn(ws))
		close(added)

		// hold the connection open until the client goes away
		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				p.Remove(k)
				return
			}
		}
	}))
	defer s.Close()

	c, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http"), nil)
	require.NoError(t, err)
	defer func() { _ = c.Close() }()
	<-added

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := p.Broadcast("TEST", i)
			require.NoError(t, err)
		}()
	}

	var last int64
	for i := 0; i < n; i++ {
		var e struct {
			Sequence int64 `json:"s"`
		}
		require.NoError(t, c.ReadJSON(&e))
		require.Greater(t, e.Sequence, last, "events should be received in sequence order")
		last = e.Sequence
	}

	wg.Wait()
}

// TestConnPool_BroadcastToStalledConnection checks that a client which stops reading does not block broadcasts
// indefinitely, and that its connection is closed
func TestConnPool_BroadcastToStalledConnection(t *testing.T) {
	writeTimeout = 100 * time.Millisecond
	t.Cleanup(func() { writeTimeout = 10 * time.Second })

	p := New(slog.Default())
	upgrader := websocket.Upgrader{}
	added := make(chan struct{})
	closed := make(chan struct{})

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		k := p.Add("user", NewConn(ws))
		close(added)

		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				p.Remove(k)
				close(closed)
				return
			}
		}
	}))
	defer s.Close()

	// the client connects but never reads
	c, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http"), nil)
	require.NoError(t, err)
	defer func() { _ = c.Close() }()
	<-added

	payload := strings.Repeat("x", 1<<20)

	done := make(chan error)
	go func() {
		// keep writing until the socket buffers fill and a write times out
		for {
			if _, err := p.Broadcast("TEST", payload); err != nil {
				done <- err
				return
			}
		}
	}()

	select {
	case err := <-done:
		require.Error(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("broadcast blocked on a stalled connection")
	}

	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("stalled connection was not closed")
	}
}
