package server

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/liteseed/goar/tag"
	"github.com/liteseed/goar/transaction/data_item"
	"github.com/liteseed/transit/internal/database/schema"
)

// DataPost
//
// Post unsigned data to Liteseed godoc
// @Summary      Post data
// @Description  Post data or file to Liteseed.
// @Tags         Upload
// @Accept       json
// @Produce      json
// @Success      200          {object}  PostResponse
// @Failure      400,424,500  {object}  HTTPError
// @Router       /tx/ [post]

func (srv *Server) DataPost(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		NewError(ctx, http.StatusBadRequest, err)
		return
	}

	tags := []tag.Tag{}

	tagsString := ctx.PostFormArray("tags[]")
	for i, tagValue := range tagsString {
		tags = append(tags, tag.Tag{Name: strconv.Itoa(i), Value: tagValue})
	}
	mf, err := file.Open()
	if err != nil {
		NewError(ctx, http.StatusBadRequest, err)
		return
	}
	defer mf.Close()

	raw, err := io.ReadAll(mf)
	if err != nil {
		NewError(ctx, http.StatusBadRequest, err)
		return
	}

	log.Println(len(raw))
	d := data_item.New(raw, "", "", &tags)

	err = d.Sign(srv.wallet.Signer)
	if err != nil {
		log.Println(err)
		NewError(ctx, http.StatusInternalServerError, errors.New("failed to sign data item"))
		return
	}

	staker, err := srv.contract.Initiate(d.ID, len(d.Raw))
	if err != nil {
		log.Println(err)
		NewError(ctx, http.StatusFailedDependency, errors.New("failed to initiate upload"))
		return
	}
	res, err := srv.bundler.DataItemPost(staker.URL, d.Raw)
	if err != nil {
		log.Println(err)
		NewError(ctx, http.StatusFailedDependency, errors.New("failed to send to bundler"))
		return
	}

	o := &schema.Order{
		Id:      d.ID,
		Address: staker.ID,
		URL:     staker.URL,
		Payment: schema.Unpaid,
		Status:  schema.Created,
		Size:    len(d.Raw),
	}

	err = srv.database.CreateOrder(o)
	if err != nil {
		NewError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, res)
}
