package server

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liteseed/goar/crypto"
	"github.com/liteseed/goar/transaction/data_item"
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

	o, err := srv.database.GetOrder(id)
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
		NewError(ctx, http.StatusInternalServerError, err)
		return
	}
	log.Println(d.ID)

	b, err := crypto.Base64Decode(d.Data)
	if err != nil {
		NewError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Data(http.StatusOK, "application/octet-stream", b)
}
