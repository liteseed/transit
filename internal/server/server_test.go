package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	server, err := New(":8080", "v1", "http://localhost:1984")
	assert.NoError(t, err)
	assert.NotNil(t, server)
}

func TestStatusHandler(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"status":"ok"}`)
}

func TestPriceGetHandler(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/price/1024", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"price":`)
}

func TestPriceGetHandler_Error(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/price/invalid", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"invalid byte size"}`)
}

func TestPriceGetHandler_ZeroByteSize(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/price/0", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"byte size must be greater than zero"}`)
}

func TestPriceGetHandler_NegativeByteSize(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/price/-1024", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"invalid byte size"}`)
}

func TestDataPostHandler(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"success":true}`)
}

func TestDataPostHandler_Error(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"invalid data"}`)
}

func TestDataPostHandler_WithData(t *testing.T) {
	// Simulate valid data
	data := []byte("test data")
	req, err := http.NewRequest("POST", "/upload", bytes.NewReader(data))
	assert.NoError(t, err)

	server, _ := New(":8080", "v1", "http://localhost:1984")
	w := httptest.NewRecorder()

	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"success":true}`)
}

func TestDataPostHandler_LargeData(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984")

	largeData := make([]byte, MAX_DATA_SIZE+1)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", bytes.NewBuffer(largeData))
	req.Header.Set("Content-Type", "application/octet-stream")
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"data size exceeds limit"}`)
}

func TestTransactionGetHandler(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tx/12345", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"transaction":}`)
}

func TestTransactionGetHandler_Error(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tx/nonexistent", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"transaction not found"}`)
}

func TestTransactionPostHandler(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tx", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"success":true}`)
}

func TestTransactionPostHandler_Error(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tx", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"missing data"}`)
}

func TestTransactionPostHandler_MissingFields(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984")

	requestBody := `{"recipient":"recipientAddress"}`
	req, err := http.NewRequest("POST", "/tx", bytes.NewBuffer([]byte(requestBody)))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	assert.Contains(t, w.Body.String(), `{"error":"missing required fields"}`)
}

func TestTransactionPostHandler_InvalidJSON(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tx", bytes.NewBuffer([]byte(`{invalid json}`)))
	req.Header.Set("Content-Type", "application/json")
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"invalid JSON"}`)
}
