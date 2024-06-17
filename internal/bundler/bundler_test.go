package bundler

import (
	"net/http"
	"testing"

	"github.com/liteseed/goar/signer"
	"github.com/liteseed/goar/transaction/data_item"
	"github.com/stretchr/testify/assert"
)

func TestDataPost(t *testing.T) {
	jwk, err := signer.New()
	assert.NoError(t, err)

	s, err := signer.FromJWK(jwk)
	assert.NoError(t, err)

	dataItem := data_item.New([]byte{1, 2, 3}, "", "", nil)
	err = dataItem.Sign(s)
	assert.NoError(t, err)
	t.Run("Success", func(t *testing.T) {

		b := Bundler{
			client: http.DefaultClient,
		}
		res, err := b.DataItemPost("localhost:8080", dataItem.Raw)
		assert.NoError(t, err)
		t.Log(res)
	})
}
