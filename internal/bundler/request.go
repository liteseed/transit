package bundler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/liteseed/transit/internal/utils"
)

func (b *Bundler) get(url string) ([]byte, error) {
	u, err := utils.ParseUrl(url)
	if err != nil {
		return nil, err
	}

	res, err := b.client.Get(u)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("%d: %s", res.StatusCode, string(body))
	}
	return body, nil
}

func (b *Bundler) post(url string, contentType string, payload []byte) ([]byte, error) {
	u, err := utils.ParseUrl(url)
	if err != nil {
		return nil, err
	}

	res, err := b.client.Post(u, contentType, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("%d: %s", res.StatusCode, string(body))
	}
	return body, nil
}

func (b *Bundler) put(url string, contentType string, payload []byte) ([]byte, error) {
	u, err := utils.ParseUrl(url)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", u, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", contentType)

	res, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("%d: %s", res.StatusCode, string(body))
	}
	return body, nil
}
