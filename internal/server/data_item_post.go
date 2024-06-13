package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/liteseed/goar/transaction/data_item"
	"github.com/liteseed/transit/internal/database/schema"
)

type DataItemPostRequestHeader struct {
	ContentType   *string `header:"content-type" binding:"required"`
	ContentLength *string `header:"content-length" binding:"required"`
}

func parseHeaders(context *gin.Context) (*DataItemPostRequestHeader, error) {
	header := &DataItemPostRequestHeader{}
	if err := context.ShouldBindHeader(header); err != nil {
		return nil, err
	}
	if *header.ContentType != CONTENT_TYPE_OCTET_STREAM {
		return nil, fmt.Errorf("required - content-type: application/octet-stream")
	}
	if *header.ContentLength == "" {
		return nil, fmt.Errorf("required - content-length")
	}
	return header, nil
}

type Transaction struct {
	ID       string `json:"id"`
	Owner    string `json:"owner"`
	Quantity string `json:"quantity"`
}

func parseBody(context *gin.Context, contentLength int) ([]byte, error) {
	rawData, err := context.GetRawData()
	if err != nil {
		return nil, err
	}
	if len(rawData) != contentLength {
		return nil, fmt.Errorf("content-length, body: length mismatch (%d, %d)", contentLength, len(rawData))
	}

	return rawData, nil
}

// POST /tx
// 1. Parse Headers - content-length, content-type, x-transaction-id
// 2.
func (srv *Server) DataItemPost(context *gin.Context) {
	header, err := parseHeaders(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contentLength, err := strconv.Atoi(*header.ContentLength)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rawData, err := parseBody(context, contentLength)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dataItem, err := data_item.Decode(rawData)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "failed to decode data item"})
		return
	}

	err = data_item.Verify(dataItem)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "failed to verify data item"})
		return
	}

	staker, err := srv.contract.Initiate(dataItem.ID, contentLength)
	if err != nil {
		context.JSON(http.StatusFailedDependency, gin.H{"error": "failed to post to bundler"})
		return
	}

	res, err := srv.bundler.DataItemPost(staker.URL, dataItem.Raw)
	if err != nil {
		context.JSON(http.StatusFailedDependency, gin.H{"error": "failed to post to bundler"})
		return
	}

	o := &schema.Order{
		ID:      dataItem.ID,
		Address: staker.ID,
		URL:     staker.URL,
		Payment: schema.Unpaid,
		Status:  schema.Created,
		Size:    len(dataItem.Raw),
	}

	err = srv.database.CreateOrder(o)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "failed to create order"})
		return
	}

	context.JSON(http.StatusCreated, res)
}
