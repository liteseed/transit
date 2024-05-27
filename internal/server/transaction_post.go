package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"

	"github.com/everFinance/goar/utils"
	"github.com/gin-gonic/gin"
	"github.com/liteseed/transit/internal/database/schema"
)

type DataItemPostRequestHeader struct {
	ContentType   *string `header:"content-type" binding:"required"`
	ContentLength *string `header:"content-length" binding:"required"`
	TransactionID *string `header:"transaction-id" binding:"required"`
}

type DataItemPostResponse struct {
	ID                  string   `json:"id"`
	Owner               string   `json:"owner"`
	DataCaches          []string `json:"dataCaches"`
	DeadlineHeight      uint     `json:"deadlineHeight"`
	FastFinalityIndexes []string `json:"fastFinalityIndexes"`
	Price               uint64   `json:"price"`
	Version             string   `json:"version"`
}

func postData(url string, b []byte) (*DataItemPostResponse, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
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
		return nil, fmt.Errorf("content-type: unsupported")
	}
	if *header.TransactionID != "" {
		return nil, fmt.Errorf("payment transaction-id is required")
	}
	if *header.ContentLength != "" {
		return nil, fmt.Errorf("content-length is required")
	}
	return header, nil
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

func (s *Server) checkPrice(transactionID string, contentLength string) (uint64, error) {
	price, err := s.PriceOfUpload(contentLength)
	if err != nil {
		return 0, err
	}

	tx, err := s.wallet.Client.GetTransactionByID(transactionID)
	if err != nil {
		return 0, err
	}
	t := big.NewInt(0)
	t.SetString(string(tx.Quantity), 10)
	payment := t.Uint64()

	if payment < price {
		return 0, errors.New("not enough ar to upload, contact support: hello@liteseed.xyz")
	}
	return payment, nil
}

// POST /tx
func (s *Server) DataItemPost(context *gin.Context) {
	header, err := parseHeaders(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Error(err)
		return
	}

	transactionID := *header.TransactionID
	contentLength, err := strconv.Atoi(*header.ContentLength)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Error(err)
		return
	}

	payment, err := s.checkPrice(transactionID, *header.ContentLength)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "price verification failed"})
		context.Error(err)
		return
	}

	rawData, err := parseBody(context, contentLength)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Error(err)
		return
	}

	dataItem, err := utils.DecodeBundleItem(rawData)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "failed to decode data item"})
		context.Error(err)
		return
	}

	err = utils.VerifyBundleItem(*dataItem)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "failed to verify data item"})
		context.Error(err)
		return
	}

	staker, err := s.contract.Initiate(dataItem.Id, contentLength)
	if err != nil {
		context.JSON(http.StatusFailedDependency, gin.H{"error": "failed to find staker"})
		context.Error(err)
		return
	}
	res, err := postData(staker.URL, rawData)

	if err != nil {
		context.JSON(http.StatusFailedDependency, gin.H{"error": "failed to post to assigned staker"})
		context.Error(err)
		return
	}

	o := &schema.Order{
		ID:     dataItem.Id,
		Status: schema.Queued,
		Price:  payment,
		Bundler: schema.Bundler{
			Address: staker.ID,
			URL:     staker.URL,
		},
	}

	err = s.database.CreateOrder(o)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "failed to create order"})
		context.Error(err)
		return
	}

	context.JSON(http.StatusCreated, res)
}
