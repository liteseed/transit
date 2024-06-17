package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseURL(t *testing.T) {
	url, err := ParseUrl("localhost:8080")
	assert.NoError(t, err)

	assert.Equal(t, "http://localhost:8080", url)

	url, err = ParseUrl("test.com")
	assert.NoError(t, err)

	assert.Equal(t, "http://test.com", url)
}
