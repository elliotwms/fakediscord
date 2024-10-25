package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/builders"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/fakediscord/ws"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/gin-gonic/gin"
)

func interactionsController(r *gin.RouterGroup) {
	// internal endpoint (not Discord) for initiating interactions
	r.POST("/", auth, createInteraction)

	// https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-callback
	r.POST("/:id/:token/callback", postCallback)
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

func postCallback(c *gin.Context) {
	res := &discordgo.InteractionResponse{}

	if err := c.BindJSON(res); err != nil {
		return
	}

	id, token := c.Param("id"), c.Param("token")

	slog.Info("Received interaction callback", "id", id, "token", token, "type", res.Type)

	// get the original interaction
	v, ok := storage.Interactions.Load(token)
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	i := v.(discordgo.Interaction)

	// only allow callbacks once
	_, ok = storage.InteractionCallbacks.LoadOrStore(id, struct{}{})
	if ok {
		c.JSON(http.StatusBadRequest, discordgo.APIErrorMessage{
			Message: "Interaction has already been acknowledged.",
			Code:    discordgo.ErrCodeInteractionHasAlreadyBeenAcknowledged,
		})
		return
	}

	switch res.Type {
	// InteractionResponsePong is for ACK ping event.
	case discordgo.InteractionResponsePong:
		//no-op
	// InteractionResponseChannelMessageWithSource is for responding with a message, showing the user's input.
	case discordgo.InteractionResponseChannelMessageWithSource:
		handleMessageInteractionResponse(c, res, token)
	// InteractionResponseDeferredChannelMessageWithSource acknowledges that the event was received, and that a follow-up will come later.
	case discordgo.InteractionResponseDeferredChannelMessageWithSource:
		emitLoadingMessage(c, i, res, token)
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

	m := builders.NewMessage(i.User, i.ChannelID, i.GuildID).
		WithType(discordgo.MessageTypeReply).
		WithContent(res.Data.Content).
		WithEmbeds(res.Data.Embeds).
		WithComponents(res.Data.Components).
		Build()

	storage.InteractionResponses.Store(token, m.ID)

	_, err := sendMessage(m)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func emitLoadingMessage(c *gin.Context, i discordgo.Interaction, res *discordgo.InteractionResponse, token string) {
	v, ok := storage.Users.Load(i.AppID)
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	u := v.(discordgo.User)

	m := builders.NewMessage(&u, i.ChannelID, i.GuildID).
		WithType(discordgo.MessageTypeReply).
		WithFlags(discordgo.MessageFlagsLoading).
		Build()

	// store the initial response so it can be updated later
	storage.InteractionResponses.Store(token, m.ID)

	_, err := sendMessage(m)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}
