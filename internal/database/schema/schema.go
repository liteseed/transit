package schema

import "database/sql/driver"

type Payment string
type Status string

const (
	// Order
	Created = "created" // Order Created
	Queued  = "queued"  // Order Transaction ID added
	Sent    = "Sent"    // Order Sent

	Failed = "failed" // Order Failed

	// Payment
	Unpaid    = "unpaid"
	Confirmed = "confirmed" // Order Transaction has > 10 confirmation
	Paid      = "paid"      // Ready to Send
	Invalid   = "invalid"   // Not enough AR
)

func (s *Status) Scan(value any) error {
	*s = Status(value.(string))
	return nil
}

func (s Status) Value() (driver.Value, error) {
	return string(s), nil
}

func (s *Payment) Scan(value any) error {
	*s = Payment(value.(string))
	return nil
}

func (s Payment) Value() (driver.Value, error) {
	return string(s), nil
}

type Order struct {
	ID            string  `json:"id"`
	TransactionID string  `json:"transaction_id"`
	URL           string  `json:"url"`
	Address       string  `json:"address"`
	Status        Status  `gorm:"index:idx_status;default:created" sql:"type:status" json:"status"`
	Payment       Payment `gorm:"index:idx_payment;default:unpaid" sql:"type:status" json:"payment"`
	Size          int    `json:"size"`
}
