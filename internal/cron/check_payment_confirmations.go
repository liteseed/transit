package cron

import "github.com/liteseed/transit/internal/database/schema"

func (c *Cron) checkPaymentConfirmations(transactionID string) *schema.Order {
	status, err := c.wallet.Client.GetTransactionStatus(transactionID)
	if err != nil {
		c.logger.Error("fail: gateway - get transaction status", "err", err)
		return nil
	}
	if status.NumberOfConfirmations >= 10 {
		return &schema.Order{Payment: schema.Confirmed}
	}
	return nil
}

// Number of Confirmation > 10
func (c *Cron) CheckPaymentConfirmations() {
	orders, err := c.database.GetOrders(&schema.Order{Status: schema.Queued, Payment: schema.Unpaid})
	if err != nil {
		c.logger.Error("fail: database - get orders", "error", err)
		return
	}
	for _, order := range *orders {
		u := c.checkPaymentConfirmations(order.TransactionID)
		if u != nil {
			err = c.database.UpdateOrder(order.ID, u)
			if err != nil {
				c.logger.Error("fail: database - update order", "err", err)
			}
		}
		continue
	}
}
