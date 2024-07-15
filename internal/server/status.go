package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (srv *Server) Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"name": "Transit", "version": srv.version})
}
