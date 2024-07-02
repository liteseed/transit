package cron

import "github.com/liteseed/transit/internal/database/schema"

// Number of Confirmation > 10
func (crn *Cron) CheckPaymentsConfirmations() {
	orders, err := crn.database.GetOrders(&schema.Order{Payment: schema.Unpaid})
	if err != nil {
		crn.logger.Error("fail: database - get orders", "error", err)
		return
	}
	for _, order := range *orders {
		status, err := crn.wallet.Client.GetTransactionStatus(order.TransactionId)
		if err != nil {
			crn.logger.Error("fail: gateway - get transaction status", "err", err)
			continue
		}
		if status.NumberOfConfirmations >= 10 {
			err = crn.database.UpdateOrder(order.Id, &schema.Order{Payment: schema.Paid})
			if err != nil {
				crn.logger.Error("fail: database - update order", "err", err)
			}
		}
		continue
	}
}
