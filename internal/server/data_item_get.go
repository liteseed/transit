package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liteseed/transit/internal/database/schema"
)

// Get /tx
func (srv *Server) DataItemGet(context *gin.Context) {
	id := context.Param("id")

	o, err := srv.database.GetOrder(&schema.Order{ID: id})
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "data-item does not exist"})
		return
	}

	raw, err := srv.bundler.DataItemGet(o.Address, o.ID)
	if err != nil {
		context.JSON(http.StatusFailedDependency, gin.H{"error": err})
		return
	}

	context.Data(
		http.StatusOK,
		"application/octet-stream",
		raw,
	)
}
