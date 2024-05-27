package server

import (
	"errors"
	"io"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) PriceGet(c *gin.Context) {
	b, valid := c.Params.Get("bytes")
	if !valid {
		c.AbortWithError(http.StatusBadRequest, errors.New("bytes size is required"))
		return
	}

	res, err := http.Get("https://arweave.net/price/" + b)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	r, err := io.ReadAll(res.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	cost := big.NewFloat(0)
	cost.SetString(string(r))

	gasApprox := big.NewFloat(0).Mul(cost, big.NewFloat(0.01))
	cost.Add(cost, gasApprox)

	approx, _ := cost.Uint64()

	c.JSON(http.StatusOK, approx)
}
