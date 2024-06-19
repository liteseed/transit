package cron

import (
	"log/slog"

	"github.com/liteseed/goar/wallet"
	"github.com/liteseed/sdk-go/contract"
	"github.com/liteseed/transit/internal/bundler"
	"github.com/liteseed/transit/internal/database"
	"github.com/robfig/cron/v3"
)

type Cron struct {
	bundler  *bundler.Bundler
	c        *cron.Cron
	contract *contract.Contract
	database *database.Database
	logger   *slog.Logger
	wallet   *wallet.Wallet
}

type Option = func(*Cron)

func New(options ...func(*Cron)) (*Cron, error) {
	c := &Cron{c: cron.New()}
	for _, o := range options {
		o(c)
	}
	return c, nil
}

func WithBundler(b *bundler.Bundler) Option {
	return func(c *Cron) {
		c.bundler = b
	}
}

func WithContracts(contract *contract.Contract) Option {
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

func WithWallet(s *wallet.Wallet) Option {
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
	_, err := c.c.AddFunc(spec, c.CheckPaymentsAmount)
	if err != nil {
		return err
	}
	_, err = c.c.AddFunc(spec, c.CheckPaymentsConfirmations)
	if err != nil {
		return err
	}
	_, err = c.c.AddFunc(spec, c.SendPayments)
	if err != nil {
		return err
	}
	return nil
}
