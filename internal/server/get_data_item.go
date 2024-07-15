package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (srv *Server) GetDataItem(ctx *gin.Context) {
	id := ctx.Param("id")
	o, err := srv.database.GetOrder(id)
	if err != nil {
		NewError(ctx, http.StatusNotFound, err)
		return
	}

	res, err := srv.bundler.DataItemGet(o.URL, id)
	if err != nil {
		NewError(ctx, http.StatusFailedDependency, err)
		return
	}

	ctx.JSON(http.StatusOK, res)
}
