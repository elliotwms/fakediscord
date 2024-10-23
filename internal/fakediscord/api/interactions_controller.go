package api

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/fakediscord/ws"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/gin-gonic/gin"
)

func interactionsController(r *gin.RouterGroup) {
	// internal endpoint (not Discord) for initiating interactions
	r.POST("/", auth, createInteraction)

	// https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-callback
	r.POST("/:id/:token/callback", createInteractionCallback)
}

func createInteraction(c *gin.Context) {
	// create interaction ID
	interaction := &discordgo.InteractionCreate{}

	if err := c.BindJSON(interaction); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	u, ok := storage.Users.Load(c.GetString(contextKeyUserID))
	if !ok {
		_ = c.AbortWithError(http.StatusInternalServerError, errors.New("user not found"))
		return
	}

	// generate ID and token
	interaction.ID = snowflake.Generate().String()
	interaction.Token = snowflake.Generate().String()
	interaction.AppID = u.(discordgo.User).ID

	storage.Interactions.Store(interaction.Token, *interaction.Interaction)

	// todo send to webhook (query param?)

	// todo only send to bot -- not broadcast
	err := ws.DispatchEvent("INTERACTION_CREATE", interaction)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, interaction)
}

func createInteractionCallback(c *gin.Context) {
	interaction := &discordgo.InteractionResponse{}

	if err := c.BindJSON(interaction); err != nil {
		return
	}

	id := c.Param("id")
	token := c.Param("token")

	slog.Info("Received interaction callback", "id", id, "token", token, "type", interaction.Type)

	// only allow callbacks once
	_, ok := storage.InteractionCallbacks.LoadOrStore(id, struct{}{})
	if ok {
		c.JSON(http.StatusBadRequest, discordgo.APIErrorMessage{
			Message: "Interaction has already been acknowledged.",
			Code:    discordgo.ErrCodeInteractionHasAlreadyBeenAcknowledged,
		})
		return
	}

	switch interaction.Type {
	// InteractionResponsePong is for ACK ping event.
	case discordgo.InteractionResponsePong:
		//no-op
	// InteractionResponseChannelMessageWithSource is for responding with a message, showing the user's input.
	case discordgo.InteractionResponseChannelMessageWithSource:
		handleMessageInteractionResponse(c, interaction, token)
	// InteractionResponseDeferredChannelMessageWithSource acknowledges that the event was received, and that a follow-up will come later.
	case discordgo.InteractionResponseDeferredChannelMessageWithSource:
		// no-op
	// InteractionResponseDeferredMessageUpdate acknowledges that the message component interaction event was received, and message will be updated later.
	case discordgo.InteractionResponseDeferredMessageUpdate:
		c.AbortWithStatus(http.StatusNotImplemented)
	// InteractionResponseUpdateMessage is for updating the message to which message component was attached.
	case discordgo.InteractionResponseUpdateMessage:
		c.AbortWithStatus(http.StatusNotImplemented)
	// InteractionApplicationCommandAutocompleteResult shows autocompletion results. Autocomplete interaction only.
	case discordgo.InteractionApplicationCommandAutocompleteResult:
		c.AbortWithStatus(http.StatusNotImplemented)
	// InteractionResponseModal is for responding to an interaction with a modal window.
	case discordgo.InteractionResponseModal:
		c.AbortWithStatus(http.StatusNotImplemented)
	default:
		c.AbortWithStatus(http.StatusBadRequest)
	}
}

func handleMessageInteractionResponse(c *gin.Context, res *discordgo.InteractionResponse, token string) {
	v, ok := storage.Interactions.Load(token)
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	i := v.(discordgo.Interaction)

	m := discordgo.Message{
		ID:         snowflake.Generate().String(),
		Type:       discordgo.MessageTypeReply,
		ChannelID:  i.ChannelID,
		GuildID:    i.GuildID,
		Content:    res.Data.Content,
		Embeds:     res.Data.Embeds,
		Components: res.Data.Components,
		Timestamp:  time.Now(),
	}

	storage.InteractionResponses.Store(token, m.ID)
	storage.Messages.Store(m.ID, m)

	_ = ws.DispatchEvent("MESSAGE_CREATE", m)
}
