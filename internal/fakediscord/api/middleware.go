package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	pkgauth "github.com/elliotwms/fakediscord/internal/fakediscord/auth"
	"github.com/gin-gonic/gin"
)

const contextKeyUserID = "user_id"

func auth(c *gin.Context) {
	split := strings.SplitN(c.GetHeader("Authorization"), " ", 2)
	if len(split) != 2 {
		_ = c.Error(errors.New("invalid Authorization header"))
		c.AbortWithStatusJSON(http.StatusUnauthorized, discordgo.APIErrorMessage{
			Message: "401: Unauthorized",
		})
		return
	}

	c.Set(contextKeyUserID, pkgauth.Authenticate(split[1]).ID)
}
