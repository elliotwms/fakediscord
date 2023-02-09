package api

import (
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/fakediscord/ws"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/elliotwms/fakediscord/pkg/config"
	"github.com/gin-gonic/gin"
)

func guildsController(r *gin.RouterGroup) {
	r.POST("", postGuild)

	r.POST("/:guild/channels", postGuildChannels)
}

func postGuild(c *gin.Context) {
	data := struct {
		Name string `json:"name"`
	}{}

	err := c.BindJSON(&data)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	guild := storage.BuildTestGuild(config.Guild{
		Name: data.Name,
	})

	storage.Guilds.Store(guild.ID, guild)

	_ = ws.DispatchEvent("GUILD_CREATE", discordgo.GuildCreate{
		Guild: &guild,
	})

	c.JSON(http.StatusCreated, guild)
}

func postGuildChannels(c *gin.Context) {
	channel := discordgo.Channel{}

	err := c.BindJSON(&channel)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	}

	channel.ID = snowflake.Generate().String()
	channel.GuildID = c.Param("guild")

	storage.Channels.Store(channel.ID, channel)

	err = ws.DispatchEvent("CHANNEL_CREATE", discordgo.ChannelCreate{
		Channel: &channel,
	})

	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, channel)
}