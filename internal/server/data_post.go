package server

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/liteseed/goar/transaction/data_item"
	"github.com/liteseed/transit/internal/database/schema"
)



type DataPostRequestHeader struct {
	ContentType   *string `header:"content-type" binding:"required"`
	ContentLength *string `header:"content-length" binding:"required"`
}

func parse(ctx *gin.Context) (*DataPostRequestHeader, error) {
	header := &DataPostRequestHeader{}
	if err := ctx.ShouldBindHeader(header); err != nil {
		return nil, errors.New("required header(s) - content-type, content-length")
	}
	if *header.ContentType != "" {
		return nil, errors.New("required header(s) - content-type")
	}
	if *header.ContentLength == "0" {
		return nil, errors.New("required header(s) - content-length")
	}
	return header, nil
}

// Post a data to Liteseed godoc
// @Summary      Post data
// @Description  Post your data using liteseed
// @Tags         Upload
// @Accept       json
// @Produce      json
// @Success      200          {object}  DataItemPostResponse
// @Failure      400,424,500  {object}  HTTPError
// @Router       /tx/ [post]
func (srv *Server) DataPost(ctx *gin.Context) {
	header, err := parse(ctx)
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
	dataItem := data_item.New(rawData, "", "", nil)

	_, err = srv.wallet.SignDataItem(dataItem)
	if err != nil {
		log.Println(err)
		NewError(ctx, http.StatusInternalServerError, errors.New("failed to sign data item"))
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
