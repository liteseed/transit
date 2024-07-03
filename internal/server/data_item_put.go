package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/liteseed/transit/internal/database/schema"
)

type DataItemPutResponse struct {
	ID        string `json:"id"`
	PaymentID string `json:"paymentId"`
}

// Update payment id to data-item godoc
// @Summary      Send a payment id for a data-item
// @Description  Once a payment is made send a transaction id for a data-item
// @Tags         Payment
// @Accept       json
// @Produce      json
// @Param        id               path      string              true  "data-item id"
// @Param        paymentId        path      string              true  "payment id"
// @Success      200              {object}  DataItemPutResponse
// @Failure      400,404          {object}  HTTPError
// @Router       /tx/{id}/{payment_id} [put]
func (srv *Server) DataItemPut(ctx *gin.Context) {
	dataItemID := ctx.Param("id")
	paymentID := ctx.Param("payment_id")
	err := srv.database.UpdateOrder(dataItemID, &schema.Order{TransactionId: paymentID, Status: schema.Queued})
	if err != nil {
		NewError(ctx, http.StatusNotFound, err)
		return
	}
	ctx.JSON(http.StatusAccepted, DataItemPutResponse{ID: dataItemID, PaymentID: paymentID})
}
