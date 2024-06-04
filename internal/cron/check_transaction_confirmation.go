package cron

import "github.com/liteseed/transit/internal/database/schema"

// Number of Confirmation > 10
func (c *Cron) CheckTransactionConfirmation() {
	orders, err := c.database.GetOrders(&schema.Order{Status: schema.Queued})
	if err != nil {
		c.logger.Error("fail: database - get orders", "error", err)
		return
	}
	for _, order := range *orders {
		status, err := c.wallet.Client.GetTransactionStatus(order.TransactionID)
		if err != nil {
			c.logger.Error("fail: gateway - get transaction status", "err", err)
		}
		if status.NumberOfConfirmations >= 10 {
			err = c.database.UpdateOrder(&schema.Order{ID: order.ID, Status: schema.Confirmed})
			if err != nil {
				c.logger.Error("fail: database - update order", "err", err)
			}
		}
	}

}
