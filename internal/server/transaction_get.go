package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Get /tx
func (s *Server) DataItemGet(context *gin.Context) {
	id := context.Param("id")
	println(id)

	raw, err := s.store.Get(id)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "transaction id does not exist"})
		context.Error(err)
		return
	}

	context.JSON(
		http.StatusOK,
		raw,
	)
}
