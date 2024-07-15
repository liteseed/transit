package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/liteseed/transit/internal/utils"
)

type PriceGetResponse struct {
	Price   string `json:"price" example:"1000000000000" format:"string"`
	Address string `json:"address" example:"Cbj95zDZBBhmyht6iFlEf7xmSCSVZGw436V6HWmm9Ek" format:"string"`
}

// PriceGet
//
// Price godoc
// @Summary      Get price of upload
// @Description  Get the current price of data upload using the Liteseed Network.
// @Description  It returns the price of upload in wei and the address to pay.
// @Tags         Payment
// @Accept       json
// @Produce      json
// @Param        bytes             path      int  true  "Size of Data" minimum(1) maximum(2147483647)
// @Success      200               {object}  PriceGetResponse
// @Failure      400,424,500       {object}  HTTPError
// @Router       /price/{bytes} [get]
func (srv *Server) PriceGet(ctx *gin.Context) {
	b, valid := ctx.Params.Get("bytes")
	if !valid {
		NewError(ctx, http.StatusBadRequest, errors.New("byte size is invalid"))
		return
	}

	size, err := strconv.Atoi(b)
	if err != nil || size <= 0 {
		NewError(ctx, http.StatusBadRequest, errors.New("byte size should be between 1 and 2^32-1"))
		return
	}

	p, err := srv.wallet.Client.GetTransactionPrice(size, srv.wallet.Signer.Address)
	if err != nil {
		NewError(ctx, http.StatusFailedDependency, errors.New("failed to fetch price"))
		return
	}

	ctx.JSON(http.StatusOK, &PriceGetResponse{Address: srv.wallet.Signer.Address, Price: utils.CalculatePriceWithFee(p)})
}
