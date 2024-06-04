package server

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liteseed/transit/internal/database/schema"
)

// Get /tx/:id
func (s *Server) DataItemStatusGet(context *gin.Context) {
	id := context.Param("id")

	b, err := s.database.GetOrder(&schema.Order{ID: id})
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "data id does not exist"})
		return
	}

	resp, err := http.Get("http://" + b.URL + "/tx/" + id + "/status")
	if err != nil {
		context.JSON(http.StatusFailedDependency, gin.H{"error": "something went wrong"})
		return
	}

	var raw []byte
	raw, err = io.ReadAll(resp.Body)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "unable to read data"})
		return
	}

	context.JSON(http.StatusOK, string(raw))
}
