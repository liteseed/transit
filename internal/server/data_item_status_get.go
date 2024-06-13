package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liteseed/transit/internal/database/schema"
)

// Get /tx/:id
func (srv *Server) DataItemStatusGet(context *gin.Context) {
	id := context.Param("id")

	o, err := srv.database.GetOrder(&schema.Order{ID: id})
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "data id does not exist"})
		return
	}

	res, err := srv.bundler.DataItemStatusGet(o.Address)
	if err != nil {
		context.JSON(http.StatusFailedDependency, gin.H{"error": err})
		return
	}

	context.JSON(http.StatusOK, string(res))
}
