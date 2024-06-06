package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/everFinance/goar"
	"github.com/liteseed/sdk-go/contract"
	"github.com/liteseed/transit/internal/cron"
	"github.com/liteseed/transit/internal/database"
	"github.com/liteseed/transit/internal/server"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Version string

type StartConfig struct {
	Database string
	Gateway  string
	Log      string
	Port     string
	Process  string
	Signer   string
}

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	configData, err := os.ReadFile("./config.json")
	if err != nil {
		log.Fatalln(err)
	}

	var config StartConfig

	err = json.Unmarshal(configData, &config)
	if err != nil {
		log.Fatalln(err)
	}

	logger := slog.New(
		slog.NewJSONHandler(
			&lumberjack.Logger{
				Filename:   config.Log,
				MaxSize:    2,
				MaxBackups: 3,
				MaxAge:     28,
				Compress:   true,
			},
			&slog.HandlerOptions{AddSource: true},
		),
	)

	db, err := database.New(config.Database)
	if err != nil {
		log.Fatalln(err)
	}

	wallet, err := goar.NewWalletFromPath(config.Signer, config.Gateway)
	if err != nil {
		log.Fatalln(err)
	}

	process := config.Process

	contracts := contract.New(process, wallet.Signer)

	c, err := cron.New(config.Gateway, cron.WthContracts(contracts), cron.WithDatabase(db), cron.WithWallet(wallet), cron.WithLogger(logger))
	if err != nil {
		log.Fatal(err)
	}
	err = c.Setup("* * * * *")
	if err != nil {
		log.Fatal(err)
	}

	s, err := server.New(":8000", Version, config.Gateway, server.WithContracts(contracts), server.WithDatabase(db), server.WithWallet(wallet))
	if err != nil {
		log.Fatal(err)
	}

	go c.Start()
	go func() {
		err := s.Start()
		if err != http.ErrServerClosed {
			log.Fatal("failed to start server", err)
		}
	}()

	<-quit

	log.Println("Shutdown")
	c.Shutdown()
	if err = s.Shutdown(); err != nil {
		log.Fatal("failed to shutdown", err)
	}
	time.Sleep(2 * time.Second)
}
