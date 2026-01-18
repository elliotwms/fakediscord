package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/builders"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/fakediscord/ws"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/gin-gonic/gin"
)

var stringContainsURLs = regexp.MustCompile(`((http|https|ftp)://(\S*))`)

func channelController(r *gin.RouterGroup) {
	r.Use(auth)

	r.GET("/:channel", getChannel)
	r.DELETE("/:channel", deleteChannel)

	r.GET("/:channel/pins", getChannelPins)
	r.PUT("/:channel/pins/:message", putChannelPin)

	r.POST("/:channel/messages", createChannelMessage)
	r.GET("/:channel/messages/:message", getChannelMessage)
	r.DELETE("/:channel/messages/:message", deleteChannelMessage)

	r.GET("/:channel/messages/:message/reactions/:reaction", getMessageReaction)
	r.PUT("/:channel/messages/:message/reactions/:reaction/:user", putMessageReaction)
	r.DELETE("/:channel/messages/:message/reactions/:reaction/:user", deleteMessageReaction)
	r.DELETE("/:channel/messages/:message/reactions", deleteMessageReactions)
}

// https://discord.com/developers/docs/resources/channel#get-channel
func getChannel(c *gin.Context) {
	channel, err := storage.State.Channel(c.Param("channel"))
	if err != nil {
		handleStateErr(c, err)
		return
	}

	c.JSON(http.StatusOK, channel)
}

// https://discord.com/developers/docs/resources/channel#deleteclose-channel
func deleteChannel(c *gin.Context) {
	channel, err := storage.State.Channel(c.Param("channel"))
	if err != nil {
		handleStateErr(c, err)
		return
	}

	if err = storage.State.ChannelRemove(channel); err != nil {
		handleStateErr(c, err)
		return
	}

	if _, err = ws.Connections.Broadcast("CHANNEL_DELETE", channel); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, channel)
}

// https://discord.com/developers/docs/resources/channel#get-pinned-messages
func getChannelPins(c *gin.Context) {
	var messages []*discordgo.Message

	pins := storage.Pins.Load(c.Param("channel"))
	for _, pin := range pins {
		message, err := storage.State.Message(c.Param("channel"), pin)
		if err != nil {
			handleStateErr(c, err)
			return
		}
		messages = append(messages, message)
	}

	c.JSON(http.StatusOK, messages)
}

// https://discord.com/developers/docs/resources/channel#pin-message
func putChannelPin(c *gin.Context) {
	channel, err := storage.State.Channel(c.Param("channel"))
	if err != nil {
		handleStateErr(c, err)
		return
	}

	storage.Pins.Store(c.Param("channel"), c.Param("message"))

	_, err = ws.Connections.Broadcast("CHANNEL_PINS_UPDATE", discordgo.ChannelPinsUpdate{
		LastPinTimestamp: time.Now().String(),
		ChannelID:        c.Param("channel"),
		GuildID:          channel.GuildID,
	})

	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusCreated)
}

func getChannelMessage(c *gin.Context) {
	m, err := storage.State.Message(c.Param("channel"), c.Param("message"))
	if err != nil {
		handleStateErr(c, err)
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

	user, done := getUser(c)
	if done {
		return
	}

	channel, err := storage.State.Channel(c.Param("channel"))
	if err != nil {
		_ = c.AbortWithError(http.StatusNotFound, errors.New("channel not found"))
		return
	}

	m := builders.NewMessage(&user, channel.ID, channel.GuildID).
		WithContent(messageSend.Content).
		WithEmbeds(messageSend.Embeds).
		WithAttachments(buildAttachments(channel.ID, messageSend.Files)).
		Build()

	messageCreate, err := sendMessage(m)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, messageCreate)
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

	messageSend.Embeds = append(messageSend.Embeds, getAdditionalEmbeds(messageSend.Content)...)

	return &messageSend, nil
}

// getAdditionalEmbeds determines any additional embeds which Discord would typically add to a message, such as url
// previews
func getAdditionalEmbeds(content string) []*discordgo.MessageEmbed {
	submatch := stringContainsURLs.FindAllStringSubmatch(content, -1)

	var embeds []*discordgo.MessageEmbed

	for _, s := range submatch {
		url := s[0]

		embeds = append(embeds, &discordgo.MessageEmbed{
			URL:         url,
			Type:        "rich",
			Title:       url,
			Description: url,
		})
	}

	return embeds
}

func buildAttachments(channelID string, files []*discordgo.File) []*discordgo.MessageAttachment {
	var attachments []*discordgo.MessageAttachment

	for _, f := range files {
		id := snowflake.Generate().String()
		// todo serve 'cdn' from local fs
		url := fmt.Sprintf("https://cdn.discordapp.com/attachments/%s/%s/%s", channelID, id, f.Name)

		attachment := &discordgo.MessageAttachment{
			ID:          id,
			URL:         url,
			ProxyURL:    url,
			Filename:    f.Name,
			ContentType: f.ContentType,
			Size:        1,
		}

		if isImage(f.ContentType) {
			config, _, err := image.DecodeConfig(f.Reader)
			if err != nil {
				return nil
			}

			attachment.Width = config.Width
			attachment.Height = config.Height
		}

		attachments = append(attachments, attachment)
	}

	return attachments
}

func isImage(contentType string) bool {
	return strings.Contains(contentType, "image/")
}

func deleteChannelMessage(c *gin.Context) {
	m, err := storage.State.Message(c.Param("channel"), c.Param("message"))
	if err != nil {
		handleStateErr(c, err)
		return
	}

	err = storage.State.MessageRemove(m)
	if err != nil {
		handleStateErr(c, err)
		return
	}

	_, err = ws.Connections.Broadcast("MESSAGE_DELETE", discordgo.MessageDelete{
		Message: m,
	})
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

// https://discord.com/developers/docs/resources/message#get-reactions
func getMessageReaction(c *gin.Context) {
	vs, _ := storage.Reactions.LoadMessageReaction(c.Param("message"), c.Param("reaction"))

	var users []*discordgo.User
	for _, v := range vs {
		users = append(users, &discordgo.User{ID: v})
	}

	c.JSON(http.StatusOK, users)
}

// https://discord.com/developers/docs/resources/message#create-reaction
func putMessageReaction(c *gin.Context) {
	channel, err := storage.State.Channel(c.Param("channel"))
	if err != nil {
		handleStateErr(c, err)
		return
	}

	var user *discordgo.User
	id := c.Param("user")
	if id == "@me" {
		v, done := getUser(c)
		if done {
			return
		}
		user = &v
	} else {
		v, ok := storage.Users.Load(id)
		if !ok {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		u := v.(discordgo.User)
		user = &u
	}

	e := &discordgo.MessageReactionAdd{
		MessageReaction: &discordgo.MessageReaction{
			UserID:    user.ID,
			MessageID: c.Param("message"),
			Emoji: discordgo.Emoji{
				Name: c.Param("reaction"),
			},
			ChannelID: c.Param("channel"),
			GuildID:   channel.GuildID,
		},
		Member: &discordgo.Member{
			User: user,
		},
	}

	storage.Reactions.Store(c.Param("message"), c.Param("reaction"), user.ID)

	_, err = ws.Connections.Broadcast("MESSAGE_REACTION_ADD", e)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// https://discord.com/developers/docs/resources/message#delete-user-reaction
func deleteMessageReaction(c *gin.Context) {
	channel, err := storage.State.Channel(c.Param("channel"))
	if err != nil {
		handleStateErr(c, err)
		return
	}

	var user *discordgo.User
	id := c.Param("user")
	if id == "@me" {
		v, done := getUser(c)
		if done {
			return
		}
		user = &v
	} else {
		v, ok := storage.Users.Load(id)
		if !ok {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		u := v.(discordgo.User)
		user = &u
	}

	storage.Reactions.DeleteMessageReaction(c.Param("message"), c.Param("reaction"), user.ID)

	_, err = ws.Connections.Broadcast("MESSAGE_REACTION_REMOVE", &discordgo.MessageReactionRemove{
		MessageReaction: &discordgo.MessageReaction{
			UserID:    user.ID,
			MessageID: c.Param("message"),
			Emoji: discordgo.Emoji{
				Name: c.Param("reaction"),
			},
			ChannelID: c.Param("channel"),
			GuildID:   channel.GuildID,
		},
	})
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// https://discord.com/developers/docs/resources/message#delete-all-reactions
func deleteMessageReactions(c *gin.Context) {
	m, err := storage.State.Message(c.Param("channel"), c.Param("message"))
	if err != nil {
		handleStateErr(c, err)
		return
	}

	storage.Reactions.DeleteMessageReactions(c.Param("message"))

	_, err = ws.Connections.Broadcast("MESSAGE_REACTION_REMOVE_ALL", discordgo.MessageReactionRemoveAll{
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
