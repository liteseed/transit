package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/liteseed/goar/crypto"
	"github.com/liteseed/goar/transaction/data_item"
)

// GetDataItem
//
// Get data-item field godoc
// @Summary      Get a field of a data-item
// @Description  If you choose to skip field  the whole data-item is returned.
// @Description  Get only the specified field of a posted data-item.
// @Description  In case the specified field is data it tries to automatically detect mime-type.
// @Description  You can specify the response mime-type by either sending a mime-type query parameter or an accept header in the request.
// @Description  Supported mime-type are listed here - https://github.com/gabriel-vasile/mimetype/blob/master/supported_mimes.md.
// @Description  If all else fails defaults to `application/octet-stream`
// @Tags         Fetch
// @Accept       json
// @Param        id           path      string    true      "id of the data-item"
// @Param        mime-type    query     string    false     "mime type of the response"
// @Success      200 		  {bytes}   data
// @Failure      404,424,500  {object}  HTTPError
// @Router       /tx/{id}/{field} [get]
func (srv *Server) GetDataItem(ctx *gin.Context) {
	id := ctx.Param("id")
	field := ctx.Param("field")

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
	switch field {
	case "":
		ctx.Data(http.StatusOK, ContentTypeOctetStream, d.Raw)
		return
	case "anchor":
		ctx.JSON(http.StatusOK, d.Anchor)
		return
	case "owner":
		ctx.JSON(http.StatusOK, d.Owner)
		return
	case "signature":
		ctx.JSON(http.StatusOK, d.Signature)
		return
	case "tags":
		tags, err := json.Marshal(d.Tags)
		if err != nil {
			NewError(ctx, http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, tags)
		return
	case "target":
		ctx.JSON(http.StatusOK, d.Target)
		return
	case "data":
		contentType := ctx.Query("mime-type")
		accept := ctx.Request.Header.Get("accept")
		b, err := decodeData(d.Data, contentType, accept)
		if err != nil {
			NewError(ctx, http.StatusInternalServerError, err)
		}
		ctx.Data(http.StatusOK, contentType, b)
		return
	default:
		NewError(ctx, http.StatusBadRequest, errors.New("field not found"))
		return
	}

}

func decodeData(data string, contentType string, accept string) ([]byte, error) {
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
	b, err := crypto.Base64URLDecode(data)
	if err != nil {
		return nil, err
	}

	detect := mimetype.Detect(b)
	if shouldDetect && detect != nil {
		contentType = detect.String()
	}
	return b, nil
}
