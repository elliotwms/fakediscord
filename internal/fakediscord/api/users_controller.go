package api

import (
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/gin-gonic/gin"
)

func usersController(r *gin.RouterGroup) {
	r.Use(auth)
	r.GET("/:id/guilds", getUserGuilds)
}

// https://discord.com/developers/docs/resources/user#get-current-user-guilds
func getUserGuilds(c *gin.Context) {
	storage.State.RLock()
	defer storage.State.RUnlock()

	guilds := make([]*discordgo.UserGuild, len(storage.State.Guilds))

	for _, g := range storage.State.Guilds {
		guilds = append(guilds, &discordgo.UserGuild{
			ID:       g.ID,
			Name:     g.Name,
			Icon:     g.Icon,
			Features: g.Features,
		})
	}

	c.JSON(http.StatusOK, guilds)
}
