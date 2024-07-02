package server

import (
	"github.com/gin-gonic/gin"
)

// NewError example
func NewError(ctx *gin.Context, status int, err error) {
	ctx.JSON(status, HTTPError{
		Code:    status,
		Message: err.Error(),
	})
}

// HTTPError example
type HTTPError struct {
	Code    int    `json:"code" format:"integer"`
	Message string `json:"message" format:"string"`
}
