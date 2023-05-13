package api

import (
	"errors"
	"net/http"
	"strings"

	pkgauth "github.com/elliotwms/fakediscord/internal/fakediscord/auth"
	"github.com/gin-gonic/gin"
)

const contextKeyUserID = "user_id"

func auth(c *gin.Context) {
	split := strings.SplitN(c.GetHeader("Authorization"), " ", 2)
	if len(split) != 2 {
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("invalid Authorization header"))
		return
	}

	c.Set(contextKeyUserID, pkgauth.Authenticate(split[1]).ID)
}
