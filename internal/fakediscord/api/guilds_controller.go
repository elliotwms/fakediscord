package api

import (
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/builders"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/fakediscord/ws"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/gin-gonic/gin"
)

func guildsController(r *gin.RouterGroup) {
	r.POST("", postGuild)
	r.DELETE("/:guild", deleteGuild)

	r.GET("/:guild/channels", getGuildChannels)
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

	guild := builders.NewGuild(data.Name).Build()

	storage.Guilds.Store(guild.ID, *guild)

	_ = ws.DispatchEvent("GUILD_CREATE", discordgo.GuildCreate{
		Guild: guild,
	})

	c.JSON(http.StatusCreated, guild)
}

// https://discord.com/developers/docs/resources/guild#delete-guild
func deleteGuild(c *gin.Context) {
	v, ok := storage.Guilds.LoadAndDelete(c.Param("guild"))
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	guild := v.(discordgo.Guild)

	if err := ws.DispatchEvent("GUILD_UPDATE", guild); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if err := ws.DispatchEvent("GUILD_DELETE", discordgo.Guild{ID: guild.ID}); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// https://discord.com/developers/docs/resources/guild#get-guild-channels
func getGuildChannels(c *gin.Context) {
	channels := []discordgo.Channel{}

	storage.Channels.Range(func(k, v interface{}) bool {
		channel := v.(discordgo.Channel)
		if channel.GuildID == c.Param("guild") {
			channels = append(channels, channel)
		}

		return true
	})

	c.JSON(http.StatusOK, channels)
}

// https://discord.com/developers/docs/resources/guild#create-guild-channel
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
