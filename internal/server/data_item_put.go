package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/liteseed/transit/internal/database/schema"
)

type DataItemPutResponse struct {
	ID        string `json:"id"`
	PaymentID string `json:"payment_id"`
}

// Update payment id to data-item godoc
// @Summary      Send a payment id for a data-item
// @Description  Once a payment is made send a transaction id for a data-item
// @Tags         Upload
// @Accept       json
// @Produce      json
// @Param        id               path      string  true  "Data-Item id"
// @Param        payment_id       path      string  true  "Payment id"
// @Success      200              {object}  DataItemPutResponse
// @Failure      400,404,424,500  {object}  HTTPError
// @Router       /tx/{id}/{payment_id} [put]
func (srv *Server) DataItemPut(context *gin.Context) {
	dataItemID := context.Param("id")
	paymentID := context.Param("payment_id")
	err := srv.database.UpdateOrder(dataItemID, &schema.Order{TransactionID: paymentID, Status: schema.Queued})
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}
	context.JSON(http.StatusAccepted, DataItemPutResponse{ID: dataItemID, PaymentID: paymentID})
}
