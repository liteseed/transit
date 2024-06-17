package bundler

import (
	"encoding/json"
	"net/http"
)

type Bundler struct {
	client *http.Client
}

func New() *Bundler {
	return &Bundler{
		client: http.DefaultClient,
	}
}

func (b *Bundler) DataItemGet(url string, id string) ([]byte, error) {
	data, err := b.get(url + "/" + "tx" + "/" + id)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type DataItemPostResponse struct {
	ID                  string   `json:"id"`
	Owner               string   `json:"owner"`
	DataCaches          []string `json:"dataCaches"`
	DeadlineHeight      uint     `json:"deadlineHeight"`
	FastFinalityIndexes []string `json:"fastFinalityIndexes"`
	Version             string   `json:"version"`
}

func (b *Bundler) DataItemPost(url string, data []byte) (*DataItemPostResponse, error) {
	data, err := b.post(url+"/"+"tx", data)
	if err != nil {
		return nil, err
	}
	var res DataItemPostResponse
	if err = json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (b *Bundler) DataItemPut(url string, id string, paymentID string) ([]byte, error) {
	data, err := b.put(url+"/", "", nil)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (b *Bundler) DataItemStatusGet(url string) ([]byte, error) {
	data, err := b.get(url)
	if err != nil {
		return nil, err
	}
	return data, nil
}
