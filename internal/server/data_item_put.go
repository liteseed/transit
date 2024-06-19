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

// PUT /tx/:id/:payment_id
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
