package api

import (
	"errors"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/internal/fakediscord/builders"
	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/elliotwms/fakediscord/internal/fakediscord/ws"
	"github.com/elliotwms/fakediscord/internal/snowflake"
	"github.com/gin-gonic/gin"
)

func guildsController(r *gin.RouterGroup) {
	r.Use(auth)
	r.POST("", postGuild)
	r.GET("/:guild", getGuild)
	r.DELETE("/:guild", deleteGuild)

	r.GET("/:guild/channels", getGuildChannels)
	r.POST("/:guild/channels", postGuildChannels)
}

// https://discord.com/developers/docs/resources/guild#create-guild
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

	err = storage.State.GuildAdd(guild)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	_ = ws.DispatchEvent("GUILD_CREATE", discordgo.GuildCreate{
		Guild: guild,
	})

	c.JSON(http.StatusCreated, guild)
}

// https://discord.com/developers/docs/resources/guild#get-guild
func getGuild(c *gin.Context) {
	g, err := storage.State.Guild(c.Param("guild"))
	if err != nil {
		if errors.Is(err, discordgo.ErrStateNotFound) {
			c.Status(http.StatusNotFound)
			return
		}

		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, g)
}

// https://discord.com/developers/docs/resources/guild#delete-guild
func deleteGuild(c *gin.Context) {
	guild := &discordgo.Guild{ID: c.Param("guild")}
	err := storage.State.GuildRemove(guild)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	if err := ws.DispatchEvent("GUILD_DELETE", guild); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// https://discord.com/developers/docs/resources/guild#get-guild-channels
func getGuildChannels(c *gin.Context) {
	guild, err := storage.State.Guild(c.Param("guild"))
	if err != nil {
		handleStateErr(c, err)
		return
	}

	c.JSON(http.StatusOK, guild.Channels)
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

	err = storage.State.ChannelAdd(&channel)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = ws.DispatchEvent("CHANNEL_CREATE", discordgo.ChannelCreate{
		Channel: &channel,
	})

	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, channel)
}
