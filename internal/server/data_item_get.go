package server

import (
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liteseed/transit/internal/database/schema"
)

// Get /tx
func (s *Server) DataItemGet(context *gin.Context) {
	id := context.Param("id")

	b, err := s.database.GetOrder(&schema.Order{ID: id})
	if err != nil {
		log.Println(err)
		context.JSON(http.StatusNotFound, gin.H{"error": "data-item does not exist"})
		return
	}

	resp, err := http.Get("http://" + b.URL + "/tx/" + id)
	if err != nil {
		context.Status(http.StatusFailedDependency)
		return
	}
	var raw []byte
	raw, err = io.ReadAll(resp.Body)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "unable to read data"})
		return
	}

	context.Data(
		http.StatusOK,
		"application/octet-stream",
		raw,
	)
}
