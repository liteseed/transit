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
}

func New(port string, version string, options ...func(*Server)) (*Server, error) {
	s := &Server{}
	for _, o := range options {
		o(s)
	}
	engine := gin.New()

	config := cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"POST", "GET", "OPTIONS", "HEAD"},
		AllowHeaders:  []string{"x-transaction-id", "Content-Type", "Content-Length"},
		ExposeHeaders: []string{"x-transaction-id"},
	}
	engine.Use(cors.New(config))

	engine.Use(gin.Recovery())
	// engine.Use(ErrorHandler)
	engine.GET("", s.Status)
	engine.GET("price/:bytes", s.PriceGet)
	engine.GET("tx/:id", s.DataItemGet)
	engine.GET("tx/:id/status", s.DataItemStatusGet)
	engine.POST("tx", s.DataItemPost)
	engine.PUT("tx/:id/:transaction_id", s.DataItemPut)

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
