package api

import (
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/fakediscord/ws"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/gin-gonic/gin"
)

func channelController(r *gin.RouterGroup) {
	r.DELETE("/:channel", deleteChannel)

	r.GET("/:channel/pins", getChannelPins)
	r.PUT("/:channel/pins/:message", putChannelPin)

	r.POST("/:channel/messages", createChannelMessage)
	r.GET("/:channel/messages/:message", getChannelMessage)
	r.GET("/:channel/messages/:message/reactions/:reaction", getMessageReaction)
	r.PUT("/:channel/messages/:message/reactions/:reaction/:user", putMessageReaction)
}

func deleteChannel(c *gin.Context) {
	channel, ok := storage.Channels.LoadAndDelete(c.Param("channel"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, channel)
}

func getChannelPins(c *gin.Context) {
	var messages []discordgo.Message

	pins := storage.Pins.Load(c.Param("channel"))
	for _, pin := range pins {
		v, ok := storage.Messages.Load(pin)
		if ok {
			messages = append(messages, v.(discordgo.Message))
		}
	}

	c.JSON(http.StatusOK, messages)
}

func putChannelPin(c *gin.Context) {
	channel, ok := storage.Channels.Load(c.Param("channel"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	storage.Pins.Store(c.Param("channel"), c.Param("message"))

	err := ws.DispatchEvent("CHANNEL_PINS_UPDATE", discordgo.ChannelPinsUpdate{
		LastPinTimestamp: time.Now().String(),
		ChannelID:        c.Param("channel"),
		GuildID:          channel.(discordgo.Channel).GuildID,
	})

	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusCreated)
}

func getChannelMessage(c *gin.Context) {
	m, ok := storage.Messages.Load(c.Param("message"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, m)
}

// https://discord.com/developers/docs/resources/channel#create-message
func createChannelMessage(c *gin.Context) {
	var messageSend discordgo.MessageSend

	if err := c.BindJSON(&messageSend); err != nil {
		return
	}

	m := discordgo.Message{
		ID:        snowflake.Generate().String(),
		ChannelID: c.Param("channel"),
		Content:   messageSend.Content,
		Timestamp: time.Now(),
		// todo set author as caller identity
		Author: &discordgo.User{
			ID:            snowflake.Generate().String(),
			Username:      "username",
			Discriminator: "0000",
		},
		Embeds: messageSend.Embeds,
	}
	messageCreate := discordgo.MessageCreate{Message: &m}
	if err := ws.DispatchEvent("MESSAGE_CREATE", messageCreate); err != nil {
		c.AbortWithStatus(500)
		return
	}

	storage.Messages.Store(m.ID, m)

	c.JSON(http.StatusCreated, discordgo.MessageCreate{
		Message: &m,
	})
}

func getMessageReaction(c *gin.Context) {
	vs, ok := storage.Reactions.LoadMessageReaction(c.Param("message"), c.Param("reaction"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	var users []*discordgo.User
	for _, v := range vs {
		users = append(users, &discordgo.User{ID: v})
	}

	c.JSON(http.StatusOK, users)
}

func putMessageReaction(c *gin.Context) {
	v, ok := storage.Channels.Load(c.Param("channel"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	channel := v.(discordgo.Channel)

	e := &discordgo.MessageReactionAdd{
		MessageReaction: &discordgo.MessageReaction{
			// todo storage for users, for auth testing and lookup (@me should resolve here)
			UserID:    c.Param("user"),
			MessageID: c.Param("message"),
			Emoji: discordgo.Emoji{
				Name: c.Param("reaction"),
			},
			ChannelID: c.Param("channel"),
			GuildID:   channel.GuildID,
		},
		Member: &discordgo.Member{
			User: &discordgo.User{
				// todo resolve an actual user
				ID: snowflake.Generate().String(),
			},
		},
	}

	storage.Reactions.Store(c.Param("message"), c.Param("reaction"), c.Param("user"))

	err := ws.DispatchEvent("MESSAGE_REACTION_ADD", e)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, c.Request.Body)
}
