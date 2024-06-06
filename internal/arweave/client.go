package client

import (
	"errors"
	"io"
	"math/big"
	"net/http"
)

type Arweave struct {
	gateway string
	client  *http.Client
}

func (arweave *Arweave) Price(b string, target string) (string, error) {
	res, err := arweave.client.Get(arweave.gateway + "/price/" + b + "/" + target)
	if err != nil {
		return "", err
	}

	r, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode >= 400 {
		return "", errors.New(string(r))
	}

	cost := big.NewInt(0)
	cost.SetString(string(r), 10)

	fee := big.NewInt(1000) // ~0.001
	fee.Quo(cost, fee)

	cost.Add(cost, fee)

	return cost.String(), nil
}

func (arweave *Arweave) SendWinston(quantity string, target string) {

}
