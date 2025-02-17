package server

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/liteseed/aogo"
	"github.com/liteseed/goar/wallet"
	"github.com/liteseed/sdk-go/contract"
	"github.com/liteseed/transit/internal/bundler"
	"github.com/liteseed/transit/internal/database/schema"
	"github.com/liteseed/transit/test"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, `{"name":"Transit","version":"test"}`, rcd.Body.String())
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
		assert.Equal(t, `{"code":400,"message":"byte size should be between 1 and 2^32-1"}`, rcd.Body.String())
	})
	t.Run("Fail:Invalid:/price/0", func(t *testing.T) {
		rcd := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/price/0", nil)
		srv.server.Handler.ServeHTTP(rcd, req)
		assert.Equal(t, http.StatusBadRequest, rcd.Code)
		assert.Equal(t, `{"code":400,"message":"byte size should be between 1 and 2^32-1"}`, rcd.Body.String())
	})

	t.Run("Fail:Invalid:/price/-10", func(t *testing.T) {
		rcd := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/price/-10", nil)
		srv.server.Handler.ServeHTTP(rcd, req)
		assert.Equal(t, http.StatusBadRequest, rcd.Code)
		assert.Equal(t, `{"code":400,"message":"byte size should be between 1 and 2^32-1"}`, rcd.Body.String())
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
		assert.Equal(t, `{"code":424,"message":"failed to fetch price"}`, rcd.Body.String())
	})
}

func TestDataItemGet(t *testing.T) {
	d := test.DataItem()
	b := test.Bundler(d)
	defer b.Close()

	mock, db := test.Database()

	srv, _ := New(":8080", "test", WithDatabase(db), WithBundler(bundler.New()))

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"Id", "URL"}).AddRow("1", b.URL[7:]))

		rcd := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/tx/1", nil)
		srv.server.Handler.ServeHTTP(rcd, req)

		assert.Equal(t, http.StatusOK, rcd.Code)
		assert.Equal(t, d.Raw, rcd.Body.Bytes())
	})

	t.Run("Fail:NotFound", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders" WHERE "orders"."id" = $1 ORDER BY "orders"."id" LIMIT $2`)).WithArgs("2", 1).WillReturnError(errors.New("not found"))

		rcd := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/tx/2", nil)
		srv.server.Handler.ServeHTTP(rcd, req)

		assert.Equal(t, http.StatusNotFound, rcd.Code)
		assert.Equal(t, `{"code":404,"message":"not found 2"}`, rcd.Body.String())
	})
}

func TestDataItemPost(t *testing.T) {
	g := test.Gateway()
	w := test.Wallet(g.URL)

	d := test.DataItem()

	b := test.Bundler(d)
	defer b.Close()

	mock, db := test.Database()

	cu := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(fmt.Sprintf(`{"Messages":[{"Data":"{\"id\":\"staker\",\"reputation\":0,\"url\":\"%s\"}"}]}`, b.URL[7:])))
		assert.NoError(t, err)

	}))
	defer cu.Close()

	mu := test.MU()
	defer mu.Close()

	ao, err := aogo.New(aogo.WthCU(cu.URL), aogo.WthMU(mu.URL))
	assert.NoError(t, err)

	c := contract.Custom(ao, "process", w.Signer)

	srv, err := New(":8000", "test", WithBundler(bundler.New()), WithDatabase(db), WithContracts(c), WithWallet(w))
	assert.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "orders" ("id","transaction_id","url","address","status","payment","size") VALUES ($1,$2,$3,$4,$5,$6,$7)`)).WithArgs(d.ID, "", b.URL[7:], "staker", "created", "unpaid", 1047).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		rcd := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/tx", bytes.NewBuffer(d.Raw))
		req.Header.Set("content-type", "application/octet-stream")
		req.Header.Set("content-length", strconv.Itoa(len(d.Raw)))

		srv.server.Handler.ServeHTTP(rcd, req)

		assert.Equal(t, http.StatusCreated, rcd.Code)
		assert.Equal(t, fmt.Sprintf("{\"id\":\"%s\",\"owner\":\"3XTR7MsJUD9LoaiFRdWswzX1X5BR7AQdl1x2v2zIVck\",\"dataCaches\":[\"localhost\"],\"deadlineHeight\":0,\"fastFinalityIndexes\":[\"localhost\"],\"version\":\"1\"}", d.ID), rcd.Body.String())
	})

	t.Run("Missing", func(t *testing.T) {
		rcd := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/tx", bytes.NewBuffer(d.Raw))
		srv.server.Handler.ServeHTTP(rcd, req)
		assert.Equal(t, http.StatusBadRequest, rcd.Code)
		assert.Equal(t, `{"code":400,"message":"required header(s) - content-type, content-length"}`, rcd.Body.String())
	})
	t.Run("Invalid", func(t *testing.T) {
		rcd := httptest.NewRecorder()

		req, _ := http.NewRequest("POST", "/tx", bytes.NewBuffer(d.Raw))
		req.Header.Set("content-type", "application/json")
		req.Header.Set("content-length", strconv.Itoa(len(d.Raw)))
		srv.server.Handler.ServeHTTP(rcd, req)

		assert.Equal(t, http.StatusBadRequest, rcd.Code)
		assert.Equal(t, `{"code":400,"message":"required header(s) - content-type: application/octet-stream"}`, rcd.Body.String())
	})
	t.Run("Invalid Content Type", func(t *testing.T) {
		rcd := httptest.NewRecorder()

		req, _ := http.NewRequest("POST", "/tx", bytes.NewBuffer(d.Raw))
		req.Header.Set("content-type", "application/json")
		req.Header.Set("content-length", strconv.Itoa(len(d.Raw)))
		srv.server.Handler.ServeHTTP(rcd, req)

		assert.Equal(t, http.StatusBadRequest, rcd.Code)
		assert.Equal(t, `{"code":400,"message":"required header(s) - content-type: application/octet-stream"}`, rcd.Body.String())
	})
	t.Run("Invalid Content Length", func(t *testing.T) {
		rcd := httptest.NewRecorder()

		req, _ := http.NewRequest("POST", "/tx", bytes.NewBuffer(d.Raw))
		req.Header.Set("content-type", "application/octet-stream")
		req.Header.Set("content-length", "-100")
		srv.server.Handler.ServeHTTP(rcd, req)

		assert.Equal(t, http.StatusBadRequest, rcd.Code)
		assert.Equal(t, `{"code":400,"message":"content-length, body: length mismatch (-100, 1047)"}`, rcd.Body.String())
	})

	t.Run("Nil Body", func(t *testing.T) {
		rcd := httptest.NewRecorder()

		req, _ := http.NewRequest("POST", "/tx", nil)
		req.Header.Set("content-type", "application/octet-stream")
		req.Header.Set("content-length", strconv.Itoa(len(d.Raw)))
		srv.server.Handler.ServeHTTP(rcd, req)
		assert.Equal(t, http.StatusBadRequest, rcd.Code)
		assert.Equal(t, `{"code":400,"message":"cannot read nil body"}`, rcd.Body.String())
	})

	t.Run("Invalid Body", func(t *testing.T) {
		rcd := httptest.NewRecorder()

		req, _ := http.NewRequest("POST", "/tx", bytes.NewBuffer([]byte{1, 2, 3}))
		req.Header.Set("content-type", "application/octet-stream")
		req.Header.Set("content-length", strconv.Itoa(len(d.Raw)))
		srv.server.Handler.ServeHTTP(rcd, req)
		assert.Equal(t, http.StatusBadRequest, rcd.Code)
		assert.Equal(t, `{"code":400,"message":"content-length, body: length mismatch (1047, 3)"}`, rcd.Body.String())
	})

	t.Run("Invalid Data Item", func(t *testing.T) {
		rcd := httptest.NewRecorder()

		req, _ := http.NewRequest("POST", "/tx", bytes.NewBuffer([]byte{1, 2, 3}))
		req.Header.Set("content-type", "application/octet-stream")
		req.Header.Set("content-length", "3")
		srv.server.Handler.ServeHTTP(rcd, req)
		assert.Equal(t, http.StatusBadRequest, rcd.Code)
		assert.Equal(t, `{"code":400,"message":"failed to decode data item"}`, rcd.Body.String())

	})
}

func TestDataItemPut(t *testing.T) {
	mock, db := test.Database()
	srv, err := New(":8000", "test", WithDatabase(db))
	assert.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "orders" SET "transaction_id"=$1,"status"=$2 WHERE id = $3`)).WithArgs("transaction", schema.Queued, "dataitem").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		rcd := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/tx/dataitem/transaction", nil)

		srv.server.Handler.ServeHTTP(rcd, req)

		assert.Equal(t, http.StatusAccepted, rcd.Code)
		assert.Equal(t, "{\"id\":\"dataitem\",\"paymentId\":\"transaction\"}", rcd.Body.String())
	})
}
