package server

import (
	"log"

	"github.com/everFinance/goar"
	"github.com/gin-gonic/gin"

	"github.com/liteseed/aogo"
)

const (
	CONTENT_TYPE_OCTET_STREAM = "application/octet-stream"
	MAX_DATA_SIZE             = 1_073_824
	MAX_DATA_ITEM_SIZE        = 1_073_824
)

type Context struct {
	ao     *aogo.AO
	engine *gin.Engine
	signer *goar.ItemSigner
}

func New(ao *aogo.AO, signer *goar.ItemSigner) *Context {
	engine := gin.New()

	engine.Use(gin.Recovery())
	s := &Context{ao: ao, engine: engine, signer: signer}

	s.engine.Use(ErrorHandler)
	s.engine.GET("/", s.Status)
	s.engine.GET("/price/:bytes", s.getPrice)
	s.engine.GET("/bundlers", s.GetBundlers)
	s.engine.POST("/data", s.uploadData)
	s.engine.GET("/data/:id", s.getData)

	return s
}

func (s *Context) Run(port string) {
	err := s.engine.Run(port)
	if err != nil {
		log.Fatalln("failed to start server", err)
	}
}
