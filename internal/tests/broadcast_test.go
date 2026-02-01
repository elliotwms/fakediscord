package tests

import (
	"sync"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

func TestBroadcast_MessageSentToMultipleConnections(t *testing.T) {
	r := require.New(t)

	// Create multiple sessions that will all receive the MESSAGE_CREATE event
	// All sessions use the same token (same user), but each creates a separate WebSocket connection
	numConnections := 3
	sessions := make([]*discordgo.Session, numConnections)
	receivedMessages := make([]chan *discordgo.MessageCreate, numConnections)

	// First session creates the guild and channel
	sessions[0], _ = newOpenSession(t, botToken)
	_, channel, err := setupGuild(t, sessions[0], "broadcast_msg")
	r.NoError(err)

	receivedMessages[0] = make(chan *discordgo.MessageCreate, 1)
	sessions[0].AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		if m.ChannelID == channel.ID {
			select {
			case receivedMessages[0] <- m:
			default:
			}
		}
	})

	// Create additional connections
	for i := 1; i < numConnections; i++ {
		receivedMessages[i] = make(chan *discordgo.MessageCreate, 1)
		sessions[i], _ = newOpenSession(t, botToken)

		idx := i
		sessions[i].AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
			if m.ChannelID == channel.ID {
				select {
				case receivedMessages[idx] <- m:
				default:
				}
			}
		})
	}

	// Cleanup sessions at the end
	t.Cleanup(func() {
		for _, s := range sessions {
			if s != nil {
				_ = s.Close()
			}
		}
	})

	// Wait for handlers to be set up
	time.Sleep(100 * time.Millisecond)

	// Send a message from the first session
	testContent := "Hello to all connections!"
	msg, err := sessions[0].ChannelMessageSend(channel.ID, testContent)
	r.NoError(err)
	r.NotNil(msg)

	// Verify all connections received the message
	var wg sync.WaitGroup
	wg.Add(numConnections)
	for i := 0; i < numConnections; i++ {
		go func(idx int) {
			defer wg.Done()
			select {
			case received := <-receivedMessages[idx]:
				r.Equal(testContent, received.Content, "connection %d should receive the correct message content", idx)
				r.Equal(msg.ID, received.ID, "connection %d should receive message with correct ID", idx)
			case <-time.After(2 * time.Second):
				t.Errorf("connection %d did not receive the message within timeout", idx)
			}
		}(i)
	}

	wg.Wait()
}

func TestBroadcast_GuildCreateSentToMultipleConnections(t *testing.T) {
	r := require.New(t)

	// Create multiple sessions that will all receive the GUILD_CREATE event
	numSessions := 3
	sessions := make([]*discordgo.Session, numSessions)
	receivedGuilds := make([]chan *discordgo.GuildCreate, numSessions)

	for i := 0; i < numSessions; i++ {
		receivedGuilds[i] = make(chan *discordgo.GuildCreate, 1)

		session, closer := newOpenSession(t, botToken)
		t.Cleanup(closer)

		sessions[i] = session

		idx := i
		session.AddHandler(func(_ *discordgo.Session, g *discordgo.GuildCreate) {
			if g.Name == "broadcast_guild_test" {
				select {
				case receivedGuilds[idx] <- g:
				default:
				}
			}
		})
	}

	// Wait for handlers to be ready
	time.Sleep(100 * time.Millisecond)

	// Create a guild from the first session - should broadcast to all
	guild, err := sessions[0].GuildCreate("broadcast_guild_test")
	r.NoError(err)
	t.Cleanup(func() {
		_ = sessions[0].GuildDelete(guild.ID)
	})

	// Verify all sessions received the GUILD_CREATE event
	var wg sync.WaitGroup
	wg.Add(numSessions)
	for i := 0; i < numSessions; i++ {
		go func(idx int) {
			defer wg.Done()
			select {
			case received := <-receivedGuilds[idx]:
				r.Equal(guild.ID, received.ID, "session %d should receive guild with correct ID", idx)
				r.Equal("broadcast_guild_test", received.Name, "session %d should receive guild with correct name", idx)
			case <-time.After(2 * time.Second):
				t.Errorf("session %d did not receive the GUILD_CREATE event within timeout", idx)
			}
		}(i)
	}

	wg.Wait()
}
