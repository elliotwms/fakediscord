package api

import (
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
)

func usersController(r *gin.RouterGroup) {
	r.GET("/:id/guilds", getUserGuilds)
}

func getUserGuilds(c *gin.Context) {
	// todo return list of guilds in storage
	c.JSON(http.StatusOK, []*discordgo.UserGuild{})
}
