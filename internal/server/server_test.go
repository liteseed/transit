package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	server, err := New(":8000", "v1", "http://localhost:1984")
	assert.NoError(t, err)
	assert.NotNil(t, server)
}
