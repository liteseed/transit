package server

import (
	"net/http"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/liteseed/goar/crypto"
	"github.com/liteseed/goar/transaction/data_item"
)

// Get Data-Item Data godoc
// @Summary      Get data of a data-item
// @Description  Get only the data of a posted data-item. It tries to automatically detect mime-type.
// @Description  You can specify the response mime-type by either sending an mime-type query parameter or an accept header in the request.
// @Description  Supported mime-type are listed here - https://github.com/gabriel-vasile/mimetype/blob/master/supported_mimes.md.
// @Description  If all else fails defaults to `application/octet-stream`
// @Tags         Fetch
// @Accept       json
// @Param        id           path      string    true      "ID of the data-Item"
// @Param        mime-type    query     string    false     "Mime type of the response"
// @Success      200  				{bytes}   data
// @Failure      404,424,500  {object}  HTTPError
// @Router       /tx/{id}/data [get]
func (srv *Server) DataItemDataGet(ctx *gin.Context) {
	id := ctx.Param("id")
	contentType := ctx.Query("mime-type")
	accept := ctx.Request.Header.Get("accept")
	shouldDetect := false

	if contentType == "" && accept == "" {
		contentType = "application/octet-stream"
		shouldDetect = true
	} else if contentType == "" {
		contentType = accept
	}

	mtype := mimetype.Lookup(contentType)
	if mtype == nil {
		contentType = "application/octet-stream"
	}

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
	d, err := data_item.Decode(res)
	if err != nil {
		NewError(ctx, http.StatusInternalServerError, err)
		return
	}
	b, err := crypto.Base64URLDecode(d.Data)
	if err != nil {
		NewError(ctx, http.StatusInternalServerError, err)
		return
	}

	detect := mimetype.Detect(b)
	if shouldDetect && detect != nil {
		contentType = detect.String()
	}

	ctx.Data(http.StatusOK, contentType, b)
}
