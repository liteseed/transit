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

type PostResponse = bundler.DataItemPostResponse

type DataItemPostRequestHeader struct {
	ContentType   *string `header:"content-type" binding:"required"`
	ContentLength *string `header:"content-length" binding:"required"`
}

func dataItemPostRequestHeader(ctx *gin.Context) (*DataItemPostRequestHeader, error) {
	header := &DataItemPostRequestHeader{}
	if err := ctx.ShouldBindHeader(header); err != nil {
		return nil, errors.New("required header(s) - content-type, content-length")
	}
	if *header.ContentType != ContentTypeOctetStream {
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

func dataItemPostRequestBody(ctx *gin.Context, contentLength int) ([]byte, error) {
	rawData, err := ctx.GetRawData()
	if err != nil {
		return nil, err
	}
	if len(rawData) != contentLength {
		return nil, fmt.Errorf("content-length, body: length mismatch (%d, %d)", contentLength, len(rawData))
	}

	return rawData, nil
}

// DataItemPost
//
// Post a data-item to Liteseed godoc
// @Summary      Post a data-item
// @Description  Post your data in a specified ANS-104 data-item format.
// @Tags         Upload
// @Accept       json
// @Produce      json
// @Success      200          {object}  PostResponse
// @Failure      400,424,500  {object}  HTTPError
// @Router       /tx [post]
func (srv *Server) DataItemPost(ctx *gin.Context) {
	headers, err := dataItemPostRequestHeader(ctx)
	if err != nil {
		NewError(ctx, http.StatusBadRequest, err)
		return
	}

	contentLength, err := strconv.Atoi(*headers.ContentLength)
	if err != nil {
		NewError(ctx, http.StatusBadRequest, err)
		return
	}
	rawData, err := dataItemPostRequestBody(ctx, contentLength)
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
