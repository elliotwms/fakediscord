package api

import (
	"encoding/json"
	"github.com/elliotwms/fakediscord/internal/fakediscord/ws"
	"github.com/elliotwms/fakediscord/internal/sequence"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
)

func channelController(r *gin.RouterGroup) {
	r.POST("/:id/messages", createChannelMessage)
}

// https://discord.com/developers/docs/resources/channel#create-message
func createChannelMessage(c *gin.Context) {
	var messageSend discordgo.MessageSend

	if err := c.BindJSON(&messageSend); err != nil {
		return
	}

	messageCreate := discordgo.MessageCreate{
		Message: &discordgo.Message{
			ID:        snowflake.Generate().String(),
			ChannelID: c.Param("id"),
			Content:   messageSend.Content,
			Timestamp: time.Time{},
		},
	}

	bs, err := json.Marshal(messageCreate)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}

	e := discordgo.Event{
		Sequence: sequence.Next(),
		Type:     "MESSAGE_CREATE",
		RawData:  bs,
	}

	if err := ws.Dispatch(e); err != nil {
		c.AbortWithStatus(500)
		return
	}

	c.JSON(http.StatusCreated, messageCreate)
}
