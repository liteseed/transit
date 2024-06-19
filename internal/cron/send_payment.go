package cron

import "github.com/liteseed/transit/internal/database/schema"

func (crn *Cron) sendPayment(o *schema.Order) *schema.Order {
	quantity, err := crn.wallet.Client.GetTransactionPrice(o.Size, "")
	if err != nil {
		crn.logger.Error("fail: gateway - get transaction price", "err", err)
		return nil
	}
	tx := crn.wallet.CreateTransaction(nil, o.Address, quantity, nil)
	_, err = crn.wallet.SignTransaction(tx)
	if err != nil {
		crn.logger.Error("fail: internal - sign transaction", "err", err)
		return nil
	}
	err = crn.wallet.SendTransaction(tx)
	if err != nil {
		crn.logger.Error("fail: gateway - send winston to address", "err", err)
		return nil
	}
	_, err = crn.bundler.DataItemPut(o.Address, o.ID, o.TransactionID)
	if err != nil {
		crn.logger.Error("fail: bundler - PUT /tx/"+o.ID+"/"+tx.ID, "err", err)
		return nil
	}
	return &schema.Order{Status: schema.Sent}
}

func (crn *Cron) SendPayments() {
	orders, err := crn.database.GetOrders(&schema.Order{Status: schema.Queued, Payment: schema.Paid})
	if err != nil {
		crn.logger.Error("fail: database - get orders", "error", err)
		return
	}

	for _, order := range *orders {
		u := crn.sendPayment(&order)
		if u != nil {
			err = crn.database.UpdateOrder(order.ID, u)
			if err != nil {
				crn.logger.Error("fail: database - update order", "err", err)
			}
		}
	}
}
