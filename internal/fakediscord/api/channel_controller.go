package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	r.DELETE("/:channel/messages/:message", deleteChannelMessage)
	r.GET("/:channel/messages/:message", getChannelMessage)
	r.DELETE("/:channel/messages/:message/reactions", deleteMessageReactions)
	r.GET("/:channel/messages/:message/reactions/:reaction", getMessageReaction)
	r.PUT("/:channel/messages/:message/reactions/:reaction/:user", putMessageReaction)
}

func deleteChannel(c *gin.Context) {
	channel, ok := storage.Channels.LoadAndDelete(c.Param("channel"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if err := ws.DispatchEvent("CHANNEL_DELETE", channel); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
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
	messageSend, err := parseMessageSend(c)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
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
		Embeds:      messageSend.Embeds,
		Attachments: buildAttachments(c.Param("channel"), messageSend.Files),
	}

	storage.Messages.Store(m.ID, m)

	messageCreate := discordgo.MessageCreate{Message: &m}
	if err := ws.DispatchEvent("MESSAGE_CREATE", messageCreate); err != nil {
		c.AbortWithStatus(500)
		return
	}

	c.JSON(http.StatusOK, discordgo.MessageCreate{
		Message: &m,
	})
}

func parseMessageSend(c *gin.Context) (*discordgo.MessageSend, error) {
	var messageSend discordgo.MessageSend

	switch c.ContentType() {
	case "application/json":
		if err := c.BindJSON(&messageSend); err != nil {
			return nil, err
		}
	case "multipart/form-data":
		form, err := c.MultipartForm()
		if err != nil {
			return nil, err
		}

		payload, ok := form.Value["payload_json"]
		if !ok {
			return nil, errors.New("missing payload_json")
		}

		if len(payload) == 0 {
			return nil, errors.New("missing payload")
		}

		err = json.Unmarshal([]byte(payload[0]), &messageSend)
		if err != nil {
			return nil, err
		}

		for s, headers := range form.File {
			log.Printf("Parsing file %s", s)

			open, err := headers[0].Open()
			if err != nil {
				return nil, err
			}
			file := &discordgo.File{
				Name:        headers[0].Filename,
				ContentType: headers[0].Header.Get("Content-Type"),
				Reader:      open,
			}
			messageSend.Files = append(messageSend.Files, file)
		}
	default:
		return nil, fmt.Errorf("unhandled content type %s", c.ContentType())
	}

	return &messageSend, nil
}

func buildAttachments(channelID string, files []*discordgo.File) []*discordgo.MessageAttachment {
	var attachments []*discordgo.MessageAttachment

	for _, f := range files {
		id := snowflake.Generate().String()
		// todo serve 'cdn' from local fs
		url := fmt.Sprintf("https://cdn.discordapp.com/attachments/%s/%s/%s", channelID, id, f.Name)

		attachments = append(attachments, &discordgo.MessageAttachment{
			ID:          id,
			URL:         url,
			ProxyURL:    url,
			Filename:    f.Name,
			ContentType: f.ContentType,
			Width:       1,
			Height:      2,
			Size:        3,
		})
	}

	return attachments
}

func deleteChannelMessage(c *gin.Context) {
	m, ok := storage.Messages.LoadAndDelete(c.Param("message"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	message := m.(discordgo.Message)
	err := ws.DispatchEvent("MESSAGE_DELETE", discordgo.MessageDelete{
		Message: &message,
	})
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

func getMessageReaction(c *gin.Context) {
	vs, _ := storage.Reactions.LoadMessageReaction(c.Param("message"), c.Param("reaction"))

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

	c.Status(http.StatusNoContent)
}

func deleteMessageReactions(c *gin.Context) {
	v, ok := storage.Messages.Load(c.Param("message"))
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	m := v.(discordgo.Message)

	storage.Reactions.DeleteMessageReactions(c.Param("message"))

	err := ws.DispatchEvent("MESSAGE_REACTION_REMOVE_ALL", discordgo.MessageReactionRemoveAll{
		MessageReaction: &discordgo.MessageReaction{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		},
	})
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}
