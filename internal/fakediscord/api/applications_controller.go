package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/gin-gonic/gin"
)

func applicationsController(r *gin.RouterGroup) {
	r.Use(auth)

	r.POST("/:application/guilds/:guild/commands", postGuildCommand)
	r.GET("/:application/guilds/:guild/commands/:id", getGuildCommand)
	r.DELETE("/:application/guilds/:guild/commands/:id", deleteGuildCommand)
}

// postGuildCommand creates a guild application command
// https://discord.com/developers/docs/interactions/application-commands#create-guild-application-command
func postGuildCommand(c *gin.Context) {
	guildID := c.Param("guild")

	u, done := getUser(c)
	if done {
		return
	}

	command := &discordgo.ApplicationCommand{
		ID:            snowflake.Generate().String(),
		ApplicationID: u.ID,
		GuildID:       guildID,
	}

	if err := c.BindJSON(command); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	v, loaded := storage.CommandNames.LoadOrStore(commandKey(command), command.ID)
	if loaded {
		slog.Info("Replacing existing command", "name", command.Name, "type", command.Type)
		command.ID = v.(string)
	}

	storage.Commands.Store(command.ID, *command)

	if loaded {
		c.JSON(http.StatusOK, command)
		return
	}

	c.JSON(http.StatusCreated, command)
}

func commandKey(command *discordgo.ApplicationCommand) string {
	return fmt.Sprintf("%s:%d:%s", command.GuildID, command.Type, command.Name)
}

// getGuildCommand gets a guild application command
// https://discord.com/developers/docs/interactions/application-commands#get-guild-application-command
func getGuildCommand(c *gin.Context) {
	id := c.Param("id")
	v, ok := storage.Commands.Load(id)
	if !ok {
		_ = c.AbortWithError(http.StatusNotFound, errors.New("command not found"))
		return
	}

	c.JSON(http.StatusOK, v)
}

// deleteGuildCommand deletes a guild command
// https://discord.com/developers/docs/interactions/application-commands#delete-guild-application-command
func deleteGuildCommand(c *gin.Context) {
	id := c.Param("id")
	_, ok := storage.Commands.LoadAndDelete(id)
	if !ok {
		_ = c.AbortWithError(http.StatusNotFound, errors.New("command not found"))
		return
	}

	c.Status(http.StatusNoContent)
}
