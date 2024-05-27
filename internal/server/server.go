package server

import (
	"context"
	"net/http"

	"github.com/everFinance/goar"
	"github.com/gin-gonic/gin"
	"github.com/liteseed/sdk-go/contract"
	"github.com/liteseed/transit/internal/database"
	"github.com/liteseed/transit/internal/store"
)

const (
	CONTENT_TYPE_OCTET_STREAM = "application/octet-stream"
	MAX_DATA_SIZE             = uint(2 * 1024 * 1024 * 1024)
)

type Server struct {
	contract *contract.Contract
	database *database.Database
	gateway  string
	server   *http.Server
	store    *store.Store
	wallet   *goar.Wallet
}

func New(port string, version string, gateway string, options ...func(*Server)) (*Server, error) {
	s := &Server{}
	for _, o := range options {
		o(s)
	}
	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.Use(ErrorHandler)
	engine.GET("/", s.Status)
	engine.GET("/price/:bytes", s.PriceGet)
	engine.POST("/upload", s.DataPost)
	engine.GET("/tx/:id", s.DataItemGet)
	engine.POST("/tx", s.DataItemPost)

	s.server = &http.Server{
		Addr:    port,
		Handler: engine,
	}
	return s, nil
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

func WithStore(s *store.Store) func(*Server) {
	return func(c *Server) {
		c.store = s
	}
}
func WithWallet(w *goar.Wallet) func(*Server) {
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
