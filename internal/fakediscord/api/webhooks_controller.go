package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/fakediscord/ws"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/gin-gonic/gin"
)

func webhooksController(r *gin.RouterGroup) {
	// https://discord.com/developers/docs/interactions/receiving-and-responding#get-original-interaction-response
	r.GET(":appID/:token/messages/@original", getResponse)

	// https://discord.com/developers/docs/interactions/receiving-and-responding#edit-original-interaction-response
	r.PATCH(":appID/:token/messages/@original", patchResponse)

	// todo DELETE /webhooks/<application_id>/<interaction_token>/messages/@original to delete your initial response to an Interaction
	// https://discord.com/developers/docs/interactions/receiving-and-responding#delete-original-interaction-response

	// todo followup messages
	// https://discord.com/developers/docs/interactions/receiving-and-responding#create-followup-message
}

func getResponse(c *gin.Context) {
	mID, ok := storage.InteractionResponses.Load(c.Param("token"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	slog.Info("loading message", "id", mID)
	m, ok := storage.Messages.Load(mID)
	if !ok {
		_ = c.AbortWithError(http.StatusNotFound, fmt.Errorf("message not found"))
		return
	}

	c.JSON(http.StatusOK, m)
}

func patchResponse(c *gin.Context) {
	edit := &discordgo.WebhookEdit{}
	err := c.BindJSON(edit)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	token := c.Param("token")
	v, ok := storage.Interactions.Load(token)
	if !ok {
		_ = c.AbortWithError(http.StatusNotFound, fmt.Errorf("interaction not found"))
		return
	}
	interaction := v.(discordgo.Interaction)

	id, _ := storage.InteractionResponses.LoadOrStore(token, snowflake.Generate().String())

	v, ok = storage.Messages.LoadOrStore(id, discordgo.Message{
		ID:        id.(string),
		ChannelID: interaction.ChannelID,
		GuildID:   interaction.GuildID,
		Timestamp: time.Now(),
		// todo author -- should be bot user, but do not have bot token in this context
	})
	if !ok {
		slog.Info("message not found, creating new message", "id", id, "token", token)
	}
	m := v.(discordgo.Message)

	m = updateMessage(m, edit)

	// store the updated message
	storage.Messages.Store(id, m)
	slog.Info("Stored message", "id", id, "mID", m.ID, "token", token, "content", m.Content)

	t := "MESSAGE_CREATE"
	if !ok {
		t = "MESSAGE_UPDATE"
	}
	if err := ws.DispatchEvent(t, m); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, m)
}

func updateMessage(m discordgo.Message, edit *discordgo.WebhookEdit) discordgo.Message {
	if edit.Content != nil {
		m.Content = *edit.Content
	}

	if edit.Embeds != nil {
		m.Embeds = *edit.Embeds
	}

	// todo files?

	if edit.Attachments != nil {
		m.Attachments = *edit.Attachments
	}

	if edit.Components != nil {
		m.Components = *edit.Components
	}

	// todo allowed mentions?

	return m
}
