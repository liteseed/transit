package cron

import (
	"log/slog"

	"github.com/everFinance/goar"
	"github.com/liteseed/sdk-go/contract"
	"github.com/liteseed/transit/internal/database"
	"github.com/robfig/cron/v3"
)

type Cron struct {
	c        *cron.Cron
	contract *contract.Contract
	database *database.Database
	logger   *slog.Logger
	wallet   *goar.Wallet
}

type Option = func(*Cron)

func New(options ...func(*Cron)) (*Cron, error) {
	c := &Cron{c: cron.New()}
	for _, o := range options {
		o(c)
	}
	return c, nil
}

func WthContracts(contract *contract.Contract) Option {
	return func(c *Cron) {
		c.contract = contract
	}
}

func WithDatabase(db *database.Database) Option {
	return func(c *Cron) {
		c.database = db
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(c *Cron) {
		c.logger = logger
	}
}

func WithWallet(s *goar.Wallet) Option {
	return func(c *Cron) {
		c.wallet = s
	}
}

func (c *Cron) Start() {
	c.c.Start()
}

func (c *Cron) Shutdown() {
	c.c.Stop()
}

func (c *Cron) Setup(spec string) error {
	_, err := c.c.AddFunc(spec, c.CheckTransactionAmount)
	if err != nil {
		return err
	}
	_, err = c.c.AddFunc(spec, c.CheckTransactionConfirmation)
	if err != nil {
		return err
	}
	_, err = c.c.AddFunc(spec, c.SendTransaction)
	if err != nil {
		return err
	}
	return nil
}
