package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	server, err := New(":8000", "v1", "http://localhost:1984")
	assert.NoError(t, err)
	assert.NotNil(t, server)
}

func TestStatusHandeler(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"status":"ok"}`)
}
