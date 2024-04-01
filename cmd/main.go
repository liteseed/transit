package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/everFinance/goar"
	"github.com/liteseed/aogo"
	"github.com/liteseed/transit/internal/server"
)

type StartConfig struct {
	Port   string
	Signer string
}

func main() {
	configData, err := os.ReadFile("./config.json")
	if err != nil {
		log.Fatalln(err)
	}

	var config StartConfig

	err = json.Unmarshal(configData, &config)
	if err != nil {
		log.Fatalln(err)
	}

	ao, err := aogo.New()
	if err != nil {
		log.Fatal(err)
	}

	signer, err := goar.NewSignerFromPath(config.Signer)
	if err != nil {
		log.Fatal(err)
	}

	itemSigner, err := goar.NewItemSigner(signer)
	if err != nil {
		log.Fatal(err)
	}

	s := server.New(ao, itemSigner)
	s.Run(":8000")

	if err != nil {
		log.Fatal(err)
	}
}
