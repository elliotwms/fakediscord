package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

// serve starts a websocket server which runs Handle for each connection, and reports the result of Handle on the
// returned channel. A panic in Handle fails the test
func serve(t *testing.T) (url string, result <-chan error) {
	t.Helper()

	results := make(chan error, 1)
	upgrader := websocket.Upgrader{}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		defer func() { _ = ws.Close() }()

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Handle panicked: %v", r)
				results <- nil
			}
		}()

		results <- Handle(ws)
	}))
	t.Cleanup(s.Close)

	return "ws" + strings.TrimPrefix(s.URL, "http"), results
}

func dial(t *testing.T, url string) *websocket.Conn {
	t.Helper()

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })

	require.NoError(t, c.SetReadDeadline(time.Now().Add(5*time.Second)))

	return c
}

func readHello(t *testing.T, c *websocket.Conn) helloOp {
	t.Helper()

	var e rawEvent
	require.NoError(t, c.ReadJSON(&e))
	require.Equal(t, opHello, e.Operation)

	var h helloOp
	require.NoError(t, json.Unmarshal(e.Data, &h))

	return h
}

func identify(t *testing.T, c *websocket.Conn, token string) {
	t.Helper()

	require.NoError(t, c.WriteJSON(map[string]any{
		"op": 2,
		"d": map[string]any{
			"token": token,
			// discordgo sends created_at as a string, which does not unmarshal into discordgo.Identify
			"presence": map[string]any{"game": map[string]any{"created_at": "0"}},
		},
	}))
}

func awaitResult(t *testing.T, result <-chan error) error {
	t.Helper()

	select {
	case err := <-result:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for Handle to return")
		return nil
	}
}

func TestHandle_HeartbeatIntervalIsInMilliseconds(t *testing.T) {
	url, _ := serve(t)
	c := dial(t, url)

	require.Equal(t, int64(10000), readHello(t, c).HeartbeatInterval)
}

func TestHandle_MalformedToken(t *testing.T) {
	url, result := serve(t)
	c := dial(t, url)

	readHello(t, c)
	identify(t, c, "malformed")

	require.ErrorContains(t, awaitResult(t, result), "malformed token")
}

func TestHandle_DisconnectBeforeIdentify(t *testing.T) {
	url, result := serve(t)
	c := dial(t, url)

	readHello(t, c)
	require.NoError(t, c.Close())

	require.ErrorContains(t, awaitResult(t, result), "read identify")
}

func TestHandle_HeartbeatIsAcknowledged(t *testing.T) {
	url, _ := serve(t)
	c := dial(t, url)

	readHello(t, c)
	identify(t, c, "Bot heartbeat")

	// skip past READY and any GUILD_CREATE events
	for {
		var e rawEvent
		require.NoError(t, c.ReadJSON(&e))
		if e.Type == "READY" {
			break
		}
	}

	require.NoError(t, c.WriteJSON(map[string]any{"op": opHeartbeat, "d": nil}))

	for {
		var e rawEvent
		require.NoError(t, c.ReadJSON(&e))
		if e.Operation == opHeartbeatACK {
			return
		}
	}
}

func TestHandle_HeartbeatBeforeIdentify(t *testing.T) {
	url, _ := serve(t)
	c := dial(t, url)

	readHello(t, c)
	require.NoError(t, c.WriteJSON(map[string]any{"op": opHeartbeat, "d": nil}))

	var e rawEvent
	require.NoError(t, c.ReadJSON(&e))
	require.Equal(t, opHeartbeatACK, e.Operation)

	identify(t, c, "Bot heartbeat_before_identify")

	require.NoError(t, c.ReadJSON(&e))
	require.Equal(t, "READY", e.Type)
}

func TestHandle_ResumeIsRejected(t *testing.T) {
	url, result := serve(t)
	c := dial(t, url)

	readHello(t, c)
	require.NoError(t, c.WriteJSON(map[string]any{
		"op": 6,
		"d":  map[string]any{"token": "Bot resume", "session_id": "abc", "seq": 1},
	}))

	require.ErrorContains(t, awaitResult(t, result), "expected identify")
}
