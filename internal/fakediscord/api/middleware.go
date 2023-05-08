package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/gin-gonic/gin"
)

const contextKeyUser = "user"

func auth(c *gin.Context) {
	split := strings.SplitN(c.GetHeader("Authorization"), " ", 2)
	if len(split) != 2 {
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("invalid Authorization header"))
		return
	}

	u := storage.Authenticate(split[1])

	if u == nil {
		_ = c.AbortWithError(http.StatusUnauthorized, errors.New("token not found"))
		return
	}

	c.Set(contextKeyUser, *u)
}
