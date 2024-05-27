package server

import (
	"errors"
	"io"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) PriceOfUpload(b string) (uint64, error) {

	res, err := http.Get("https://arweave.net/price/" + b)
	if err != nil {
		return 0, err
	}

	r, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	cost := big.NewFloat(0)
	cost.SetString(string(r))

	gasApprox := big.NewFloat(0).Mul(cost, big.NewFloat(0.01))
	cost.Add(cost, gasApprox)

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
	c.JSON(http.StatusOK, p)
}
