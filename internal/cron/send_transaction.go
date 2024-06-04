package cron

import (
	"math/big"
	"net/http"

	"github.com/liteseed/transit/internal/database/schema"
)

func putData(u string, id string, transactionId string) error {
	req, err := http.NewRequest("PUT", "http://"+u+"/tx/"+id+"/"+transactionId, nil)
	if err != nil {
		return err
	}

	c := http.DefaultClient
	_, err = c.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cron) SendTransaction() {
	orders, err := c.database.GetOrders(&schema.Order{Status: schema.Paid})
	if err != nil {
		c.logger.Error("fail: database - get orders", "error", err)
		return
	}

	for _, order := range *orders {
		price, err := c.wallet.Client.GetTransactionPrice(int(order.Size), &order.Address)
		if err != nil {
			c.logger.Error("fail: gateway - get transaction status", "err", err)
		}

		tx, err := c.wallet.SendWinston(big.NewInt(price), order.Address, nil)
		if err != nil {
			c.logger.Error("fail: gateway - send winston to address", "err", err)
		}
		err = putData(order.URL, order.ID, tx.ID)
		if err != nil {
			c.logger.Error("fail: bundler - PUT /tx/"+order.ID+"/"+tx.ID, "err", err)
		}
	}
}
