package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (srv *Server) GetDataItem(ctx *gin.Context) {
	id := ctx.Param("id")
	o, err := srv.database.GetOrder(id)
	if err != nil {
		NewError(ctx, http.StatusNotFound, fmt.Errorf("not found %s", id))
		return
	}

	res, err := srv.bundler.DataItemGet(o.URL, id)
	if err != nil {
		NewError(ctx, http.StatusFailedDependency, err)
		return
	}

	ctx.Data(http.StatusOK, "application/octet-stream", res)
}
