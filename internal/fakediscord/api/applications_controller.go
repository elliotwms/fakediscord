package api

import (
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
)

func applicationsController(r *gin.RouterGroup) {
	r.POST("/:application/guilds/:guild/commands", createGuildApplicationCommands)
}

// https://discord.com/developers/docs/interactions/application-commands#create-guild-application-command
func createGuildApplicationCommands(c *gin.Context) {
	cmd := &discordgo.ApplicationCommand{}
	if err := c.BindJSON(cmd); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// todo persist commands

	// todo 200 if command already exists
	c.JSON(http.StatusCreated, cmd)
}
