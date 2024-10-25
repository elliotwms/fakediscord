package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/builders"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
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
		_ = c.AbortWithError(http.StatusNotFound, errors.New("token not found"))
		return
	}

	slog.Info("loading message", "id", mID)
	m, ok := storage.Messages.Load(mID)
	if !ok {
		_ = c.AbortWithError(http.StatusNotFound, errors.New("message not found"))
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

	v, ok = storage.Users.Load(interaction.AppID)
	if !ok {
		_ = c.AbortWithError(http.StatusNotFound, fmt.Errorf("user not found"))
		return
	}
	user := v.(discordgo.User)

	v, ok = storage.Messages.LoadOrStore(
		id,
		*builders.NewMessage(&user, interaction.ChannelID, interaction.GuildID).
			WithID(id.(string)).
			WithType(discordgo.MessageTypeReply).
			Build(),
	)
	if !ok {
		slog.Info("message not found, creating new message", "id", id, "token", token)
	}
	m := v.(discordgo.Message)

	m = updateMessage(m, edit)

	mc, err := sendMessage(&m)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, mc)
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
