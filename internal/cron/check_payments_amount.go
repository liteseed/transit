package cron

import (
	"strconv"

	"github.com/liteseed/transit/internal/database/schema"
	"github.com/liteseed/transit/internal/utils"
)

func (crn *Cron) checkSinglePaymentAmount(o *schema.Order) *schema.Order {
	tx, err := crn.wallet.Client.GetTransactionByID(o.TransactionId)
	if err != nil {
		crn.logger.Error("fail: gateway - get transaction by id", "err", err)
		return nil
	}

	payment, err := strconv.ParseUint(tx.Quantity, 10, 64)
	if err != nil {
		crn.logger.Error("fail: internal - conversion to uint", "err", err)
		return nil
	}

	r, err := crn.wallet.Client.GetTransactionPrice(o.Size, "")
	if err != nil {
		crn.logger.Error("fail: gateway - get transaction status", "err", err)
		return nil
	}
	price, err := strconv.ParseUint(utils.CalculatePriceWithFee(r), 10, 64)
	if err != nil {
		crn.logger.Error("fail: internal - conversion to uint", "err", err)
		return nil
	}

	u := &schema.Order{}
	if payment >= price && tx.Target == crn.wallet.Signer.Address {
		u.Payment = schema.Paid
	} else {
		u.Payment = schema.Invalid
		u.Status = schema.Failed
	}
	return u
}

func (crn *Cron) CheckPaymentsAmount() {
	orders, err := crn.database.GetOrders(&schema.Order{Payment: schema.Queued})
	if err != nil {
		crn.logger.Error("fail: database - get orders", "error", err)
		return
	}
	for _, order := range *orders {
		u := crn.checkSinglePaymentAmount(&order)
		err = crn.database.UpdateOrder(order.Id, u)
		if err != nil {
			crn.logger.Error("fail: database - update order", "err", err)
			return
		}
	}
}
