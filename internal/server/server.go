package server

import (
	"context"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/liteseed/goar/wallet"
	"github.com/liteseed/sdk-go/contract"
	"github.com/liteseed/transit/internal/bundler"
	"github.com/liteseed/transit/internal/database"
)

const (
	CONTENT_TYPE_OCTET_STREAM = "application/octet-stream"
	MAX_DATA_SIZE             = uint(2 * 1024 * 1024 * 1024)
)

type Server struct {
	bundler  *bundler.Bundler
	contract *contract.Contract
	database *database.Database
	server   *http.Server
	wallet   *wallet.Wallet
	version  string
}

// @title          Liteseed API
// @version        0.0.1
// @description    The API is currently live at https://api.liteseed.xyz
// @contact.name   API Support
// @contact.url    https://liteseed.xyz/support
// @contact.email  support@liteseed.xyz
// @host           https://api.liteseed.xyz
func New(port string, version string, options ...func(*Server)) (*Server, error) {
	s := &Server{version: version}
	for _, o := range options {
		o(s)
	}
	engine := gin.New()
	engine.Use(cors.Default())
	engine.Use(gin.Recovery())

	engine.GET("/", s.Status)
	engine.GET("/price/:bytes", s.PriceGet)
	engine.GET("/tx/:id", s.DataItemGet)
	engine.GET("/tx/:id/data", s.DataItemDataGet)
	engine.GET("/tx/:id/status", s.DataItemStatusGet)
	engine.POST("/tx", s.DataItemPost)
	engine.PUT("/tx/:id/:payment_id", s.DataItemPut)
	engine.POST("/data", s.DataPost)

	s.server = &http.Server{
		Addr:    port,
		Handler: engine,
	}
	return s, nil
}

func WithBundler(b *bundler.Bundler) func(*Server) {
	return func(s *Server) {
		s.bundler = b
	}
}

func WithContracts(contract *contract.Contract) func(*Server) {
	return func(c *Server) {
		c.contract = contract
	}
}

func WithDatabase(db *database.Database) func(*Server) {
	return func(c *Server) {
		c.database = db
	}
}

func WithWallet(w *wallet.Wallet) func(*Server) {
	return func(c *Server) {
		c.wallet = w
	}
}
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown() error {
	return s.server.Shutdown(context.TODO())
}
