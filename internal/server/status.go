package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type StatusResponse struct {
	Name    string `json:"name" format:"string" example:"transit"`
	Version string `json:"version" format:"string" example:"v0.0.1"`
}

func (s *Server) Status(c *gin.Context) {
	c.JSON(http.StatusOK, StatusResponse{
		Name:    "Transit",
		Version: s.version,
	})
}
