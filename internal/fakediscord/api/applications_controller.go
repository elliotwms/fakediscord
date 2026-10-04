package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/gin-gonic/gin"
)

type commandKey struct {
	guildID, name string
	commandType   discordgo.ApplicationCommandType
}

func toCommandKey(command *discordgo.ApplicationCommand) commandKey {
	return commandKey{
		guildID:     command.GuildID,
		commandType: command.Type,
		name:        command.Name,
	}
}

func applicationsController(r *gin.RouterGroup) {
	r.Use(auth)

	// two sets of routes use the same handlers, as the only distinction is the guild ID used in the map lookup
	r.GET("/:application/commands", getCommands)
	r.PUT("/:application/commands", putCommands)
	r.POST("/:application/commands", postCommand)
	r.GET("/:application/commands/:id", getCommand)
	r.DELETE("/:application/commands/:id", deleteCommand)

	r.GET("/:application/guilds/:guild/commands", getCommands)
	r.PUT("/:application/guilds/:guild/commands", putCommands)
	r.POST("/:application/guilds/:guild/commands", postCommand)
	r.GET("/:application/guilds/:guild/commands/:id", getCommand)
	r.DELETE("/:application/guilds/:guild/commands/:id", deleteCommand)
}

// getCommands returns application/guild commands
// https://discord.com/developers/docs/interactions/application-commands#get-global-application-commands
// https://discord.com/developers/docs/interactions/application-commands#get-guild-application-commands
func getCommands(c *gin.Context) {
	var commands []*discordgo.ApplicationCommand
	storage.Commands.Range(func(k, v interface{}) bool {
		command := v.(*discordgo.ApplicationCommand)
		if inScope(command, c.Param("application"), c.Param("guild")) {
			commands = append(commands, command)
		}

		return true
	})

	c.JSON(http.StatusOK, commands)
}

// putCommands bulk overwrites application/guild commands
// https://discord.com/developers/docs/interactions/application-commands#bulk-overwrite-global-application-commands
// https://discord.com/developers/docs/interactions/application-commands#bulk-overwrite-guild-application-commands
func putCommands(c *gin.Context) {
	appID, guildID := c.Param("application"), c.Param("guild")

	commands := make([]*discordgo.ApplicationCommand, 0)

	if err := c.ShouldBindJSON(&commands); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// clear the existing commands for the application in this scope (global or guild)
	storage.Commands.Range(func(k, v interface{}) bool {
		command := v.(*discordgo.ApplicationCommand)
		if inScope(command, appID, guildID) {
			storage.Commands.Delete(k)
			storage.CommandNames.Delete(toCommandKey(command))
		}

		return true
	})

	for _, command := range commands {
		if command.ID == "" {
			command.ID = snowflake.Generate().String()
		}

		command.ApplicationID = appID
		command.GuildID = guildID

		storage.CommandNames.Store(toCommandKey(command), command.ID)
		storage.Commands.Store(command.ID, command)
	}

	c.JSON(http.StatusOK, commands)
}

// postCommand creates an application/guild command
// https://discord.com/developers/docs/interactions/application-commands#create-global-application-command
// https://discord.com/developers/docs/interactions/application-commands#create-guild-application-command
func postCommand(c *gin.Context) {
	command := &discordgo.ApplicationCommand{
		ID:            snowflake.Generate().String(),
		ApplicationID: c.Param("application"),
		GuildID:       c.Param("guild"),
	}

	if err := c.ShouldBindJSON(command); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	v, loaded := storage.CommandNames.LoadOrStore(toCommandKey(command), command.ID)
	if loaded {
		slog.Info("Replacing existing command", "name", command.Name, "type", command.Type)
		command.ID = v.(string)
	}

	storage.Commands.Store(command.ID, command)

	if loaded {
		c.JSON(http.StatusOK, command)
		return
	}

	c.JSON(http.StatusCreated, command)
}

// getCommand gets a guild/application command
// https://discord.com/developers/docs/interactions/application-commands#get-global-application-command
// https://discord.com/developers/docs/interactions/application-commands#get-guild-application-command
func getCommand(c *gin.Context) {
	command, ok := loadCommand(c)
	if !ok {
		_ = c.AbortWithError(http.StatusNotFound, errors.New("command not found"))
		return
	}

	c.JSON(http.StatusOK, command)
}

// deleteCommand deletes an application/guild command
// https://discord.com/developers/docs/interactions/application-commands#delete-application-application-command
// https://discord.com/developers/docs/interactions/application-commands#delete-guild-application-command
func deleteCommand(c *gin.Context) {
	command, ok := loadCommand(c)
	if !ok {
		_ = c.AbortWithError(http.StatusNotFound, errors.New("command not found"))
		return
	}

	storage.Commands.Delete(command.ID)
	storage.CommandNames.Delete(toCommandKey(command))

	c.Status(http.StatusNoContent)
}

// loadCommand loads the command from the request path, provided it belongs to the application and guild (or global
// scope) in the path
func loadCommand(c *gin.Context) (*discordgo.ApplicationCommand, bool) {
	v, ok := storage.Commands.Load(c.Param("id"))
	if !ok {
		return nil, false
	}

	command := v.(*discordgo.ApplicationCommand)
	if !inScope(command, c.Param("application"), c.Param("guild")) {
		return nil, false
	}

	return command, true
}

// inScope reports whether the command belongs to the application, and to the guild (or is global if guildID is empty)
func inScope(command *discordgo.ApplicationCommand, appID, guildID string) bool {
	return command.ApplicationID == appID && command.GuildID == guildID
}
