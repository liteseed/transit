package server

import (
	"errors"
	"fmt"
	"log"
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
		return nil, errors.New("required header(s) - content-type, content-length")
	}
	if *header.ContentType != CONTENT_TYPE_OCTET_STREAM {
		return nil, errors.New("required header(s) - content-type: application/octet-stream")
	}
	if *header.ContentLength == "0" {
		return nil, errors.New("required header(s) - content-length")
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

func (srv *Server) DataItemPost(ctx *gin.Context) {
	header, err := parseHeaders(ctx)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contentLength, err := strconv.Atoi(*header.ContentLength)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rawData, err := parseBody(ctx, contentLength)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dataItem, err := data_item.Decode(rawData)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to decode data item"})
		return
	}

	err = data_item.Verify(dataItem)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to verify data item"})
		return
	}

	staker, err := srv.contract.Initiate(dataItem.ID, contentLength)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusFailedDependency, gin.H{"error": "failed to initiate"})
		return
	}
	res, err := srv.bundler.DataItemPost(staker.URL, dataItem.Raw)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusFailedDependency, gin.H{"error": "failed to post to bundler"})
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
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to create order"})
		return
	}

	ctx.JSON(http.StatusCreated, res)
}
