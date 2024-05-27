package server

import (
	"log"
	"net/http"

	"github.com/everFinance/goar/types"
	"github.com/everFinance/goar/utils"
	"github.com/gin-gonic/gin"
)

type PostDataResponse struct {
	Id string `json:"id"`
}

// POST /data
func (s *Server) DataPost(c *gin.Context) {
	header, err := verifyHeaders(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	rawData, err := decodeBody(c, *header.ContentLength)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	tags := []types.Tag{{Name: "", Value: ""}}
	item, err := utils.NewBundleItem("", 1, "", "", rawData, tags)
	if err != nil {
		log.Println("goar: failed to create data-item", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, PostDataResponse{Id: item.Id})
}
