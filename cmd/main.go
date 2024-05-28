package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/everFinance/goar"
	"github.com/liteseed/sdk-go/contract"
	"github.com/liteseed/transit/internal/database"
	"github.com/liteseed/transit/internal/server"
	"github.com/liteseed/transit/internal/store"
)

var Version string

type StartConfig struct {
	Database string
	Gateway  string
	Port     string
	Process  string
	Signer   string
	Store    string
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
	db, err := database.New(config.Database)
	if err != nil {
		log.Fatalln(err)
	}

	wallet, err := goar.NewWalletFromPath(config.Signer, "http://localhost:8008")
	if err != nil {
		log.Fatalln(err)
	}

	process := config.Process

	contracts := contract.New(process, wallet.Signer)

	store := store.New(config.Store)

	s, err := server.New(":8000", Version, "http://localhost:8008", server.WithContracts(contracts), server.WithDatabase(db), server.WithWallet(wallet), server.WithStore(store))
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		err := s.Start()
		if err != http.ErrServerClosed {
			log.Fatal("failed to start server", err)
		}
	}()

	<-quit

	log.Println("Shutdown")

	if err = store.Shutdown(); err != nil {
		log.Fatal("failed to shutdown", err)
	}

	time.Sleep(2 * time.Second)
	if err = s.Shutdown(); err != nil {
		log.Fatal("failed to shutdown", err)
	}
}
