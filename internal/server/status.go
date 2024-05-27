package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET Status reports if the server is operational.
//
// GET /status
func (s *Server) Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Name":    "Transit",
		"Version": "v0.1.0",
	})
}
