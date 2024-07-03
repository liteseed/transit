package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/liteseed/goar/transaction/data_item"
	"github.com/liteseed/transit/internal/bundler"
	"github.com/liteseed/transit/internal/database/schema"
)

type DataItemPostResponse = bundler.DataItemPostResponse

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

func parseBody(ctx *gin.Context, contentLength int) ([]byte, error) {
	rawData, err := ctx.GetRawData()
	if err != nil {
		return nil, err
	}
	if len(rawData) != contentLength {
		return nil, fmt.Errorf("content-length, body: length mismatch (%d, %d)", contentLength, len(rawData))
	}

	return rawData, nil
}

// Post a data-item to Liteseed godoc
// @Summary      Post a data-item
// @Description  Post your data in a specified ANS-104 data-item format.
// @Tags         Upload
// @Accept       json
// @Produce      json
// @Success      200          {object}  DataItemPostResponse
// @Failure      400,424,500  {object}  HTTPError
// @Router       /tx/ [post]
func (srv *Server) DataItemPost(ctx *gin.Context) {
	header, err := parseHeaders(ctx)
	if err != nil {
		NewError(ctx, http.StatusBadRequest, err)
		return
	}

	contentLength, err := strconv.Atoi(*header.ContentLength)
	if err != nil {
		NewError(ctx, http.StatusBadRequest, err)
		return
	}
	rawData, err := parseBody(ctx, contentLength)
	if err != nil {
		NewError(ctx, http.StatusBadRequest, err)
		return
	}

	dataItem, err := data_item.Decode(rawData)
	if err != nil {
		log.Println(err)
		NewError(ctx, http.StatusBadRequest, errors.New("failed to decode data item"))
		return
	}

	err = dataItem.Verify()
	if err != nil {
		log.Println(err)
		NewError(ctx, http.StatusBadRequest, errors.New("failed to verify data item"))
		return
	}

	staker, err := srv.contract.Initiate(dataItem.ID, contentLength)
	if err != nil {
		NewError(ctx, http.StatusFailedDependency, err)
		return
	}
	res, err := srv.bundler.DataItemPost(staker.URL, dataItem.Raw)
	if err != nil {
		NewError(ctx, http.StatusFailedDependency, err)
		return
	}

	o := &schema.Order{
		Id:      dataItem.ID,
		Address: staker.ID,
		URL:     staker.URL,
		Payment: schema.Unpaid,
		Status:  schema.Created,
		Size:    len(dataItem.Raw),
	}

	err = srv.database.CreateOrder(o)
	if err != nil {
		NewError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, res)
}
