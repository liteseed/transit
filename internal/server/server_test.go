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
	// Assuming your PriceGet handler checks for invalid byte size and returns an error
	server, _ := New(":8080", "v1", "http://localhost:1984")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/price/invalid", nil)
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
	//Error in DataPost handler due to invalid data
	server, _ := New(":8080", "v1", "http://localhost:1984")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"invalid data"}`)
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
	//Error in TransactionGet handler due to transaction not existing (maybe.?)
	server, _ := New(":8080", "v1", "http://localhost:1984")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tx/nonexistent", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"transaction id not found"}`)
}
