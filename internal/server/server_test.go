package server

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/liteseed/sdk-go/contract"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock dependencies
type MockContract struct {
	mock.Mock
}

func (m *MockContract) Balance(target string) (string, error) {
	args := m.Called(target)
	return args.String(0), args.Error(1)
}

func (m *MockContract) Balances() (*map[string]string, error) {
	args := m.Called()
	return args.Get(0).(*map[string]string), args.Error(1)
}

func (m *MockContract) Initiate(dataItemId string, size int) (*contract.Staker, error) {
	args := m.Called(dataItemId, size)
	return args.Get(0).(*contract.Staker), args.Error(1)
}

func (m *MockContract) Posted(dataItemId string, transactionId string) error {
	args := m.Called(dataItemId, transactionId)
	return args.Error(0)
}

func (m *MockContract) Release(dataItemId string, transactionId string) error {
	args := m.Called(dataItemId, transactionId)
	return args.Error(0)
}

func (m *MockContract) Stake(url string) (string, error) {
	args := m.Called(url)
	return args.String(0), args.Error(1)
}

func (m *MockContract) Staked() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockContract) Stakers() (*[]contract.Staker, error) {
	args := m.Called()
	return args.Get(0).(*[]contract.Staker), args.Error(1)
}

func (m *MockContract) Transfer(recipient string, quantity string) error {
	args := m.Called(recipient, quantity)
	return args.Error(0)
}

func (m *MockContract) Unstake() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockContract) Upload(id string) (*contract.Upload, error) {
	args := m.Called(id)
	return args.Get(0).(*contract.Upload), args.Error(1)
}

type MockDatabase struct {
	mock.Mock
}

func TestNewServer(t *testing.T) {
	server, err := New(":8000", "v1", "http://localhost:1984", WithContracts(&MockContract{}), WithDatabase(&MockDatabase{}))
	assert.NoError(t, err)
	assert.NotNil(t, server)
}

func TestStatusHandler(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984", WithContracts(&MockContract{}), WithDatabase(&MockDatabase{}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"status":"ok"}`)
}

func TestPriceGetHandler(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984", WithContracts(&MockContract{}), WithDatabase(&MockDatabase{}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/price/1024", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"price":`)
}

func TestPriceGetHandler_LargeInput(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984", WithContracts(&MockContract{}), WithDatabase(&MockDatabase{}))

	// Test with byte size exceeding the limit
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/price/2147483648", nil) // 2GB + 1 byte
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"byte size exceeds limit"}`)
}

func TestDataPostHandler(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984", WithContracts(&MockContract{}), WithDatabase(&MockDatabase{}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"success":true}`)
}

func TestDataPostHandler_EmptyBody(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984", WithContracts(&MockContract{}), WithDatabase(&MockDatabase{}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", bytes.NewBuffer([]byte{}))
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"invalid data"}`)
}

func TestDataPostHandler_InvalidContentType(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984", WithContracts(&MockContract{}), WithDatabase(&MockDatabase{}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", bytes.NewBuffer([]byte("data")))
	req.Header.Set("Content-Type", "text/plain")
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"unsupported media type"}`)
}

func TestDataPostHandler_MaxDataSize(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984", WithContracts(&MockContract{}), WithDatabase(&MockDatabase{}))

	// Test with content exactly at the maximum allowed size
	w := httptest.NewRecorder()
	data := make([]byte, MAX_DATA_SIZE)
	req, _ := http.NewRequest("POST", "/upload", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", CONTENT_TYPE_OCTET_STREAM)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"success":true}`)
}

func TestDataPostHandler_ExceedMaxDataSize(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984", WithContracts(&MockContract{}), WithDatabase(&MockDatabase{}))

	// Test with content exceeding the maximum allowed size
	w := httptest.NewRecorder()
	data := make([]byte, MAX_DATA_SIZE+1)
	req, _ := http.NewRequest("POST", "/upload", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", CONTENT_TYPE_OCTET_STREAM)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"data size exceeds limit"}`)
}

func TestTransactionGetHandler(t *testing.T) {
	mockContract := new(MockContract)
	mockContract.On("Balance", "12345").Return("1000", nil)
	server, _ := New(":8080", "v1", "http://localhost:1984", WithContracts(mockContract), WithDatabase(&MockDatabase{}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tx/12345", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"transaction":"1000"}`)
}

func TestTransactionGetHandler_NonExistent(t *testing.T) {
	mockContract := new(MockContract)
	mockContract.On("Balance", "nonexistent").Return("", errors.New("transaction id not found"))
	server, _ := New(":8080", "v1", "http://localhost:1984", WithContracts(mockContract), WithDatabase(&MockDatabase{}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tx/nonexistent", nil)
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"transaction id not found"}`)
}

func TestTransactionPostHandler(t *testing.T) {
	mockContract := new(MockContract)
	mockContract.On("Transfer", "recipientAddress", "100").Return(nil)
	server, _ := New(":8080", "v1", "http://localhost:1984", WithContracts(mockContract), WithDatabase(&MockDatabase{}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tx", bytes.NewBuffer([]byte(`{"recipient":"recipientAddress","quantity":"100"}`)))
	req.Header.Set("Content-Type", "application/json")
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"success":true}`)
}

func TestTransactionPostHandler_Error(t *testing.T) {
	mockContract := new(MockContract)
	mockContract.On("Transfer", "recipientAddress", "100").Return(errors.New("insufficient funds"))
	server, _ := New(":8080", "v1", "http://localhost:1984", WithContracts(mockContract), WithDatabase(&MockDatabase{}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tx", bytes.NewBuffer([]byte(`{"recipient":"recipientAddress","quantity":"100"}`)))
	req.Header.Set("Content-Type", "application/json")
	server.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"insufficient funds"}`)
}

func TestUnsupportedMethods(t *testing.T) {
	server, _ := New(":8080", "v1", "http://localhost:1984", WithContracts(&MockContract{}), WithDatabase(&MockDatabase{}))

	unsupportedEndpoints := []string{"/", "/price/1024", "/upload", "/tx/12345", "/tx"}
	methods := []string{"PUT", "DELETE"}

	for _, endpoint := range unsupportedEndpoints {
		for _, method := range methods {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, endpoint, nil)
			server.server.Handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
			assert.Contains(t, w.Body.String(), `{"error":"method not allowed"}`)
		}
	}
}
