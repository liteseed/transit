package server

import (
	"encoding/json"
	"net/http"

	"github.com/everFinance/goar/types"
	"github.com/gin-gonic/gin"
)

const PROCESS = "lJLnoDsq8z0NJrTbQqFQ1arJayfuqWPqwRaW_3aNCgk"

type Staker struct {
	Id  string `json:"id"`
	Url string `json:"url"`
}

func (s *Context) GetBundlers(c *gin.Context) {
	mId, err := s.ao.SendMessage(PROCESS, "staker", []types.Tag{{Name: "Action", Value: "Stakers"}}, "", s.signer)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	r, err := s.ao.ReadResult(PROCESS, mId)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	var stakers []Staker
	err = json.Unmarshal([]byte(r.Messages[0]["Data"].(string)), &stakers)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, stakers)
}
