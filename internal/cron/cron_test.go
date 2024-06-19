package cron

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/liteseed/goar/wallet"
	"github.com/liteseed/transit/internal/database"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
)

func TestCheckPaymentsAmount(t *testing.T) {
	mockDb, mock, _ := sqlmock.New()
	db, err := database.FromDialector(postgres.New(postgres.Config{
		Conn:       mockDb,
		DriverName: "postgres",
	}))

	assert.NoError(t, err)
	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"ID", "TransactionID", "Payment", "Size"}).AddRow("dataitem", "transaction", "unpaid", 1000))
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "orders" SET "payment"=$1 WHERE id = $2`)).WithArgs("paid", "dataitem").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		arweave := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/price/1000" {
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte("10000"))
					assert.NoError(t, err)
				} else {
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(`{"id":"transaction","quantity":"100100","target":"3XTR7MsJUD9LoaiFRdWswzX1X5BR7AQdl1x2v2zIVck"}`))
					assert.NoError(t, err)
				}
			}))

		defer arweave.Close()

		w, err := wallet.FromPath("../../test/signer.json", arweave.URL)
		assert.NoError(t, err)
		crn, err := New(WithDatabase(db), WithWallet(w))
		assert.NoError(t, err)

		crn.CheckPaymentsAmount()
		assert.NoError(t, mock.ExpectationsWereMet())

	})

	t.Run("Not Enough Fee", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"ID", "TransactionID", "Payment", "Size"}).AddRow("dataitem", "transaction", "unpaid", 1000))
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "orders" SET "status"=$1,"payment"=$2 WHERE id = $3`)).WithArgs("failed", "invalid", "dataitem").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		arweave := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/price/1000" {
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte("10000"))
					assert.NoError(t, err)
				} else {
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(`{"id":"transaction","quantity":"10000","target":"3XTR7MsJUD9LoaiFRdWswzX1X5BR7AQdl1x2v2zIVck"}`))
					assert.NoError(t, err)
				}
			}))

		defer arweave.Close()

		w, err := wallet.FromPath("../../test/signer.json", arweave.URL)
		assert.NoError(t, err)
		crn, err := New(WithDatabase(db), WithWallet(w))
		assert.NoError(t, err)

		crn.CheckPaymentsAmount()
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"ID", "TransactionID", "Payment", "Size"}).AddRow("dataitem", "transaction", "unpaid", 1000))
		arweave := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}))

		defer arweave.Close()

		w, err := wallet.FromPath("../../test/signer.json", arweave.URL)
		assert.NoError(t, err)
		crn, err := New(WithDatabase(db), WithLogger(slog.Default()), WithWallet(w))
		assert.NoError(t, err)

		crn.CheckPaymentsAmount()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCheckPaymentsConfirmations(t *testing.T) {
	mockDb, mock, _ := sqlmock.New()
	db, err := database.FromDialector(postgres.New(postgres.Config{
		Conn:       mockDb,
		DriverName: "postgres",
	}))

	assert.NoError(t, err)
	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"ID", "TransactionID", "Payment"}).AddRow("dataitem", "transaction", "confirmed"))
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "orders" SET "payment"=$1 WHERE id = $2`)).WithArgs("paid", "dataitem").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		arweave := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"block_height":1000,"block_indep_hash":"block_indep_hash","number_of_confirmations":11}`))
				assert.NoError(t, err)
			}))

		defer arweave.Close()

		w, err := wallet.FromPath("../../test/signer.json", arweave.URL)
		assert.NoError(t, err)
		crn, err := New(WithDatabase(db), WithWallet(w))
		assert.NoError(t, err)

		crn.CheckPaymentsConfirmations()
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Not Enough Confirmation", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"ID", "TransactionID", "Payment"}).AddRow("dataitem", "transaction", "confirmed"))
		arweave := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"block_height":1000,"block_indep_hash":"block_indep_hash","number_of_confirmations":5}`))
				assert.NoError(t, err)
			}))

		defer arweave.Close()

		w, err := wallet.FromPath("../../test/signer.json", arweave.URL)
		assert.NoError(t, err)

		crn, err := New(WithDatabase(db), WithWallet(w))
		assert.NoError(t, err)

		crn.CheckPaymentsConfirmations()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
