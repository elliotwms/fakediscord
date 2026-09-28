package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elliotwms/fakediscord/internal/fakediscord/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGetResponse_InteractionMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	webhooksController(r.Group("webhooks"))

	// a response is stored without its interaction
	token := "orphaned-response"
	storage.InteractionResponses.Store(token, "123")
	t.Cleanup(func() { storage.InteractionResponses.Delete(token) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/webhooks/app/"+token+"/messages/@original", nil)

	require.NotPanics(t, func() { r.ServeHTTP(w, req) })
	require.Equal(t, http.StatusNotFound, w.Code)
}
