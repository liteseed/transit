package utils

import (
	"math/big"
	"net/url"

	"github.com/gin-gonic/gin"
)

func CalculatePriceWithFee(p string) string {
	cost := big.NewInt(0)
	cost.SetString(p, 10)

	fee := big.NewInt(1000) // ~0.001
	fee.Quo(cost, fee)

	cost.Add(cost, fee) // cost = cost * fee
	return cost.String()
}

func ParseUrl(u string) (string, error) {
	if gin.Mode() == gin.DebugMode {
		return "https://" + u, nil
	}
	p, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	p.Scheme = "https"

	return p.String(), nil
}
