package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liteseed/goar/transaction/data_item"
	"github.com/liteseed/transit/internal/database/schema"
)

// Get Data-Item Data godoc
// @Summary      Get data of a data-item
// @Description  get only the data of a posted data-item
// @Tags         Upload
// @Accept       json
// @Produce      octet-stream
// @Param        id   path    string    true      "ID of the Data-Item"
// @Success      200  				{bytes}   data
// @Failure      404,424,500  {object}  HTTPError
// @Router       /tx/{id}/data [get]
func (srv *Server) DataItemDataGet(ctx *gin.Context) {
	id := ctx.Param("id")

	o, err := srv.database.GetOrder(&schema.Order{ID: id})
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "data id does not exist"})
		return
	}

	res, err := srv.bundler.DataItemGet(o.URL, id)
	if err != nil {
		ctx.JSON(http.StatusFailedDependency, gin.H{"error": err})
		return
	}

	d, err := data_item.Decode(res)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, d.Data)
}
