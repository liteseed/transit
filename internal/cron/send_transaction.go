package cron

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/everFinance/goar/types"
	"github.com/liteseed/transit/internal/database/schema"
)

func SendTransaction(u string, tx *types.Transaction) error {
	b, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", u+"/tx", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/json")

	c := http.DefaultClient
	_, err = c.Do(req)
	if err != nil {
		return err
	}
	return nil
}

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
	orders, err := c.database.GetOrders(&schema.Order{Status: schema.Queued, Payment: schema.Paid})
	if err != nil {
		c.logger.Error("fail: database - get orders", "error", err)
		return
	}

	for _, order := range *orders {

		transferPrice, err := c.PriceOfUpload("0", order.Address)
		if err != nil {
			c.logger.Error("fail: gateway - get transaction price", "err", err)
			continue
		}
		price, err := c.PriceOfUpload(strconv.FormatUint(uint64(order.Size), 10), "")
		if err != nil {
			c.logger.Error("fail: gateway - get transaction price", "err", err)
			continue
		}

		tx := &types.Transaction{
			Format:   2,
			Owner:    c.wallet.Signer.Owner(),
			Quantity: strconv.FormatUint(price, 10),
			Target:   order.Address,
			Data:     "",
			DataSize: "0",
			DataRoot: "",
			Tags:     []types.Tag{},
			Reward:   strconv.FormatUint(transferPrice, 10),
		}
		anchor, err := c.wallet.Client.GetTransactionAnchor()
		if err != nil {
			c.logger.Error("fail: gateway - get transaction anchor", "err", err)
			continue
		}
		tx.LastTx = anchor
		tx.Owner = c.wallet.Owner()
		err = c.wallet.Signer.SignTx(tx)
		if err != nil {
			c.logger.Error("fail: internal - sign transaction", "err", err)
			continue
		}
		err = SendTransaction(c.gateway, tx)
		if err != nil {
			c.logger.Error("fail: gateway - send winston to address", "err", err)
			continue
		}

		err = putData(order.URL, order.ID, tx.ID)
		if err != nil {
			c.logger.Error("fail: bundler - PUT /tx/"+order.ID+"/"+tx.ID, "err", err)
			continue
		}
		err = c.database.UpdateOrder(&schema.Order{ID: order.ID, Status: schema.Sent})
		if err != nil {
			c.logger.Error("fail: database - update order", "err", err)
		}
		continue
	}
}
