package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/fakediscord/ws"
	"github.com/gin-gonic/gin"
)

func getUser(c *gin.Context) (discordgo.User, bool) {
	u, ok := storage.Users.Load(c.GetString(contextKeyUserID))
	if !ok {
		_ = c.AbortWithError(http.StatusInternalServerError, errors.New("user missing from state"))
		return discordgo.User{}, true
	}

	user := u.(discordgo.User)
	return user, false
}

// sendMessage stores the message in fakediscord's internal state and dispatches a create/update event (depending on
// if the message is new).
func sendMessage(m *discordgo.Message) (*discordgo.MessageCreate, error) {
	t := "MESSAGE_CREATE"

	_, loaded := storage.Messages.Swap(m.ID, *m)
	if loaded {
		// message already exists, update it instead
		t = "MESSAGE_UPDATE"
	}

	messageCreate := &discordgo.MessageCreate{Message: m}
	if err := ws.DispatchEvent(t, messageCreate); err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}
	return messageCreate, nil
}

func handleStateErr(c *gin.Context, err error) {
	if errors.Is(err, discordgo.ErrStateNotFound) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	_ = c.AbortWithError(http.StatusInternalServerError, err)
}
