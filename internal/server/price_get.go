package server

import (
	"errors"
	"io"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PriceGetResponse struct {
	Price   uint64 `json:"price"`
	Address string `json:"address"`
}

func (s *Server) PriceOfUpload(b string) (uint64, error) {
	res, err := http.Get(s.gateway + "/price/" + b)
	if err != nil {
		return 0, err
	}

	r, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	if res.StatusCode >= 400 {
		return 0, errors.New(string(r))
	}

	cost := big.NewInt(0)
	cost.SetString(string(r), 10)

	fee := big.NewInt(1000) // ~0.001
	fee.Quo(cost, fee)

	cost.Add(cost, fee)

	return cost.Uint64(), nil
}

func (s *Server) PriceGet(c *gin.Context) {
	b, valid := c.Params.Get("bytes")
	if !valid {
		c.AbortWithError(http.StatusBadRequest, errors.New("bytes size is required"))
		return
	}

	p, err := s.PriceOfUpload(string(b))
	if err != nil {
		c.JSON(http.StatusInternalServerError, "failed to fetch price")
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, &PriceGetResponse{
		Address: s.wallet.Signer.Address,
		Price:   p,
	})
}
