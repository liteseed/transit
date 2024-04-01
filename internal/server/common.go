package server

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

type uploadRequestHeader struct {
	ContentType   *string `header:"content-type" binding:"required"`
	ContentLength *int    `header:"content-length" binding:"required"`
}

func verifyHeaders(c *gin.Context) (*uploadRequestHeader, error) {
	header := &uploadRequestHeader{}
	if err := c.ShouldBindHeader(header); err != nil {
		return nil, err
	}
	if *header.ContentLength == 0 || *header.ContentLength > MAX_DATA_SIZE {
		return nil, fmt.Errorf("content-length: supported range 1B - %dB", MAX_DATA_SIZE)
	}
	if *header.ContentType != CONTENT_TYPE_OCTET_STREAM {
		return nil, fmt.Errorf("content-type: unsupported")
	}
	return header, nil
}

func decodeBody(c *gin.Context, contentLength int) ([]byte, error) {
	rawData, err := c.GetRawData()
	if err != nil {
		return nil, err
	}
	if len(rawData) == 0 {
		return nil, errors.New("body: required")
	}
	if len(rawData) != contentLength {
		return nil, fmt.Errorf("content-length, body: length mismatch (%d, %d)", contentLength, len(rawData))
	}
	return rawData, nil
}
