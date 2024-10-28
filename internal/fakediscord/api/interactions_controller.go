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
	u, done := getUser(c)
	if done {
		return
	}

	interaction := &discordgo.Interaction{}
	if err := c.BindJSON(interaction); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setInteractionDefaults(interaction, u)

	if err := validateInteraction(interaction); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	storage.Interactions.Store(interaction.Token, *interaction)

	// todo send to webhook (query param?)

	// if webhook not registered, send interaction via connection
	_, err := ws.Connections.Send(interaction.AppID, "INTERACTION_CREATE", interaction)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, interaction)
}

// setInteractionDefaults sets some default values when creating a new interaction
func setInteractionDefaults(interaction *discordgo.Interaction, u discordgo.User) {
	if interaction.ID == "" {
		interaction.ID = snowflake.Generate().String()
	}

	if interaction.Token == "" {
		// token appears to be just random bytes
		interaction.Token = base64.StdEncoding.EncodeToString([]byte(
			fmt.Sprintf("interaction:%s:%s", interaction.ID, snowflake.Generate().String()),
		))
	}

	if interaction.AppID == "" {
		interaction.AppID = u.ID
	}
}

func validateInteraction(interaction *discordgo.Interaction) error {
	var errs []error

	if interaction.ID == "" {
		errs = append(errs, errors.New("missing id"))
	}

	if interaction.AppID == "" {
		errs = append(errs, errors.New("missing app_id"))
	}

	if interaction.Type == 0 {
		errs = append(errs, errors.New("missing type"))
	} else {
		switch interaction.Type {
		case discordgo.InteractionApplicationCommand, discordgo.InteractionApplicationCommandAutocomplete:
			errs = validateApplicationCommandData(interaction, errs)
		case discordgo.InteractionMessageComponent:
			errs = validateMessageComponentData(interaction, errs)
		case discordgo.InteractionModalSubmit:
			errs = validateModalSubmitData(interaction, errs)
		}
	}

	if interaction.GuildID == "" {
		errs = append(errs, errors.New("missing guild_id"))
	}

	if interaction.ChannelID == "" {
		errs = append(errs, errors.New("missing channel_id"))
	}

	return errors.Join(errs...)
}

func validateModalSubmitData(interaction *discordgo.Interaction, errs []error) []error {
	data := interaction.ModalSubmitData()
	if data.CustomID == "" {
		errs = append(errs, errors.New("missing data.custom_id"))
	}
	return errs
}

func validateMessageComponentData(interaction *discordgo.Interaction, errs []error) []error {
	data := interaction.MessageComponentData()
	if data.CustomID == "" {
		errs = append(errs, errors.New("missing data.custom_id"))
	}

	if data.ComponentType == 0 {
		errs = append(errs, errors.New("missing data.component_type"))
	}
	return errs
}

func validateApplicationCommandData(interaction *discordgo.Interaction, errs []error) []error {
	data := interaction.ApplicationCommandData()
	if data.ID == "" {
		errs = append(errs, errors.New("missing data.id"))
	}

	if data.Name == "" {
		errs = append(errs, errors.New("missing data.name"))
	}

	if data.CommandType == 0 {
		errs = append(errs, errors.New("missing data.command_type"))
	}
	return errs
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
