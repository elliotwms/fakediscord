package api

import (
	"encoding/base64"
	"errors"
	"fmt"
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

	id := snowflake.Generate().String()

	// token appears to be just random bytes
	token := base64.StdEncoding.EncodeToString([]byte(
		fmt.Sprintf("interaction:%s:%s", id, snowflake.Generate().String()),
	))

	// generate ID and token
	interaction.ID = id
	interaction.Token = token
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

	h, ok := interactionHandlers[res.Type]
	if !ok {
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}

	code, err := h(&i, res)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(code)
}

type interactionHandler func(i *discordgo.Interaction, res *discordgo.InteractionResponse) (statusCode int, err error)

var interactionHandlers = map[discordgo.InteractionResponseType]interactionHandler{
	discordgo.InteractionResponsePong:                             handlePong,
	discordgo.InteractionResponseChannelMessageWithSource:         handleMessageInteractionResponse,
	discordgo.InteractionResponseDeferredChannelMessageWithSource: emitLoadingMessage,
}

func handlePong(*discordgo.Interaction, *discordgo.InteractionResponse) (int, error) {
	// responding to pong via callback does not work 🤷
	return http.StatusNotFound, nil
}

func handleMessageInteractionResponse(i *discordgo.Interaction, res *discordgo.InteractionResponse) (statusCode int, err error) {
	m := builders.NewMessage(i.User, i.ChannelID, i.GuildID).
		WithType(discordgo.MessageTypeReply).
		WithContent(res.Data.Content).
		WithEmbeds(res.Data.Embeds).
		WithComponents(res.Data.Components).
		WithFlags(0). // remove loading flag if exists
		Build()

	storage.InteractionResponses.Store(i.Token, m.ID)

	_, err = sendMessage(m)
	if err != nil {
		return
	}

	return http.StatusNoContent, nil
}

func emitLoadingMessage(i *discordgo.Interaction, _ *discordgo.InteractionResponse) (statusCode int, err error) {
	v, ok := storage.Users.Load(i.AppID)
	if !ok {
		return http.StatusNotFound, nil
	}
	u := v.(discordgo.User)

	m := builders.NewMessage(&u, i.ChannelID, i.GuildID).
		WithType(discordgo.MessageTypeReply).
		WithFlags(discordgo.MessageFlagsLoading).
		Build()

	// store the initial response so it can be updated later
	storage.InteractionResponses.Store(i.Token, m.ID)

	_, err = sendMessage(m)
	if err != nil {
		return
	}

	return http.StatusNoContent, nil
}
