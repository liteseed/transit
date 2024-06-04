package cron

import (
	"strconv"

	"github.com/liteseed/transit/internal/database/schema"
)

// Check Transaction ID
// Price of Upload
// Number of Confirmation > 10
func (c *Cron) CheckTransactionAmount() {
	orders, err := c.database.GetOrders(&schema.Order{Status: schema.Confirmed})
	if err != nil {
		c.logger.Error("fail: database - get orders", "error", err)
		return
	}

	for _, order := range *orders {
		o := schema.Order{ID: order.ID}

		transaction, err := c.wallet.Client.GetTransactionByID(order.TransactionID)
		if err != nil {
			return
		}

		payment, err := strconv.ParseUint(transaction.Quantity, 10, 32)
		if err != nil {
			return
		}
		price, err := c.wallet.Client.GetTransactionPrice(int(order.Size), nil)
		if err != nil {
			c.logger.Error("fail: gateway - get transaction status", "err", err)
		}
		if payment < uint64(price) && transaction.Target == c.wallet.Signer.Address {
			o.Status = schema.Paid
		} else {
			o.Status = schema.Invalid
		}
		err = c.database.UpdateOrder(&order)
		if err != nil {
			return
		}
	}

}
