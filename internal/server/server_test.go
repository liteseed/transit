package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/liteseed/goar/tag"
	"github.com/liteseed/goar/wallet"
	"github.com/liteseed/sdk-go/contract"
	"github.com/liteseed/transit/internal/bundler"
	"github.com/liteseed/transit/internal/database"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
)

func TestNewServer(t *testing.T) {
	srv, err := New(":8080", "test")
	assert.NoError(t, err)
	assert.NotNil(t, srv)
}

func TestStatusHandler(t *testing.T) {
	srv, _ := New(":8080", "test")

	rcd := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	srv.server.Handler.ServeHTTP(rcd, req)

	assert.Equal(t, http.StatusOK, rcd.Code)
	assert.Equal(t, `{"Name":"Transit","Version":"test"}`, rcd.Body.String())
}

func TestPriceGet(t *testing.T) {
	arweave := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("1000"))
			assert.NoError(t, err)
		}))

	defer arweave.Close()

	w, err := wallet.FromPath("../../test/signer.json", arweave.URL)
	assert.NoError(t, err)

	srv, err := New(":8080", "test", WithWallet(w))
	assert.NoError(t, err)

	t.Run("Success:/price/1000", func(t *testing.T) {
		rcd := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/price/1000", nil)
		srv.server.Handler.ServeHTTP(rcd, req)
		assert.Equal(t, http.StatusOK, rcd.Code)
		assert.Equal(t, `{"price":"1001","address":"3XTR7MsJUD9LoaiFRdWswzX1X5BR7AQdl1x2v2zIVck"}`, rcd.Body.String())
	})

	t.Run("Fail:Invalid:/price/invalid", func(t *testing.T) {
		rcd := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/price/invalid", nil)
		srv.server.Handler.ServeHTTP(rcd, req)
		assert.Equal(t, http.StatusBadRequest, rcd.Code)
		assert.Equal(t, `{"error":"size should be between 1 and 2^32-1"}`, rcd.Body.String())
	})
	t.Run("Fail:Invalid:/price/0", func(t *testing.T) {
		rcd := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/price/0", nil)
		srv.server.Handler.ServeHTTP(rcd, req)
		assert.Equal(t, http.StatusBadRequest, rcd.Code)
		assert.Equal(t, `{"error":"size should be between 1 and 2^32-1"}`, rcd.Body.String())
	})

	t.Run("Fail:Invalid:/price/-10", func(t *testing.T) {
		rcd := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/price/-10", nil)
		srv.server.Handler.ServeHTTP(rcd, req)
		assert.Equal(t, http.StatusBadRequest, rcd.Code)
		assert.Equal(t, `{"error":"size should be between 1 and 2^32-1"}`, rcd.Body.String())
	})
	t.Run("Fail:Gateway", func(t *testing.T) {
		w, err := wallet.FromPath("../../test/signer.json", "")
		assert.NoError(t, err)
		srv, err := New(":8080", "test", WithWallet(w))
		assert.NoError(t, err)
		rcd := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/price/1024", nil)
		srv.server.Handler.ServeHTTP(rcd, req)
		assert.Equal(t, http.StatusFailedDependency, rcd.Code)
		assert.Equal(t, `{"error":"failed to fetch price"}`, rcd.Body.String())
	})
}

func TestDataItemGet(t *testing.T) {
	d, err := os.ReadFile("../../test/dataitem")
	assert.NoError(t, err)
	b := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(d)
			assert.NoError(t, err)
		}))
	defer b.Close()

	mockDb, mock, _ := sqlmock.New()
	db, err := database.FromDialector(postgres.New(postgres.Config{
		Conn:       mockDb,
		DriverName: "postgres",
	}))
	assert.NoError(t, err)

	srv, _ := New(":8080", "test", WithDatabase(db), WithBundler(bundler.New()))

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"ID", "URL"}).AddRow("1", b.URL[7:])
		mock.ExpectQuery("SELECT").WillReturnRows(rows)
		rcd := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/tx/1", nil)
		srv.server.Handler.ServeHTTP(rcd, req)

		assert.Equal(t, http.StatusOK, rcd.Code)
		assert.Equal(t, rcd.Body.Bytes(), d)
	})

	t.Run("Fail:NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT")
		rcd := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/tx/2", nil)
		srv.server.Handler.ServeHTTP(rcd, req)

		assert.Equal(t, http.StatusNotFound, rcd.Code)
		assert.Equal(t, rcd.Body.Bytes(), nil)
	})
}

func TestDataItemPostHandler(t *testing.T) {
	b := bundler.New()

	db, err := database.New("postgresql://localhost:5432/postgres")
	assert.NoError(t, err)

	err = db.Migrate()
	assert.NoError(t, err)

	w, err := wallet.New("http://localhost:1984")
	assert.NoError(t, err)

	c := contract.New("PWSr59Cf6jxY7aA_cfz69rs0IiJWWbmQA8bAKknHeMo", w.Signer)

	srv, err := New(":8000", "test", WithBundler(b), WithDatabase(db), WithContracts(c), WithWallet(w))
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
