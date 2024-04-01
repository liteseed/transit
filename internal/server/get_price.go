package server

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Context) getPrice(c *gin.Context) {
	b, valid := c.Params.Get("bytes")
	if !valid {
		c.AbortWithError(http.StatusBadRequest, errors.New("bytes size is required"))
		return
	}

	res, err := http.Get("https://arweave.net/price/" + b)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	r, err := io.ReadAll(res.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, string(r))
}
