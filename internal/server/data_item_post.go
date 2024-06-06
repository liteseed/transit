package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/everFinance/goar/utils"
	"github.com/gin-gonic/gin"
	"github.com/liteseed/transit/internal/database/schema"
)

type DataItemPostRequestHeader struct {
	ContentType   *string `header:"content-type" binding:"required"`
	ContentLength *string `header:"content-length" binding:"required"`
}

type DataItemPostResponse struct {
	ID                  string   `json:"id"`
	Owner               string   `json:"owner"`
	DataCaches          []string `json:"dataCaches"`
	DeadlineHeight      uint     `json:"deadlineHeight"`
	FastFinalityIndexes []string `json:"fastFinalityIndexes"`
	Version             string   `json:"version"`
}

func postData(u string, b []byte) (*DataItemPostResponse, error) {
	req, err := http.NewRequest("POST", "http://"+u+"/tx", bytes.NewBuffer(b))
	req.Header.Set("content-type", "application/octet-stream")
	req.Header.Set("content-length", fmt.Sprint(len(b)))

	if err != nil {
		return nil, err
	}

	c := http.DefaultClient
	resp, err := c.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		return nil, errors.New("failed to post data to bundler")
	}
	var res DataItemPostResponse
	r, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(r, &res); err != nil {
		return nil, err
	}
	return &res, nil
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
func (s *Server) DataItemPost(context *gin.Context) {
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

	dataItem, err := utils.DecodeBundleItem(rawData)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "failed to decode data item"})
		return
	}

	err = utils.VerifyBundleItem(*dataItem)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "failed to verify data item"})
		return
	}

	staker, err := s.contract.Initiate(dataItem.Id, contentLength)
	if err != nil {
		context.JSON(http.StatusFailedDependency, gin.H{"error": "failed to post to bundler"})
		return
	}

	res, err := postData(staker.URL, rawData)
	if err != nil {
		context.JSON(http.StatusFailedDependency, gin.H{"error": "failed to post to bundler"})
		return
	}

	o := &schema.Order{
		ID:      dataItem.Id,
		Address: staker.ID,
		URL:     staker.URL,
		Payment: schema.Unpaid,
		Status:  schema.Created,
		Size:    uint(len(dataItem.ItemBinary)),
	}

	err = s.database.CreateOrder(o)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "failed to create order"})
		return
	}

	context.JSON(http.StatusCreated, res)
}
