package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/liteseed/transit/internal/utils"
)

type PriceGetResponse struct {
	Price   string `json:"price"`
	Address string `json:"address"`
}

func (srv *Server) PriceGet(c *gin.Context) {
	b, valid := c.Params.Get("bytes")
	if !valid {
		c.JSON(http.StatusBadRequest, errors.New("bytes size is required"))
		return
	}

	size, err := strconv.Atoi(b)
	if err != nil {
		c.JSON(http.StatusBadRequest, "size should be between 0 and 2^32-1")
		return
	}

	p, err := srv.wallet.Client.GetTransactionPrice(size, srv.wallet.Signer.Address)
	if err != nil {
		c.JSON(http.StatusFailedDependency, "failed to fetch price")
		return
	}

	c.JSON(http.StatusOK, &PriceGetResponse{Address: srv.wallet.Signer.Address, Price: utils.CalculatePriceWithFee(p)})
}
