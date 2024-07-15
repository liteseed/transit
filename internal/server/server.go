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

const ContentTypeOctetStream = "application/octet-stream"

type Server struct {
	bundler  *bundler.Bundler
	contract *contract.Contract
	database *database.Database
	server   *http.Server
	wallet   *wallet.Wallet
	version  string
}

// @title          Liteseed API
// @version        0.0.6
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
	engine.GET("/tx/:id", s.GetDataItem)
	engine.GET("/tx/:id/:field", s.GetDataItemField)
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
	return func(srv *Server) {
		srv.bundler = b
	}
}

func WithContracts(contract *contract.Contract) func(*Server) {
	return func(srv *Server) {
		srv.contract = contract
	}
}

func WithDatabase(db *database.Database) func(*Server) {
	return func(srv *Server) {
		srv.database = db
	}
}

func WithWallet(w *wallet.Wallet) func(*Server) {
	return func(srv *Server) {
		srv.wallet = w
	}
}
func (srv *Server) Start() error {
	return srv.server.ListenAndServe()
}

func (srv *Server) Shutdown() error {
	return srv.server.Shutdown(context.TODO())
}
