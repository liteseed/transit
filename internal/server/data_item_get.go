package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Get a complete data-item godoc
// @Summary      Get a posted data-item
// @Description  get all the fields of a posted data-item
// @Tags         Upload
// @Accept       json
// @Produce      json
// @Param        id           path      string  true  "ID of the data-item"
// @Success      200          {string}  dataitem
// @Failure      404,424,500  {object}  HTTPError
// @Router       /tx/{id} [get]
func (srv *Server) DataItemGet(ctx *gin.Context) {
	id := ctx.Param("id")
	o, err := srv.database.GetOrder(id)
	if err != nil {
		NewError(ctx, 404, err)
		return
	}

	raw, err := srv.bundler.DataItemGet(o.URL, o.ID)
	if err != nil {
		NewError(ctx, 424, err)
		return
	}

	ctx.Data(http.StatusOK, "application/octet-stream", raw)
}
