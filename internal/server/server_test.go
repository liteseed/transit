package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/liteseed/goar/tag"
	"github.com/liteseed/goar/wallet"
	"github.com/liteseed/sdk-go/contract"
	"github.com/liteseed/transit/internal/database"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	server, err := New(":8080", "test")
	assert.NoError(t, err)
	assert.NotNil(t, server)
}

func TestStatusHandler(t *testing.T) {
	server, _ := New(":8080", "test")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"status":"ok"}`)
}

func TestPriceGetHandler(t *testing.T) {
	server, _ := New(":8080", "test")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/price/1024", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"price":`)
}

func TestPriceGetHandler_Error(t *testing.T) {
	server, _ := New(":8080", "test")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/price/invalid", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"invalid byte size"}`)
}

func TestPriceGetHandler_ZeroByteSize(t *testing.T) {
	server, _ := New(":8080", "test")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/price/0", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"byte size must be greater than zero"}`)
}

func TestPriceGetHandler_NegativeByteSize(t *testing.T) {
	server, _ := New(":8080", "test")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/price/-1024", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"invalid byte size"}`)
}

func TestDataPostHandler(t *testing.T) {
	server, _ := New(":8080", "test")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"success":true}`)
}

func TestDataPostHandler_Error(t *testing.T) {
	server, _ := New(":8080", "test")

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

	server, _ := New(":8080", "test")
	w := httptest.NewRecorder()

	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"success":true}`)
}

func TestDataPostHandler_LargeData(t *testing.T) {
	server, _ := New(":8080", "test")

	largeData := make([]byte, MAX_DATA_SIZE+1)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", bytes.NewBuffer(largeData))
	req.Header.Set("Content-Type", "application/octet-stream")
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"data size exceeds limit"}`)
}

func TestDataItemGetHandler(t *testing.T) {
	server, _ := New(":8080", "test")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tx/12345", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"transaction":}`)
}

func TestDataItemGetHandler_Error(t *testing.T) {
	server, _ := New(":8080", "test")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tx/nonexistent", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"transaction not found"}`)
}

func TestDataItemPostHandler(t *testing.T) {

	db, err := database.New("postgresql://localhost:5432/postgres")
	assert.NoError(t, err)

	err = db.Migrate()
	assert.NoError(t, err)

	w, err := wallet.New("http://localhost:1984")
	assert.NoError(t, err)

	c := contract.New("PWSr59Cf6jxY7aA_cfz69rs0IiJWWbmQA8bAKknHeMo", w.Signer)

	srv, err := New(":8000", "test", WithDatabase(db), WithContracts(c), WithWallet(w))
	assert.NoError(t, err)

	rec := httptest.NewRecorder()

	t.Run("Success", func(t *testing.T) {
		d := w.CreateDataItem([]byte{1, 2, 3}, "", "", []tag.Tag{})
		_, err = w.SignDataItem(d)
		req, _ := http.NewRequest("POST", "/tx", bytes.NewBuffer(d.Raw))
		req.Header.Set("content-type", "application/octet-stream")
		req.Header.Set("content-length", strconv.Itoa(len(d.Raw)))

		srv.server.Handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `{"success":true}`)
	})
}
