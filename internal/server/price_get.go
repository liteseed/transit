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
	println("bytes", b)
	res, err := http.Get("http://localhost:8008/price/" + b)
	if err != nil {
		return 0, err
	}

	r, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	cost := big.NewFloat(0)
	cost.SetString(string(r))
	cost.Add(cost, big.NewFloat(1000))

	approx, _ := cost.Uint64()
	return approx, nil
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
