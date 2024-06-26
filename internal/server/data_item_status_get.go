package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Get the status of the posted data-item godoc
// @Summary      Get the status of a data-item
// @Description  Get the current status of a posted data-item.
// @Description  Response "created", "queued", "sent", "confirmed", "failed", "invalid"
// @Tags         Upload
// @Accept       json
// @Produce      json
// @Param        id            path          string    true  "ID of the data-item"
// @Success      200           {string}      status
// @Failure      404,424,500   {object}      HTTPError
// @Router       /tx/{id}/status [get]
func (srv *Server) DataItemStatusGet(ctx *gin.Context) {
	id := ctx.Param("id")

	o, err := srv.database.GetOrder(id)
	if err != nil {
		NewError(ctx, http.StatusNotFound, err)
		return
	}

	res, err := srv.bundler.DataItemStatusGet(o.URL, id)
	if err != nil {
		NewError(ctx, http.StatusFailedDependency, err)
		return
	}

	ctx.JSON(http.StatusOK, string(res))
}
