package tests

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord"
	"github.com/elliotwms/fakediscord/pkg/config"
	pkgfakediscord "github.com/elliotwms/fakediscord/pkg/fakediscord"
	"gopkg.in/yaml.v2"
)

const baseURL = "http://localhost:8080/"
const botToken = "token"
const appID = "1290742494824366183"

//go:embed files/config.yml
var configDir embed.FS

func TestMain(m *testing.M) {
	setup()

	code := m.Run()

	if err := waitForSessionsToClose(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}

	os.Exit(code)
}

func setup() {
	c := readConfig()

	go func() {
		if err := fakediscord.Run(context.Background(), ":8080", c); err != nil {
			panic(err)
		}
	}()

	// Wait for server to be ready
	waitForServer("http://localhost:8080/api/v9/gateway", 5*time.Second)
}

func waitForServer(url string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	panic("server did not start within timeout")
}

func readConfig() config.Config {
	bs, err := configDir.ReadFile("files/config.yml")
	if err != nil {
		panic(err)
	}

	var c config.Config
	if err := yaml.Unmarshal(bs, &c); err != nil {
		panic(err)
	}

	return c
}

func newSession(token string) *discordgo.Session {
	session, _ := discordgo.New("Bot " + token)
	if err := pkgfakediscord.ConfigureSession(session, baseURL); err != nil {
		panic(err)
	}

	if os.Getenv("DEBUG") != "" {
		session.LogLevel = discordgo.LogDebug
		session.Debug = true
	}

	session.State.MaxMessageCount = 100

	return session
}

func newOpenSession(t *testing.T, token string) (session *discordgo.Session, closer func()) {
	session = newSession(token)

	err := session.Open()
	if err != nil {
		t.Fatal(err)
	}

	return session, func() {
		closeSession(session)
	}
}

// discordgo's Session.Close sleeps for a second after sending the close frame, waiting for the server to close the
// connection. Closing sessions synchronously in each test's cleanup therefore adds a second to every test, so they
// are closed in the background instead, and TestMain waits for them all to finish
var (
	closing     sync.WaitGroup
	closeErrsMx sync.Mutex
	closeErrs   []error
)

// closeSession closes the session in the background
func closeSession(s *discordgo.Session) {
	closing.Add(1)
	go func() {
		defer closing.Done()

		if err := s.Close(); err != nil {
			closeErrsMx.Lock()
			closeErrs = append(closeErrs, fmt.Errorf("close session: %w", err))
			closeErrsMx.Unlock()
		}
	}()
}

// waitForSessionsToClose waits for sessions closed with closeSession, and returns any errors from closing them
func waitForSessionsToClose() error {
	closing.Wait()

	closeErrsMx.Lock()
	defer closeErrsMx.Unlock()

	return errors.Join(closeErrs...)
}

func setupGuild(t *testing.T, s *discordgo.Session, name string) (*discordgo.Guild, *discordgo.Channel, error) {
	guild, err := s.GuildCreate(fmt.Sprintf("%s_test", name))
	if err != nil {
		return nil, nil, err
	}
	t.Cleanup(func() {
		_ = s.GuildDelete(guild.ID)
	})

	channel, err := s.GuildChannelCreate(guild.ID, "test", discordgo.ChannelTypeGuildText)

	return guild, channel, err
}
