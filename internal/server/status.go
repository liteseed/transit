package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type StatusResponse struct {
	Name    string `json:"name" format:"string" example:"transit"`
	Version string `json:"version" format:"string" example:"v0.0.1"`
}

// Status godoc
// @Summary      Get status of the server
// @Description  Get the current status of the server
// @Tags         Server
// @Accept       json
// @Produce      json
// @Success      200  {object} StatusResponse
// @Router       / [get]
func (s *Server) Status(c *gin.Context) {
	c.JSON(http.StatusOK, StatusResponse{
		Name:    "Transit",
		Version: s.version,
	})
}
