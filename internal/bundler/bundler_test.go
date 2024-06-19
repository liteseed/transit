package bundler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/liteseed/goar/signer"
	"github.com/liteseed/goar/transaction/data_item"
	"github.com/stretchr/testify/assert"
)

func TestDataPost(t *testing.T) {
	s, err := signer.FromPath("../../test/signer.json")
	assert.NoError(t, err)

	dataItem := data_item.New([]byte{1, 2, 3}, "", "", nil)
	err = dataItem.Sign(s)
	assert.NoError(t, err)

	var expectedRes DataItemPostResponse
	v := fmt.Sprintf(`{"id":"%s","owner":"3XTR7MsJUD9LoaiFRdWswzX1X5BR7AQdl1x2v2zIVck","dataCaches":["localhost"],"deadlineHeight":0,"fastFinalityIndexes":["localhost"],"version":"1"}`, dataItem.ID)
	err = json.Unmarshal([]byte(v), &expectedRes)

	bun := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(v))
		assert.NoError(t, err)
	}))
	defer bun.Close()

	t.Run("Success", func(t *testing.T) {

		b := Bundler{
			client: http.DefaultClient,
		}
		res, err := b.DataItemPost(bun.URL[7:], dataItem.Raw)
		assert.NoError(t, err)
		assert.Equal(t, expectedRes, *res)
	})
}
