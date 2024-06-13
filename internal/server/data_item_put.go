package server

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/liteseed/transit/internal/database/schema"
)

type DataItemPutResponse struct {
	ID             string `json:"id"`
	DeadlineHeight uint   `json:"deadlineHeight"`
}

// PUT /tx/:id/:transaction_id
func (srv *Server) DataItemPut(context *gin.Context) {
	dataItemID := context.Param("id")
	transactionID := context.Param("transaction_id")
	err := srv.database.UpdateOrder(&schema.Order{ID: dataItemID, TransactionID: transactionID, Status: schema.Queued})
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}
	info, err := srv.wallet.Client.GetNetworkInfo()
	if err != nil {
		context.Status(http.StatusFailedDependency)
		log.Println(err)
		return
	}

	deadline := uint(info.Height) + 220

	context.JSON(http.StatusAccepted, DataItemPutResponse{
		ID:             dataItemID,
		DeadlineHeight: deadline,
	})
}
