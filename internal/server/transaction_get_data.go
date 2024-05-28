package server

import (
	"encoding/base64"
	"net/http"

	"github.com/everFinance/goar/utils"
	"github.com/gin-gonic/gin"
)

// Get /tx
func (s *Server) DataItemDataGet(context *gin.Context) {
	id := context.Param("id")
	println(id)

	raw, err := s.store.Get(id)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "transaction id does not exist"})
		return
	}

	dataItem, err := utils.DecodeBundleItem(raw)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "transaction id does not exist"})
		return
	}

	data, err := base64.RawURLEncoding.DecodeString(dataItem.Data)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "unable to decode"})
		return
	}

	context.Data(
		http.StatusOK,
		"application/octet-stream",
		data,
	)
}
