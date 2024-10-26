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
		_ = c.AbortWithError(http.StatusNotFound, errors.New("interaction not found"))
		return
	}

	v, ok := storage.Interactions.Load(c.Param("token"))
	if !ok {
		_ = c.AbortWithError(http.StatusNotFound, errors.New("interaction not found"))
	}
	i := v.(discordgo.Interaction)

	slog.Info("loading message", "id", mID)
	m, err := storage.State.Message(i.ChannelID, mID.(string))
	if err != nil {
		handleStateErr(c, err)
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

	v, _ = storage.InteractionResponses.LoadOrStore(token, snowflake.Generate().String())
	id := v.(string)

	v, ok = storage.Users.Load(interaction.AppID)
	if !ok {
		_ = c.AbortWithError(http.StatusNotFound, fmt.Errorf("user not found"))
		return
	}
	user := v.(discordgo.User)

	m, err := storage.State.Message(interaction.ChannelID, id)
	if err != nil {
		if !errors.Is(err, discordgo.ErrStateNotFound) {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// build a new message
		slog.Info("message not found, creating new message", "id", id, "token", token)

		m = builders.NewMessage(&user, interaction.ChannelID, interaction.GuildID).
			WithID(id).
			WithType(discordgo.MessageTypeReply).
			Build()
	}

	updateMessage(m, edit)

	mc, err := sendMessage(m)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, mc)
}

// updateMessage applies the discordgo.WebhookEdit to the message
func updateMessage(m *discordgo.Message, edit *discordgo.WebhookEdit) {
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
}
