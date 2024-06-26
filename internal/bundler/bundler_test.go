package bundler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/liteseed/transit/test"
	"github.com/stretchr/testify/assert"
)

func TestDataPost(t *testing.T) {
	d := test.DataItem()

	var expectedRes DataItemPostResponse
	v := fmt.Sprintf(`{"id":"%s","owner":"3XTR7MsJUD9LoaiFRdWswzX1X5BR7AQdl1x2v2zIVck","dataCaches":["localhost"],"deadlineHeight":0,"fastFinalityIndexes":["localhost"],"version":"1"}`, d.ID)
	err := json.Unmarshal([]byte(v), &expectedRes)
	assert.NoError(t, err)

	bun := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		if r.URL.Path == "/tx" && r.Method == "POST" && string(body) == string(d.Raw) {
			w.Write([]byte(v))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer bun.Close()

	t.Run("Success", func(t *testing.T) {

		b := Bundler{
			client: http.DefaultClient,
		}
		res, err := b.DataItemPost(bun.URL[7:], d.Raw)
		assert.NoError(t, err)
		assert.Equal(t, expectedRes, *res)
	})
}

func TestDataPut(t *testing.T) {
	dID := "ak-DBusKYdgM_d6kUqXxIUeAQm_pP1AIE2Bw9jh1a6o"
	txID := "ZyFuVdPioST7Z57LcFrZjCkIdxUNhkt9oItERnwmxuQ"

	var expectedRes DataItemPutResponse
	v := fmt.Sprintf(`{"id":"%s","payment_id":"%s"}`, dID, txID)
	err := json.Unmarshal([]byte(v), &expectedRes)
	assert.NoError(t, err)

	bun := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tx/"+dID+"/"+txID && r.Method == "PUT" {
			w.Write([]byte(v))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer bun.Close()

	t.Run("Success", func(t *testing.T) {

		b := Bundler{
			client: http.DefaultClient,
		}
		res, err := b.DataItemPut(bun.URL[7:], dID, txID)
		assert.NoError(t, err)
		assert.Equal(t, expectedRes, *res)
	})
}
